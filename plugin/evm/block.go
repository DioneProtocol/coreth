// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package evm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/DioneProtocol/coreth/core/types"

	"github.com/DioneProtocol/odysseygo/ids"
	"github.com/DioneProtocol/odysseygo/snow/choices"
)

var (
	bonusBlockMainnetHeights = make(map[uint64]ids.ID)
	// first height that processed a TX included on a
	// bonus block is the canonical height for that TX.
	canonicalBlockMainnetHeights = []uint64{}

	errMissingUTXOs = errors.New("missing UTXOs")
)

func init() {
	mainnetBonusBlocks := map[uint64]string{}

	for height, blkIDStr := range mainnetBonusBlocks {
		blkID, err := ids.FromString(blkIDStr)
		if err != nil {
			panic(err)
		}
		bonusBlockMainnetHeights[height] = blkID
	}
}

// Block implements the snowman.Block interface
type Block struct {
	id        ids.ID
	ethBlock  *types.Block
	vm        *VM
	status    choices.Status
	atomicTxs []*Tx
}

// newBlock returns a new Block wrapping the ethBlock type and implementing the snowman.Block interface
func (vm *VM) newBlock(ethBlock *types.Block) (*Block, error) {
	isOdyPhase5 := vm.chainConfig.IsOdyPhase5(ethBlock.Time())
	atomicTxs, err := ExtractAtomicTxs(ethBlock.ExtData(), isOdyPhase5, vm.codec)
	if err != nil {
		return nil, err
	}

	return &Block{
		id:        ids.ID(ethBlock.Hash()),
		ethBlock:  ethBlock,
		vm:        vm,
		atomicTxs: atomicTxs,
	}, nil
}

// ID implements the snowman.Block interface
func (b *Block) ID() ids.ID { return b.id }

// Accept implements the snowman.Block interface
func (b *Block) Accept(context.Context) error {
	vm := b.vm

	// Although returning an error from Accept is considered fatal, it is good
	// practice to cleanup the batch we were modifying in the case of an error.
	defer vm.db.Abort()

	b.status = choices.Accepted
	log.Debug(fmt.Sprintf("Accepting block %s (%s) at height %d", b.ID().Hex(), b.ID(), b.Height()))
	if err := vm.blockChain.Accept(b.ethBlock); err != nil {
		return fmt.Errorf("chain could not accept %s: %w", b.ID(), err)
	}
	if err := vm.acceptedBlockDB.Put(lastAcceptedKey, b.id[:]); err != nil {
		return fmt.Errorf("failed to put %s as the last accepted block: %w", b.ID(), err)
	}

	for _, tx := range b.atomicTxs {
		// Remove the accepted transaction from the mempool
		vm.mempool.RemoveTx(tx)
	}

	// Update VM state for atomic txs in this block. This includes updating the
	// atomic tx repo, atomic trie, and shared memory.
	atomicState, err := b.vm.atomicBackend.GetVerifiedAtomicState(common.Hash(b.ID()))
	if err != nil {
		// should never occur since [b] must be verified before calling Accept
		return err
	}
	commitBatch, err := b.vm.db.CommitBatch()
	if err != nil {
		return fmt.Errorf("could not create commit batch processing block[%s]: %w", b.ID(), err)
	}
	return atomicState.Accept(commitBatch)
}

// Reject implements the snowman.Block interface
// If [b] contains an atomic transaction, attempt to re-issue it
func (b *Block) Reject(context.Context) error {
	b.status = choices.Rejected
	log.Debug(fmt.Sprintf("Rejecting block %s (%s) at height %d", b.ID().Hex(), b.ID(), b.Height()))
	for _, tx := range b.atomicTxs {
		b.vm.mempool.RemoveTx(tx)
		if err := b.vm.issueTx(tx, false /* set local to false when re-issuing */); err != nil {
			log.Debug("Failed to re-issue transaction in rejected block", "txID", tx.ID(), "err", err)
		}
	}
	atomicState, err := b.vm.atomicBackend.GetVerifiedAtomicState(common.Hash(b.ID()))
	if err != nil {
		// should never occur since [b] must be verified before calling Reject
		return err
	}
	if err := atomicState.Reject(); err != nil {
		return err
	}
	return b.vm.blockChain.Reject(b.ethBlock)
}

// SetStatus implements the InternalBlock interface allowing ChainState
// to set the status on an existing block
func (b *Block) SetStatus(status choices.Status) { b.status = status }

// Status implements the snowman.Block interface
func (b *Block) Status() choices.Status {
	return b.status
}

// Parent implements the snowman.Block interface
func (b *Block) Parent() ids.ID {
	return ids.ID(b.ethBlock.ParentHash())
}

// Height implements the snowman.Block interface
func (b *Block) Height() uint64 {
	return b.ethBlock.NumberU64()
}

// Timestamp implements the snowman.Block interface
func (b *Block) Timestamp() time.Time {
	return time.Unix(int64(b.ethBlock.Time()), 0)
}

// syntacticVerify verifies that a *Block is well-formed.
func (b *Block) syntacticVerify() error {
	if b == nil || b.ethBlock == nil {
		return errInvalidBlock
	}

	header := b.ethBlock.Header()
	rules := b.vm.chainConfig.OdysseyRules(header.Number, header.Time)
	return b.vm.syntacticBlockValidator.SyntacticVerify(b, rules)
}

// Verify implements the snowman.Block interface
func (b *Block) Verify(context.Context) error {
	return b.verify(true)
}

func (b *Block) verify(writes bool) error {
	if err := b.syntacticVerify(); err != nil {
		return fmt.Errorf("syntactic block verification failed: %w", err)
	}

	// verify UTXOs named in import txs are present in shared memory.
	if err := b.verifyUTXOsPresent(); err != nil {
		return err
	}

	err := b.vm.blockChain.InsertBlockManual(b.ethBlock, writes)
	if err != nil || !writes {
		// if an error occurred inserting the block into the chain
		// or if we are not pinning to memory, unpin the atomic trie
		// changes from memory (if they were pinned).
		if atomicState, err := b.vm.atomicBackend.GetVerifiedAtomicState(b.ethBlock.Hash()); err == nil {
			_ = atomicState.Reject() // ignore this error so we can return the original error instead.
		}
	}
	return err
}

// verifyUTXOsPresent returns an error if any of the atomic transactions name UTXOs that
// are not present in shared memory.
func (b *Block) verifyUTXOsPresent() error {
	blockHash := common.Hash(b.ID())
	if b.vm.atomicBackend.IsBonus(b.Height(), blockHash) {
		log.Info("skipping atomic tx verification on bonus block", "block", blockHash)
		return nil
	}

	if !b.vm.bootstrapped {
		return nil
	}

	// verify UTXOs named in import txs are present in shared memory.
	for _, atomicTx := range b.atomicTxs {
		utx := atomicTx.UnsignedAtomicTx
		chainID, requests, err := utx.AtomicOps()
		if err != nil {
			return err
		}
		if _, err := b.vm.ctx.SharedMemory.Get(chainID, requests.RemoveRequests); err != nil {
			return fmt.Errorf("%w: %s", errMissingUTXOs, err)
		}
	}
	return nil
}

// Bytes implements the snowman.Block interface
func (b *Block) Bytes() []byte {
	res, err := rlp.EncodeToBytes(b.ethBlock)
	if err != nil {
		panic(err)
	}
	return res
}

func (b *Block) String() string { return fmt.Sprintf("EVM block, ID = %s", b.ID()) }
