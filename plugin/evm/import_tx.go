// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package evm

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/exp/slices"

	"github.com/DioneProtocol/coreth/core/state"
	"github.com/DioneProtocol/coreth/params"

	"github.com/DioneProtocol/odysseygo/chains/atomic"
	"github.com/DioneProtocol/odysseygo/ids"
	"github.com/DioneProtocol/odysseygo/snow"
	"github.com/DioneProtocol/odysseygo/utils"
	"github.com/DioneProtocol/odysseygo/utils/crypto/secp256k1"
	"github.com/DioneProtocol/odysseygo/utils/math"
	"github.com/DioneProtocol/odysseygo/utils/set"
	"github.com/DioneProtocol/odysseygo/vms/components/dione"
	"github.com/DioneProtocol/odysseygo/vms/components/verify"
	"github.com/DioneProtocol/odysseygo/vms/secp256k1fx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

var (
	_                           UnsignedAtomicTx       = &UnsignedImportTx{}
	_                           secp256k1fx.UnsignedTx = &UnsignedImportTx{}
	errImportNonDIONEInputBanff                         = errors.New("import input cannot contain non-DIONE in Banff")
	errImportNonDIONEOutputBanff                        = errors.New("import output cannot contain non-DIONE in Banff")
)

// UnsignedImportTx is an unsigned ImportTx
type UnsignedImportTx struct {
	dione.Metadata
	// ID of the network on which this tx was issued
	NetworkID uint32 `serialize:"true" json:"networkID"`
	// ID of this blockchain.
	BlockchainID ids.ID `serialize:"true" json:"blockchainID"`
	// Which chain to consume the funds from
	SourceChain ids.ID `serialize:"true" json:"sourceChain"`
	// Inputs that consume UTXOs produced on the chain
	ImportedInputs []*dione.TransferableInput `serialize:"true" json:"importedInputs"`
	// Outputs
	Outs []EVMOutput `serialize:"true" json:"outputs"`
}

// InputUTXOs returns the UTXOIDs of the imported funds
func (utx *UnsignedImportTx) InputUTXOs() set.Set[ids.ID] {
	set := set.NewSet[ids.ID](len(utx.ImportedInputs))
	for _, in := range utx.ImportedInputs {
		set.Add(in.InputID())
	}
	return set
}

// Verify this transaction is well-formed
func (utx *UnsignedImportTx) Verify(
	ctx *snow.Context,
	rules params.Rules,
) error {
	switch {
	case utx == nil:
		return errNilTx
	case len(utx.ImportedInputs) == 0:
		return errNoImportInputs
	case utx.NetworkID != ctx.NetworkID:
		return errWrongNetworkID
	case ctx.ChainID != utx.BlockchainID:
		return errWrongBlockchainID
	case rules.IsOdyPhase3 && len(utx.Outs) == 0:
		return errNoEVMOutputs
	}

	// Make sure that the tx has a valid peer chain ID
	if rules.IsOdyPhase5 {
		// Note that SameSubnet verifies that [tx.SourceChain] isn't this
		// chain's ID
		if err := verify.SameSubnet(context.TODO(), ctx, utx.SourceChain); err != nil {
			return errWrongChainID
		}
	} else {
		if utx.SourceChain != ctx.AChainID {
			return errWrongChainID
		}
	}

	for _, out := range utx.Outs {
		if err := out.Verify(); err != nil {
			return fmt.Errorf("EVM Output failed verification: %w", err)
		}
		if rules.IsBanff && out.AssetID != ctx.DIONEAssetID {
			return errImportNonDIONEOutputBanff
		}
	}

	for _, in := range utx.ImportedInputs {
		if err := in.Verify(); err != nil {
			return fmt.Errorf("atomic input failed verification: %w", err)
		}
		if rules.IsBanff && in.AssetID() != ctx.DIONEAssetID {
			return errImportNonDIONEInputBanff
		}
	}
	if !utils.IsSortedAndUnique(utx.ImportedInputs) {
		return errInputsNotSortedUnique
	}

	if rules.IsOdyPhase2 {
		if !utils.IsSortedAndUnique(utx.Outs) {
			return errOutputsNotSortedUnique
		}
	} else if rules.IsOdyPhase1 {
		if !slices.IsSortedFunc(utx.Outs, func(i, j EVMOutput) bool {
			return i.Less(j)
		}) {
			return errOutputsNotSorted
		}
	}

	return nil
}

func (utx *UnsignedImportTx) GasUsed(fixedFee bool) (uint64, error) {
	var (
		cost = calcBytesCost(len(utx.Bytes()))
		err  error
	)
	for _, in := range utx.ImportedInputs {
		inCost, err := in.In.Cost()
		if err != nil {
			return 0, err
		}
		cost, err = math.Add64(cost, inCost)
		if err != nil {
			return 0, err
		}
	}
	if fixedFee {
		cost, err = math.Add64(cost, params.AtomicTxBaseCost)
		if err != nil {
			return 0, err
		}
	}
	return cost, nil
}

// Amount of [assetID] burned by this transaction
func (utx *UnsignedImportTx) Burned(assetID ids.ID) (uint64, error) {
	var (
		spent uint64
		input uint64
		err   error
	)
	for _, out := range utx.Outs {
		if out.AssetID == assetID {
			spent, err = math.Add64(spent, out.Amount)
			if err != nil {
				return 0, err
			}
		}
	}
	for _, in := range utx.ImportedInputs {
		if in.AssetID() == assetID {
			input, err = math.Add64(input, in.Input().Amount())
			if err != nil {
				return 0, err
			}
		}
	}

	return math.Sub(input, spent)
}

// SemanticVerify this transaction is valid.
func (utx *UnsignedImportTx) SemanticVerify(
	vm *VM,
	stx *Tx,
	parent *Block,
	baseFee *big.Int,
	rules params.Rules,
) error {
	if err := utx.Verify(vm.ctx, rules); err != nil {
		return err
	}

	// Check the transaction consumes and produces the right amounts
	fc := dione.NewFlowChecker()
	switch {
	// Apply dynamic fees to import transactions as of Ody Phase 3
	case rules.IsOdyPhase3:
		gasUsed, err := stx.GasUsed(rules.IsOdyPhase5)
		if err != nil {
			return err
		}
		txFee, err := CalculateDynamicFee(gasUsed, baseFee)
		if err != nil {
			return err
		}
		fc.Produce(vm.ctx.DIONEAssetID, txFee)

	// Apply fees to import transactions as of Ody Phase 2
	case rules.IsOdyPhase2:
		fc.Produce(vm.ctx.DIONEAssetID, params.OdysseyAtomicTxFee)
	}
	for _, out := range utx.Outs {
		fc.Produce(out.AssetID, out.Amount)
	}
	for _, in := range utx.ImportedInputs {
		fc.Consume(in.AssetID(), in.Input().Amount())
	}

	if err := fc.Verify(); err != nil {
		return fmt.Errorf("import tx flow check failed due to: %w", err)
	}

	if len(stx.Creds) != len(utx.ImportedInputs) {
		return fmt.Errorf("import tx contained mismatched number of inputs/credentials (%d vs. %d)", len(utx.ImportedInputs), len(stx.Creds))
	}

	if !vm.bootstrapped {
		// Allow for force committing during bootstrapping
		return nil
	}

	utxoIDs := make([][]byte, len(utx.ImportedInputs))
	for i, in := range utx.ImportedInputs {
		inputID := in.UTXOID.InputID()
		utxoIDs[i] = inputID[:]
	}
	// allUTXOBytes is guaranteed to be the same length as utxoIDs
	allUTXOBytes, err := vm.ctx.SharedMemory.Get(utx.SourceChain, utxoIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch import UTXOs from %s due to: %w", utx.SourceChain, err)
	}

	for i, in := range utx.ImportedInputs {
		utxoBytes := allUTXOBytes[i]

		utxo := &dione.UTXO{}
		if _, err := vm.codec.Unmarshal(utxoBytes, utxo); err != nil {
			return fmt.Errorf("failed to unmarshal UTXO: %w", err)
		}

		cred := stx.Creds[i]

		utxoAssetID := utxo.AssetID()
		inAssetID := in.AssetID()
		if utxoAssetID != inAssetID {
			return errAssetIDMismatch
		}

		if err := vm.fx.VerifyTransfer(utx, in.In, cred, utxo.Out); err != nil {
			return fmt.Errorf("import tx transfer failed verification: %w", err)
		}
	}

	return vm.conflicts(utx.InputUTXOs(), parent)
}

// AtomicOps returns imported inputs spent on this transaction
// We spend imported UTXOs here rather than in semanticVerify because
// we don't want to remove an imported UTXO in semanticVerify
// only to have the transaction not be Accepted. This would be inconsistent.
// Recall that imported UTXOs are not kept in a versionDB.
func (utx *UnsignedImportTx) AtomicOps() (ids.ID, *atomic.Requests, error) {
	utxoIDs := make([][]byte, len(utx.ImportedInputs))
	for i, in := range utx.ImportedInputs {
		inputID := in.InputID()
		utxoIDs[i] = inputID[:]
	}
	return utx.SourceChain, &atomic.Requests{RemoveRequests: utxoIDs}, nil
}

// newImportTx returns a new ImportTx
func (vm *VM) newImportTx(
	chainID ids.ID, // chain to import from
	to common.Address, // Address of recipient
	baseFee *big.Int, // fee to use post-OP3
	keys []*secp256k1.PrivateKey, // Keys to import the funds
) (*Tx, error) {
	kc := secp256k1fx.NewKeychain()
	for _, key := range keys {
		kc.Add(key)
	}

	atomicUTXOs, _, _, err := vm.GetAtomicUTXOs(chainID, kc.Addresses(), ids.ShortEmpty, ids.Empty, -1)
	if err != nil {
		return nil, fmt.Errorf("problem retrieving atomic UTXOs: %w", err)
	}

	return vm.newImportTxWithUTXOs(chainID, to, baseFee, kc, atomicUTXOs)
}

// newImportTx returns a new ImportTx
func (vm *VM) newImportTxWithUTXOs(
	chainID ids.ID, // chain to import from
	to common.Address, // Address of recipient
	baseFee *big.Int, // fee to use post-OP3
	kc *secp256k1fx.Keychain, // Keychain to use for signing the atomic UTXOs
	atomicUTXOs []*dione.UTXO, // UTXOs to spend
) (*Tx, error) {
	importedInputs := []*dione.TransferableInput{}
	signers := [][]*secp256k1.PrivateKey{}

	importedAmount := make(map[ids.ID]uint64)
	now := vm.clock.Unix()
	for _, utxo := range atomicUTXOs {
		inputIntf, utxoSigners, err := kc.Spend(utxo.Out, now)
		if err != nil {
			continue
		}
		input, ok := inputIntf.(dione.TransferableIn)
		if !ok {
			continue
		}
		aid := utxo.AssetID()
		importedAmount[aid], err = math.Add64(importedAmount[aid], input.Amount())
		if err != nil {
			return nil, err
		}
		importedInputs = append(importedInputs, &dione.TransferableInput{
			UTXOID: utxo.UTXOID,
			Asset:  utxo.Asset,
			In:     input,
		})
		signers = append(signers, utxoSigners)
	}
	dione.SortTransferableInputsWithSigners(importedInputs, signers)
	importedDIONEAmount := importedAmount[vm.ctx.DIONEAssetID]

	outs := make([]EVMOutput, 0, len(importedAmount))
	// This will create unique outputs (in the context of sorting)
	// since each output will have a unique assetID
	for assetID, amount := range importedAmount {
		// Skip the DIONE amount since it is included separately to account for
		// the fee
		if assetID == vm.ctx.DIONEAssetID || amount == 0 {
			continue
		}
		outs = append(outs, EVMOutput{
			Address: to,
			Amount:  amount,
			AssetID: assetID,
		})
	}

	rules := vm.currentRules()

	var (
		txFeeWithoutChange uint64
		txFeeWithChange    uint64
	)
	switch {
	case rules.IsOdyPhase3:
		if baseFee == nil {
			return nil, errNilBaseFeeOdyPhase3
		}
		utx := &UnsignedImportTx{
			NetworkID:      vm.ctx.NetworkID,
			BlockchainID:   vm.ctx.ChainID,
			Outs:           outs,
			ImportedInputs: importedInputs,
			SourceChain:    chainID,
		}
		tx := &Tx{UnsignedAtomicTx: utx}
		if err := tx.Sign(vm.codec, nil); err != nil {
			return nil, err
		}

		gasUsedWithoutChange, err := tx.GasUsed(rules.IsOdyPhase5)
		if err != nil {
			return nil, err
		}
		gasUsedWithChange := gasUsedWithoutChange + EVMOutputGas

		txFeeWithoutChange, err = CalculateDynamicFee(gasUsedWithoutChange, baseFee)
		if err != nil {
			return nil, err
		}
		txFeeWithChange, err = CalculateDynamicFee(gasUsedWithChange, baseFee)
		if err != nil {
			return nil, err
		}
	case rules.IsOdyPhase2:
		txFeeWithoutChange = params.OdysseyAtomicTxFee
		txFeeWithChange = params.OdysseyAtomicTxFee
	}

	// DIONE output
	if importedDIONEAmount < txFeeWithoutChange { // imported amount goes toward paying tx fee
		return nil, errInsufficientFundsForFee
	}

	if importedDIONEAmount > txFeeWithChange {
		outs = append(outs, EVMOutput{
			Address: to,
			Amount:  importedDIONEAmount - txFeeWithChange,
			AssetID: vm.ctx.DIONEAssetID,
		})
	}

	// If no outputs are produced, return an error.
	// Note: this can happen if there is exactly enough DIONE to pay the
	// transaction fee, but no other funds to be imported.
	if len(outs) == 0 {
		return nil, errNoEVMOutputs
	}

	utils.Sort(outs)

	// Create the transaction
	utx := &UnsignedImportTx{
		NetworkID:      vm.ctx.NetworkID,
		BlockchainID:   vm.ctx.ChainID,
		Outs:           outs,
		ImportedInputs: importedInputs,
		SourceChain:    chainID,
	}
	tx := &Tx{UnsignedAtomicTx: utx}
	if err := tx.Sign(vm.codec, signers); err != nil {
		return nil, err
	}
	return tx, utx.Verify(vm.ctx, vm.currentRules())
}

// EVMStateTransfer performs the state transfer to increase the balances of
// accounts accordingly with the imported EVMOutputs
func (utx *UnsignedImportTx) EVMStateTransfer(ctx *snow.Context, state *state.StateDB) error {
	for _, to := range utx.Outs {
		if to.AssetID == ctx.DIONEAssetID {
			log.Debug("crosschain", "src", utx.SourceChain, "addr", to.Address, "amount", to.Amount, "assetID", "DIONE")
			// If the asset is DIONE, convert the input amount in nDIONE to gWei by
			// multiplying by the x2c rate.
			amount := new(big.Int).Mul(
				new(big.Int).SetUint64(to.Amount), x2cRate)
			state.AddBalance(to.Address, amount)
		} else {
			log.Debug("crosschain", "src", utx.SourceChain, "addr", to.Address, "amount", to.Amount, "assetID", to.AssetID)
			amount := new(big.Int).SetUint64(to.Amount)
			state.AddBalanceMultiCoin(to.Address, common.Hash(to.AssetID), amount)
		}
	}
	return nil
}
