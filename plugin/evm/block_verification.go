// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package evm

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	safemath "github.com/DioneProtocol/odysseygo/utils/math"

	"github.com/DioneProtocol/coreth/constants"
	"github.com/DioneProtocol/coreth/core/types"
	"github.com/DioneProtocol/coreth/params"
	"github.com/DioneProtocol/coreth/trie"
)

var (
	odysseyPhase0MinGasPrice = big.NewInt(params.LaunchMinGasPrice)
	odysseyPhase1MinGasPrice = big.NewInt(params.OdysseyPhase1MinGasPrice)
)

type BlockValidator interface {
	SyntacticVerify(b *Block, rules params.Rules) error
}

type blockValidator struct {
	extDataHashes map[common.Hash]common.Hash
}

func NewBlockValidator(extDataHashes map[common.Hash]common.Hash) BlockValidator {
	return &blockValidator{
		extDataHashes: extDataHashes,
	}
}

func (v blockValidator) SyntacticVerify(b *Block, rules params.Rules) error {
	if b == nil || b.ethBlock == nil {
		return errInvalidBlock
	}

	ethHeader := b.ethBlock.Header()
	blockHash := b.ethBlock.Hash()

	if !rules.IsOdysseyPhase1 {
		if v.extDataHashes != nil {
			extData := b.ethBlock.ExtData()
			extDataHash := types.CalcExtDataHash(extData)
			// If there is no extra data, check that there is no extra data in the hash map either to ensure we do not
			// have a block that is unexpectedly missing extra data.
			expectedExtDataHash, ok := v.extDataHashes[blockHash]
			if len(extData) == 0 {
				if ok {
					return fmt.Errorf("found block with unexpected missing extra data (%s, %d), expected extra data hash: %s", blockHash, b.Height(), expectedExtDataHash)
				}
			} else {
				// If there is extra data, check to make sure that the extra data hash matches the expected extra data hash for this
				// block
				if extDataHash != expectedExtDataHash {
					return fmt.Errorf("extra data hash in block (%s, %d): %s, did not match the expected extra data hash: %s", blockHash, b.Height(), extDataHash, expectedExtDataHash)
				}
			}
		}
	}

	// Skip verification of the genesis block since it should already be marked as accepted.
	if blockHash == b.vm.genesisHash {
		return nil
	}

	// Verify the ExtDataHash field
	if rules.IsOdysseyPhase1 {
		if hash := types.CalcExtDataHash(b.ethBlock.ExtData()); ethHeader.ExtDataHash != hash {
			return fmt.Errorf("extra data hash mismatch: have %x, want %x", ethHeader.ExtDataHash, hash)
		}
	} else {
		if ethHeader.ExtDataHash != (common.Hash{}) {
			return fmt.Errorf(
				"expected ExtDataHash to be empty but got %x",
				ethHeader.ExtDataHash,
			)
		}
	}

	if !ethHeader.Number.IsUint64() {
		return fmt.Errorf("invalid block number: %v", ethHeader.Number)
	}
	if !ethHeader.Difficulty.IsUint64() || ethHeader.Difficulty.Cmp(common.Big1) != 0 {
		return fmt.Errorf("invalid difficulty: %d", ethHeader.Difficulty)
	}
	if ethHeader.Nonce.Uint64() != 0 {
		return fmt.Errorf("invalid block nonce: %v", ethHeader.Nonce)
	}
	if ethHeader.MixDigest != (common.Hash{}) {
		return fmt.Errorf("invalid mix digest: %v", ethHeader.MixDigest)
	}

	// Enforce static gas limit after OdysseyPhase1 (prior to OdysseyPhase1 it's handled in processing).
	if rules.IsCortina {
		if ethHeader.GasLimit != params.CortinaGasLimit {
			return fmt.Errorf(
				"expected gas limit to be %d after cortina but got %d",
				params.CortinaGasLimit, ethHeader.GasLimit,
			)
		}
	} else if rules.IsOdysseyPhase1 {
		if ethHeader.GasLimit != params.OdysseyPhase1GasLimit {
			return fmt.Errorf(
				"expected gas limit to be %d after odyssey phase 1 but got %d",
				params.OdysseyPhase1GasLimit, ethHeader.GasLimit,
			)
		}
	}

	// Check that the size of the header's Extra data field is correct for [rules].
	headerExtraDataSize := uint64(len(ethHeader.Extra))
	switch {
	case rules.IsOdysseyPhase1:
		if headerExtraDataSize != params.OdysseyPhase1ExtraDataSize {
			return fmt.Errorf(
				"expected header ExtraData to be %d but got %d",
				params.OdysseyPhase1ExtraDataSize, headerExtraDataSize,
			)
		}
	default:
		if headerExtraDataSize > params.MaximumExtraDataSize {
			return fmt.Errorf(
				"expected header ExtraData to be <= %d but got %d",
				params.MaximumExtraDataSize, headerExtraDataSize,
			)
		}
	}

	if b.ethBlock.Version() != 0 {
		return fmt.Errorf("invalid version: %d", b.ethBlock.Version())
	}

	// Check that the tx hash in the header matches the body
	txsHash := types.DeriveSha(b.ethBlock.Transactions(), new(trie.Trie))
	if txsHash != ethHeader.TxHash {
		return fmt.Errorf("invalid txs hash %v does not match calculated txs hash %v", ethHeader.TxHash, txsHash)
	}
	// Check that the uncle hash in the header matches the body
	uncleHash := types.CalcUncleHash(b.ethBlock.Uncles())
	if uncleHash != ethHeader.UncleHash {
		return fmt.Errorf("invalid uncle hash %v does not match calculated uncle hash %v", ethHeader.UncleHash, uncleHash)
	}
	// Coinbase must match the BlackholeAddr on C-Chain
	if ethHeader.Coinbase != constants.BlackholeAddr {
		return fmt.Errorf("invalid coinbase %v does not match required blackhole address %v", ethHeader.Coinbase, constants.BlackholeAddr)
	}
	// Block must not have any uncles
	if len(b.ethBlock.Uncles()) > 0 {
		return errUnclesUnsupported
	}

	// Block must not be empty
	txs := b.ethBlock.Transactions()
	if len(txs) == 0 && len(b.atomicTxs) == 0 {
		return errEmptyBlock
	}

	// Enforce minimum gas prices here prior to dynamic fees going into effect.
	if !rules.IsOdysseyPhase1 {
		// If we are prior to OdysseyPhase1, enforce each transaction has a minimum gas price of at least the OdysseyPhase1MinGasPrice
		for _, tx := range b.ethBlock.Transactions() {
			if tx.GasPrice().Cmp(odysseyPhase1MinGasPrice) < 0 {
				return fmt.Errorf("block contains tx %s with gas price too low (%d < %d)", tx.Hash(), tx.GasPrice(), params.OdysseyPhase1MinGasPrice)
			}
		}
    }

	// Make sure the block isn't too far in the future
	// TODO: move this to only be part of semantic verification.
	blockTimestamp := b.ethBlock.Time()
	if maxBlockTime := uint64(b.vm.clock.Time().Add(maxFutureBlockTime).Unix()); blockTimestamp > maxBlockTime {
		return fmt.Errorf("block timestamp is too far in the future: %d > allowed %d", blockTimestamp, maxBlockTime)
	}

	// Ensure BaseFee is non-nil as of OdysseyPhase1.
	if rules.IsOdysseyPhase1 {
		if ethHeader.BaseFee == nil {
			return errNilBaseFeeOdysseyPhase1
		}
		// TODO: this should be removed as 256 is the maximum possible bit length of a big int
		if bfLen := ethHeader.BaseFee.BitLen(); bfLen > 256 {
			return fmt.Errorf("too large base fee: bitlen %d", bfLen)
		}
	}

	// If we are in OdysseyPhase1, ensure that ExtDataGasUsed is populated correctly.
	if rules.IsOdysseyPhase1 {
		// Make sure ExtDataGasUsed is not nil and correct
		if ethHeader.ExtDataGasUsed == nil {
			return errNilExtDataGasUsedOdysseyPhase1
		}
        if ethHeader.ExtDataGasUsed.Cmp(params.AtomicGasLimit) == 1 {
            return fmt.Errorf("too large extDataGasUsed: %d", ethHeader.ExtDataGasUsed)
        }
		var totalGasUsed uint64
		for _, atomicTx := range b.atomicTxs {
			// We perform this check manually here to avoid the overhead of having to
			// reparse the atomicTx in `CalcExtDataGasUsed`.
			fixedFee := rules.IsOdysseyPhase1 // Charge the atomic tx fixed fee as of OdysseyPhase1
			gasUsed, err := atomicTx.GasUsed(fixedFee)
			if err != nil {
				return err
			}
			totalGasUsed, err = safemath.Add64(totalGasUsed, gasUsed)
			if err != nil {
				return err
			}
		}

		switch {
		case ethHeader.ExtDataGasUsed.Cmp(new(big.Int).SetUint64(totalGasUsed)) != 0:
			return fmt.Errorf("invalid extDataGasUsed: have %d, want %d", ethHeader.ExtDataGasUsed, totalGasUsed)

		// Make sure BlockGasCost is not nil
		// NOTE: ethHeader.BlockGasCost correctness is checked in header verification
		case ethHeader.BlockGasCost == nil:
			return errNilBlockGasCostOdysseyPhase1
		case !ethHeader.BlockGasCost.IsUint64():
			return fmt.Errorf("too large blockGasCost: %d", ethHeader.BlockGasCost)
		}
	}

	return nil
}
