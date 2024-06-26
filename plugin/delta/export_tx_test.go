// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package delta

import (
	"bytes"
	"context"
	"math/big"
	"testing"

	"github.com/DioneProtocol/coreth/params"
	"github.com/DioneProtocol/odysseygo/chains/atomic"
	"github.com/DioneProtocol/odysseygo/ids"
	engCommon "github.com/DioneProtocol/odysseygo/snow/engine/common"
	"github.com/DioneProtocol/odysseygo/utils/constants"
	"github.com/DioneProtocol/odysseygo/utils/crypto/secp256k1"
	"github.com/DioneProtocol/odysseygo/utils/units"
	"github.com/DioneProtocol/odysseygo/vms/components/dione"
	"github.com/DioneProtocol/odysseygo/vms/secp256k1fx"
	"github.com/ethereum/go-ethereum/common"
)

// createExportTxOptions adds funds to shared memory, imports them, and returns a list of export transactions
// that attempt to send the funds to each of the test keys (list of length 3).
func createExportTxOptions(t *testing.T, vm *VM, issuer chan engCommon.Message, sharedMemory *atomic.Memory) []*Tx {
	// Add a UTXO to shared memory
	utxo := &dione.UTXO{
		UTXOID: dione.UTXOID{TxID: ids.GenerateTestID()},
		Asset:  dione.Asset{ID: vm.ctx.DIONEAssetID},
		Out: &secp256k1fx.TransferOutput{
			Amt: uint64(50000000),
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{testKeys[0].PublicKey().Address()},
			},
		},
	}
	utxoBytes, err := vm.codec.Marshal(codecVersion, utxo)
	if err != nil {
		t.Fatal(err)
	}

	aChainSharedMemory := sharedMemory.NewSharedMemory(vm.ctx.AChainID)
	inputID := utxo.InputID()
	if err := aChainSharedMemory.Apply(map[ids.ID]*atomic.Requests{vm.ctx.ChainID: {PutRequests: []*atomic.Element{{
		Key:   inputID[:],
		Value: utxoBytes,
		Traits: [][]byte{
			testKeys[0].PublicKey().Address().Bytes(),
		},
	}}}}); err != nil {
		t.Fatal(err)
	}

	// Import the funds
	importTx, err := vm.newImportTx(vm.ctx.AChainID, testEthAddrs[0], initialBaseFee, []*secp256k1.PrivateKey{testKeys[0]})
	if err != nil {
		t.Fatal(err)
	}

	if err := vm.issueTx(importTx, true /*=local*/); err != nil {
		t.Fatal(err)
	}

	<-issuer

	blk, err := vm.BuildBlock(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err := blk.Verify(context.Background()); err != nil {
		t.Fatal(err)
	}

	if err := vm.SetPreference(context.Background(), blk.ID()); err != nil {
		t.Fatal(err)
	}

	if err := blk.Accept(context.Background()); err != nil {
		t.Fatal(err)
	}

	// Use the funds to create 3 conflicting export transactions sending the funds to each of the test addresses
	exportTxs := make([]*Tx, 0, 3)
	for _, addr := range testShortIDAddrs {
		exportTx, err := vm.newExportTx(vm.ctx.DIONEAssetID, uint64(5000000), vm.ctx.AChainID, addr, initialBaseFee, []*secp256k1.PrivateKey{testKeys[0]})
		if err != nil {
			t.Fatal(err)
		}
		exportTxs = append(exportTxs, exportTx)
	}

	return exportTxs
}

func TestExportTxDELTAStateTransfer(t *testing.T) {
	key := testKeys[0]
	addr := key.PublicKey().Address()
	ethAddr := GetEthAddress(key)

	dioneAmount := 50 * units.MilliDione
	dioneUTXOID := dione.UTXOID{
		OutputIndex: 0,
	}
	dioneInputID := dioneUTXOID.InputID()

	customAmount := uint64(100)
	customAssetID := ids.ID{1, 2, 3, 4, 5, 7}
	customUTXOID := dione.UTXOID{
		OutputIndex: 1,
	}
	customInputID := customUTXOID.InputID()

	customUTXO := &dione.UTXO{
		UTXOID: customUTXOID,
		Asset:  dione.Asset{ID: customAssetID},
		Out: &secp256k1fx.TransferOutput{
			Amt: customAmount,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{addr},
			},
		},
	}

	tests := []struct {
		name          string
		tx            []DELTAInput
		dioneBalance  *big.Int
		balances      map[ids.ID]*big.Int
		expectedNonce uint64
		shouldErr     bool
	}{
		{
			name:         "no transfers",
			tx:           nil,
			dioneBalance: big.NewInt(int64(dioneAmount) * x2cRateInt64),
			balances: map[ids.ID]*big.Int{
				customAssetID: big.NewInt(int64(customAmount)),
			},
			expectedNonce: 0,
			shouldErr:     false,
		},
		{
			name: "spend half DIONE",
			tx: []DELTAInput{
				{
					Address: ethAddr,
					Amount:  dioneAmount / 2,
					AssetID: testDioneAssetID,
					Nonce:   0,
				},
			},
			dioneBalance: big.NewInt(int64(dioneAmount/2) * x2cRateInt64),
			balances: map[ids.ID]*big.Int{
				customAssetID: big.NewInt(int64(customAmount)),
			},
			expectedNonce: 1,
			shouldErr:     false,
		},
		{
			name: "spend all DIONE",
			tx: []DELTAInput{
				{
					Address: ethAddr,
					Amount:  dioneAmount,
					AssetID: testDioneAssetID,
					Nonce:   0,
				},
			},
			dioneBalance: big.NewInt(0),
			balances: map[ids.ID]*big.Int{
				customAssetID: big.NewInt(int64(customAmount)),
			},
			expectedNonce: 1,
			shouldErr:     false,
		},
		{
			name: "spend too much DIONE",
			tx: []DELTAInput{
				{
					Address: ethAddr,
					Amount:  dioneAmount + 1,
					AssetID: testDioneAssetID,
					Nonce:   0,
				},
			},
			dioneBalance: big.NewInt(0),
			balances: map[ids.ID]*big.Int{
				customAssetID: big.NewInt(int64(customAmount)),
			},
			expectedNonce: 1,
			shouldErr:     true,
		},
		{
			name: "spend half custom",
			tx: []DELTAInput{
				{
					Address: ethAddr,
					Amount:  customAmount / 2,
					AssetID: customAssetID,
					Nonce:   0,
				},
			},
			dioneBalance: big.NewInt(int64(dioneAmount) * x2cRateInt64),
			balances: map[ids.ID]*big.Int{
				customAssetID: big.NewInt(int64(customAmount / 2)),
			},
			expectedNonce: 1,
			shouldErr:     false,
		},
		{
			name: "spend all custom",
			tx: []DELTAInput{
				{
					Address: ethAddr,
					Amount:  customAmount,
					AssetID: customAssetID,
					Nonce:   0,
				},
			},
			dioneBalance: big.NewInt(int64(dioneAmount) * x2cRateInt64),
			balances: map[ids.ID]*big.Int{
				customAssetID: big.NewInt(0),
			},
			expectedNonce: 1,
			shouldErr:     false,
		},
		{
			name: "spend too much custom",
			tx: []DELTAInput{
				{
					Address: ethAddr,
					Amount:  customAmount + 1,
					AssetID: customAssetID,
					Nonce:   0,
				},
			},
			dioneBalance: big.NewInt(int64(dioneAmount) * x2cRateInt64),
			balances: map[ids.ID]*big.Int{
				customAssetID: big.NewInt(0),
			},
			expectedNonce: 1,
			shouldErr:     true,
		},
		{
			name: "spend everything",
			tx: []DELTAInput{
				{
					Address: ethAddr,
					Amount:  customAmount,
					AssetID: customAssetID,
					Nonce:   0,
				},
				{
					Address: ethAddr,
					Amount:  dioneAmount,
					AssetID: testDioneAssetID,
					Nonce:   0,
				},
			},
			dioneBalance: big.NewInt(0),
			balances: map[ids.ID]*big.Int{
				customAssetID: big.NewInt(0),
			},
			expectedNonce: 1,
			shouldErr:     false,
		},
		{
			name: "spend everything wrong nonce",
			tx: []DELTAInput{
				{
					Address: ethAddr,
					Amount:  customAmount,
					AssetID: customAssetID,
					Nonce:   1,
				},
				{
					Address: ethAddr,
					Amount:  dioneAmount,
					AssetID: testDioneAssetID,
					Nonce:   1,
				},
			},
			dioneBalance: big.NewInt(0),
			balances: map[ids.ID]*big.Int{
				customAssetID: big.NewInt(0),
			},
			expectedNonce: 1,
			shouldErr:     true,
		},
		{
			name: "spend everything changing nonces",
			tx: []DELTAInput{
				{
					Address: ethAddr,
					Amount:  customAmount,
					AssetID: customAssetID,
					Nonce:   0,
				},
				{
					Address: ethAddr,
					Amount:  dioneAmount,
					AssetID: testDioneAssetID,
					Nonce:   1,
				},
			},
			dioneBalance: big.NewInt(0),
			balances: map[ids.ID]*big.Int{
				customAssetID: big.NewInt(0),
			},
			expectedNonce: 1,
			shouldErr:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issuer, vm, _, sharedMemory, _ := GenesisVM(t, true, genesisJSONApricotPhase0, "", "")
			defer func() {
				if err := vm.Shutdown(context.Background()); err != nil {
					t.Fatal(err)
				}
			}()

			dioneUTXO := &dione.UTXO{
				UTXOID: dioneUTXOID,
				Asset:  dione.Asset{ID: vm.ctx.DIONEAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: dioneAmount,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				},
			}

			dioneUTXOBytes, err := vm.codec.Marshal(codecVersion, dioneUTXO)
			if err != nil {
				t.Fatal(err)
			}

			customUTXOBytes, err := vm.codec.Marshal(codecVersion, customUTXO)
			if err != nil {
				t.Fatal(err)
			}

			aChainSharedMemory := sharedMemory.NewSharedMemory(vm.ctx.AChainID)
			if err := aChainSharedMemory.Apply(map[ids.ID]*atomic.Requests{vm.ctx.ChainID: {PutRequests: []*atomic.Element{
				{
					Key:   dioneInputID[:],
					Value: dioneUTXOBytes,
					Traits: [][]byte{
						addr.Bytes(),
					},
				},
				{
					Key:   customInputID[:],
					Value: customUTXOBytes,
					Traits: [][]byte{
						addr.Bytes(),
					},
				},
			}}}); err != nil {
				t.Fatal(err)
			}

			tx, err := vm.newImportTx(vm.ctx.AChainID, testEthAddrs[0], initialBaseFee, []*secp256k1.PrivateKey{testKeys[0]})
			if err != nil {
				t.Fatal(err)
			}

			if err := vm.issueTx(tx, true /*=local*/); err != nil {
				t.Fatal(err)
			}

			<-issuer

			blk, err := vm.BuildBlock(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			if err := blk.Verify(context.Background()); err != nil {
				t.Fatal(err)
			}

			if err := vm.SetPreference(context.Background(), blk.ID()); err != nil {
				t.Fatal(err)
			}

			if err := blk.Accept(context.Background()); err != nil {
				t.Fatal(err)
			}

			newTx := UnsignedExportTx{
				Ins: test.tx,
			}

			stateDB, err := vm.blockChain.State()
			if err != nil {
				t.Fatal(err)
			}

			err = newTx.DELTAStateTransfer(vm.ctx, stateDB)
			if test.shouldErr {
				if err == nil {
					t.Fatal("expected DELTAStateTransfer to fail")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			dioneBalance := stateDB.GetBalance(ethAddr)
			if dioneBalance.Cmp(test.dioneBalance) != 0 {
				t.Fatalf("address balance %s equal %s not %s", addr.String(), dioneBalance, test.dioneBalance)
			}

			for assetID, expectedBalance := range test.balances {
				balance := stateDB.GetBalanceMultiCoin(ethAddr, common.Hash(assetID))
				if dioneBalance.Cmp(test.dioneBalance) != 0 {
					t.Fatalf("%s address balance %s equal %s not %s", assetID, addr.String(), balance, expectedBalance)
				}
			}

			if stateDB.GetNonce(ethAddr) != test.expectedNonce {
				t.Fatalf("failed to set nonce to %d", test.expectedNonce)
			}
		})
	}
}

func TestExportTxSemanticVerify(t *testing.T) {
	_, vm, _, _, _ := GenesisVM(t, true, genesisJSONApricotPhase0, "", "")

	defer func() {
		if err := vm.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
	}()

	parent := vm.LastAcceptedBlockInternal().(*Block)

	key := testKeys[0]
	addr := key.PublicKey().Address()
	ethAddr := testEthAddrs[0]

	var (
		dioneBalance          = 10 * units.Dione
		custom0Balance uint64 = 100
		custom0AssetID        = ids.ID{1, 2, 3, 4, 5}
		custom1Balance uint64 = 1000
		custom1AssetID        = ids.ID{1, 2, 3, 4, 5, 6}
	)

	validExportTx := &UnsignedExportTx{
		NetworkID:        vm.ctx.NetworkID,
		BlockchainID:     vm.ctx.ChainID,
		DestinationChain: vm.ctx.AChainID,
		Ins: []DELTAInput{
			{
				Address: ethAddr,
				Amount:  dioneBalance,
				AssetID: vm.ctx.DIONEAssetID,
				Nonce:   0,
			},
			{
				Address: ethAddr,
				Amount:  custom0Balance,
				AssetID: custom0AssetID,
				Nonce:   0,
			},
			{
				Address: ethAddr,
				Amount:  custom1Balance,
				AssetID: custom1AssetID,
				Nonce:   0,
			},
		},
		ExportedOutputs: []*dione.TransferableOutput{
			{
				Asset: dione.Asset{ID: custom0AssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: custom0Balance,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				},
			},
		},
	}

	validDIONEExportTx := &UnsignedExportTx{
		NetworkID:        vm.ctx.NetworkID,
		BlockchainID:     vm.ctx.ChainID,
		DestinationChain: vm.ctx.AChainID,
		Ins: []DELTAInput{
			{
				Address: ethAddr,
				Amount:  dioneBalance,
				AssetID: vm.ctx.DIONEAssetID,
				Nonce:   0,
			},
		},
		ExportedOutputs: []*dione.TransferableOutput{
			{
				Asset: dione.Asset{ID: vm.ctx.DIONEAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: dioneBalance / 2,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		tx        *Tx
		signers   [][]*secp256k1.PrivateKey
		baseFee   *big.Int
		rules     params.Rules
		shouldErr bool
	}{
		{
			name: "valid",
			tx:   &Tx{UnsignedAtomicTx: validExportTx},
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: false,
		},
		{
			name: "O-chain before AP5",
			tx: func() *Tx {
				validExportTx := *validDIONEExportTx
				validExportTx.DestinationChain = constants.OmegaChainID
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "O-chain after AP5",
			tx: func() *Tx {
				validExportTx := *validDIONEExportTx
				validExportTx.DestinationChain = constants.OmegaChainID
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase5,
			shouldErr: false,
		},
		{
			name: "random chain after AP5",
			tx: func() *Tx {
				validExportTx := *validDIONEExportTx
				validExportTx.DestinationChain = ids.GenerateTestID()
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase5,
			shouldErr: true,
		},
		{
			name: "O-chain multi-coin before AP5",
			tx: func() *Tx {
				validExportTx := *validExportTx
				validExportTx.DestinationChain = constants.OmegaChainID
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "O-chain multi-coin after AP5",
			tx: func() *Tx {
				validExportTx := *validExportTx
				validExportTx.DestinationChain = constants.OmegaChainID
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase5,
			shouldErr: true,
		},
		{
			name: "random chain multi-coin  after AP5",
			tx: func() *Tx {
				validExportTx := *validExportTx
				validExportTx.DestinationChain = ids.GenerateTestID()
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase5,
			shouldErr: true,
		},
		{
			name: "no outputs",
			tx: func() *Tx {
				validExportTx := *validExportTx
				validExportTx.ExportedOutputs = nil
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "wrong networkID",
			tx: func() *Tx {
				validExportTx := *validExportTx
				validExportTx.NetworkID++
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "wrong chainID",
			tx: func() *Tx {
				validExportTx := *validExportTx
				validExportTx.BlockchainID = ids.GenerateTestID()
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "invalid input",
			tx: func() *Tx {
				validExportTx := *validExportTx
				validExportTx.Ins = append([]DELTAInput{}, validExportTx.Ins...)
				validExportTx.Ins[2].Amount = 0
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "invalid output",
			tx: func() *Tx {
				validExportTx := *validExportTx
				validExportTx.ExportedOutputs = []*dione.TransferableOutput{{
					Asset: dione.Asset{ID: custom0AssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: custom0Balance,
						OutputOwners: secp256k1fx.OutputOwners{
							Threshold: 0,
							Addrs:     []ids.ShortID{addr},
						},
					},
				}}
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "unsorted outputs",
			tx: func() *Tx {
				validExportTx := *validExportTx
				exportOutputs := []*dione.TransferableOutput{
					{
						Asset: dione.Asset{ID: custom0AssetID},
						Out: &secp256k1fx.TransferOutput{
							Amt: custom0Balance/2 + 1,
							OutputOwners: secp256k1fx.OutputOwners{
								Threshold: 1,
								Addrs:     []ids.ShortID{addr},
							},
						},
					},
					{
						Asset: dione.Asset{ID: custom0AssetID},
						Out: &secp256k1fx.TransferOutput{
							Amt: custom0Balance/2 - 1,
							OutputOwners: secp256k1fx.OutputOwners{
								Threshold: 1,
								Addrs:     []ids.ShortID{addr},
							},
						},
					},
				}
				// Sort the outputs and then swap the ordering to ensure that they are ordered incorrectly
				dione.SortTransferableOutputs(exportOutputs, Codec)
				exportOutputs[0], exportOutputs[1] = exportOutputs[1], exportOutputs[0]
				validExportTx.ExportedOutputs = exportOutputs
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "not unique inputs",
			tx: func() *Tx {
				validExportTx := *validExportTx
				validExportTx.Ins = append([]DELTAInput{}, validExportTx.Ins...)
				validExportTx.Ins[2] = validExportTx.Ins[1]
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "custom asset insufficient funds",
			tx: func() *Tx {
				validExportTx := *validExportTx
				validExportTx.ExportedOutputs = []*dione.TransferableOutput{
					{
						Asset: dione.Asset{ID: custom0AssetID},
						Out: &secp256k1fx.TransferOutput{
							Amt: custom0Balance + 1,
							OutputOwners: secp256k1fx.OutputOwners{
								Threshold: 1,
								Addrs:     []ids.ShortID{addr},
							},
						},
					},
				}
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "dione insufficient funds",
			tx: func() *Tx {
				validExportTx := *validExportTx
				validExportTx.ExportedOutputs = []*dione.TransferableOutput{
					{
						Asset: dione.Asset{ID: vm.ctx.DIONEAssetID},
						Out: &secp256k1fx.TransferOutput{
							Amt: dioneBalance, // after fees this should be too much
							OutputOwners: secp256k1fx.OutputOwners{
								Threshold: 1,
								Addrs:     []ids.ShortID{addr},
							},
						},
					},
				}
				return &Tx{UnsignedAtomicTx: &validExportTx}
			}(),
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "too many signatures",
			tx:   &Tx{UnsignedAtomicTx: validExportTx},
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "too few signatures",
			tx:   &Tx{UnsignedAtomicTx: validExportTx},
			signers: [][]*secp256k1.PrivateKey{
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "too many signatures on credential",
			tx:   &Tx{UnsignedAtomicTx: validExportTx},
			signers: [][]*secp256k1.PrivateKey{
				{key, testKeys[1]},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "too few signatures on credential",
			tx:   &Tx{UnsignedAtomicTx: validExportTx},
			signers: [][]*secp256k1.PrivateKey{
				{},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name: "wrong signature on credential",
			tx:   &Tx{UnsignedAtomicTx: validExportTx},
			signers: [][]*secp256k1.PrivateKey{
				{testKeys[1]},
				{key},
				{key},
			},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
		{
			name:      "no signatures",
			tx:        &Tx{UnsignedAtomicTx: validExportTx},
			signers:   [][]*secp256k1.PrivateKey{},
			baseFee:   initialBaseFee,
			rules:     apricotRulesPhase3,
			shouldErr: true,
		},
	}
	for _, test := range tests {
		if err := test.tx.Sign(vm.codec, test.signers); err != nil {
			t.Fatal(err)
		}

		t.Run(test.name, func(t *testing.T) {
			tx := test.tx
			exportTx := tx.UnsignedAtomicTx

			err := exportTx.SemanticVerify(vm, tx, parent, test.baseFee, test.rules)
			if test.shouldErr && err == nil {
				t.Fatalf("should have errored but returned valid")
			}
			if !test.shouldErr && err != nil {
				t.Fatalf("shouldn't have errored but returned %s", err)
			}
		})
	}
}

func TestExportTxAccept(t *testing.T) {
	_, vm, _, sharedMemory, _ := GenesisVM(t, true, genesisJSONApricotPhase0, "", "")

	aChainSharedMemory := sharedMemory.NewSharedMemory(vm.ctx.AChainID)

	defer func() {
		if err := vm.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
	}()

	key := testKeys[0]
	addr := key.PublicKey().Address()
	ethAddr := testEthAddrs[0]

	var (
		dioneBalance          = 10 * units.Dione
		custom0Balance uint64 = 100
		custom0AssetID        = ids.ID{1, 2, 3, 4, 5}
	)

	exportTx := &UnsignedExportTx{
		NetworkID:        vm.ctx.NetworkID,
		BlockchainID:     vm.ctx.ChainID,
		DestinationChain: vm.ctx.AChainID,
		Ins: []DELTAInput{
			{
				Address: ethAddr,
				Amount:  dioneBalance,
				AssetID: vm.ctx.DIONEAssetID,
				Nonce:   0,
			},
			{
				Address: ethAddr,
				Amount:  custom0Balance,
				AssetID: custom0AssetID,
				Nonce:   0,
			},
		},
		ExportedOutputs: []*dione.TransferableOutput{
			{
				Asset: dione.Asset{ID: vm.ctx.DIONEAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: dioneBalance,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				},
			},
			{
				Asset: dione.Asset{ID: custom0AssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: custom0Balance,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				},
			},
		},
	}

	tx := &Tx{UnsignedAtomicTx: exportTx}

	signers := [][]*secp256k1.PrivateKey{
		{key},
		{key},
		{key},
	}

	if err := tx.Sign(vm.codec, signers); err != nil {
		t.Fatal(err)
	}

	commitBatch, err := vm.db.CommitBatch()
	if err != nil {
		t.Fatalf("Failed to create commit batch for VM due to %s", err)
	}
	chainID, atomicRequests, err := tx.AtomicOps()
	if err != nil {
		t.Fatalf("Failed to accept export transaction due to: %s", err)
	}

	if err := vm.ctx.SharedMemory.Apply(map[ids.ID]*atomic.Requests{chainID: {PutRequests: atomicRequests.PutRequests}}, commitBatch); err != nil {
		t.Fatal(err)
	}
	indexedValues, _, _, err := aChainSharedMemory.Indexed(vm.ctx.ChainID, [][]byte{addr.Bytes()}, nil, nil, 3)
	if err != nil {
		t.Fatal(err)
	}

	if len(indexedValues) != 2 {
		t.Fatalf("expected 2 values but got %d", len(indexedValues))
	}

	dioneUTXOID := dione.UTXOID{
		TxID:        tx.ID(),
		OutputIndex: 0,
	}
	dioneInputID := dioneUTXOID.InputID()

	customUTXOID := dione.UTXOID{
		TxID:        tx.ID(),
		OutputIndex: 1,
	}
	customInputID := customUTXOID.InputID()

	fetchedValues, err := aChainSharedMemory.Get(vm.ctx.ChainID, [][]byte{
		customInputID[:],
		dioneInputID[:],
	})
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(fetchedValues[0], indexedValues[0]) {
		t.Fatalf("inconsistent values returned fetched %x indexed %x", fetchedValues[0], indexedValues[0])
	}
	if !bytes.Equal(fetchedValues[1], indexedValues[1]) {
		t.Fatalf("inconsistent values returned fetched %x indexed %x", fetchedValues[1], indexedValues[1])
	}

	customUTXOBytes, err := Codec.Marshal(codecVersion, &dione.UTXO{
		UTXOID: customUTXOID,
		Asset:  dione.Asset{ID: custom0AssetID},
		Out:    exportTx.ExportedOutputs[1].Out,
	})
	if err != nil {
		t.Fatal(err)
	}

	dioneUTXOBytes, err := Codec.Marshal(codecVersion, &dione.UTXO{
		UTXOID: dioneUTXOID,
		Asset:  dione.Asset{ID: vm.ctx.DIONEAssetID},
		Out:    exportTx.ExportedOutputs[0].Out,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(fetchedValues[0], customUTXOBytes) {
		t.Fatalf("incorrect values returned expected %x got %x", customUTXOBytes, fetchedValues[0])
	}
	if !bytes.Equal(fetchedValues[1], dioneUTXOBytes) {
		t.Fatalf("incorrect values returned expected %x got %x", dioneUTXOBytes, fetchedValues[1])
	}
}

func TestExportTxVerify(t *testing.T) {
	var exportAmount uint64 = 10000000
	exportTx := &UnsignedExportTx{
		NetworkID:        testNetworkID,
		BlockchainID:     testDChainID,
		DestinationChain: testAChainID,
		Ins: []DELTAInput{
			{
				Address: testEthAddrs[0],
				Amount:  exportAmount,
				AssetID: testDioneAssetID,
				Nonce:   0,
			},
			{
				Address: testEthAddrs[2],
				Amount:  exportAmount,
				AssetID: testDioneAssetID,
				Nonce:   0,
			},
		},
		ExportedOutputs: []*dione.TransferableOutput{
			{
				Asset: dione.Asset{ID: testDioneAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: exportAmount,
					OutputOwners: secp256k1fx.OutputOwners{
						Locktime:  0,
						Threshold: 1,
						Addrs:     []ids.ShortID{testShortIDAddrs[0]},
					},
				},
			},
			{
				Asset: dione.Asset{ID: testDioneAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: exportAmount,
					OutputOwners: secp256k1fx.OutputOwners{
						Locktime:  0,
						Threshold: 1,
						Addrs:     []ids.ShortID{testShortIDAddrs[1]},
					},
				},
			},
		},
	}

	// Sort the inputs and outputs to ensure the transaction is canonical
	dione.SortTransferableOutputs(exportTx.ExportedOutputs, Codec)
	// Pass in a list of signers here with the appropriate length
	// to avoid causing a nil-pointer error in the helper method
	emptySigners := make([][]*secp256k1.PrivateKey, 2)
	SortDELTAInputsAndSigners(exportTx.Ins, emptySigners)

	ctx := NewContext()

	tests := map[string]atomicTxVerifyTest{
		"nil tx": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				return (*UnsignedExportTx)(nil)
			},
			ctx:         ctx,
			rules:       apricotRulesPhase0,
			expectedErr: errNilTx.Error(),
		},
		"valid export tx": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				return exportTx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase0,
			expectedErr: "",
		},
		"valid export tx banff": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				return exportTx
			},
			ctx:         ctx,
			rules:       banffRules,
			expectedErr: "",
		},
		"incorrect networkID": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.NetworkID++
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase0,
			expectedErr: errWrongNetworkID.Error(),
		},
		"incorrect blockchainID": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.BlockchainID = nonExistentID
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase0,
			expectedErr: errWrongBlockchainID.Error(),
		},
		"incorrect destination chain": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.DestinationChain = nonExistentID
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase0,
			expectedErr: errWrongChainID.Error(), // TODO make this error more specific to destination not just chainID
		},
		"no exported outputs": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.ExportedOutputs = nil
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase0,
			expectedErr: errNoExportOutputs.Error(),
		},
		"unsorted outputs": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.ExportedOutputs = []*dione.TransferableOutput{
					tx.ExportedOutputs[1],
					tx.ExportedOutputs[0],
				}
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase0,
			expectedErr: errOutputsNotSorted.Error(),
		},
		"invalid exported output": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.ExportedOutputs = []*dione.TransferableOutput{tx.ExportedOutputs[0], nil}
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase0,
			expectedErr: "nil transferable output is not valid",
		},
		"unsorted DELTA inputs before AP1": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.Ins = []DELTAInput{
					tx.Ins[1],
					tx.Ins[0],
				}
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase0,
			expectedErr: "",
		},
		"unsorted DELTA inputs after AP1": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.Ins = []DELTAInput{
					tx.Ins[1],
					tx.Ins[0],
				}
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase1,
			expectedErr: errInputsNotSortedUnique.Error(),
		},
		"DELTA input with amount 0": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.Ins = []DELTAInput{
					{
						Address: testEthAddrs[0],
						Amount:  0,
						AssetID: testDioneAssetID,
						Nonce:   0,
					},
				}
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase0,
			expectedErr: errNoValueInput.Error(),
		},
		"non-unique DELTA input before AP1": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.Ins = []DELTAInput{tx.Ins[0], tx.Ins[0]}
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase0,
			expectedErr: "",
		},
		"non-unique DELTA input after AP1": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.Ins = []DELTAInput{tx.Ins[0], tx.Ins[0]}
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase1,
			expectedErr: errInputsNotSortedUnique.Error(),
		},
		"non-DIONE input Apricot Phase 6": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.Ins = []DELTAInput{
					{
						Address: testEthAddrs[0],
						Amount:  1,
						AssetID: nonExistentID,
						Nonce:   0,
					},
				}
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase6,
			expectedErr: "",
		},
		"non-DIONE output Apricot Phase 6": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.ExportedOutputs = []*dione.TransferableOutput{
					{
						Asset: dione.Asset{ID: nonExistentID},
						Out: &secp256k1fx.TransferOutput{
							Amt: exportAmount,
							OutputOwners: secp256k1fx.OutputOwners{
								Locktime:  0,
								Threshold: 1,
								Addrs:     []ids.ShortID{testShortIDAddrs[0]},
							},
						},
					},
				}
				return &tx
			},
			ctx:         ctx,
			rules:       apricotRulesPhase6,
			expectedErr: "",
		},
		"non-DIONE input Banff": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.Ins = []DELTAInput{
					{
						Address: testEthAddrs[0],
						Amount:  1,
						AssetID: nonExistentID,
						Nonce:   0,
					},
				}
				return &tx
			},
			ctx:         ctx,
			rules:       banffRules,
			expectedErr: errExportNonDIONEInputBanff.Error(),
		},
		"non-DIONE output Banff": {
			generate: func(t *testing.T) UnsignedAtomicTx {
				tx := *exportTx
				tx.ExportedOutputs = []*dione.TransferableOutput{
					{
						Asset: dione.Asset{ID: nonExistentID},
						Out: &secp256k1fx.TransferOutput{
							Amt: exportAmount,
							OutputOwners: secp256k1fx.OutputOwners{
								Locktime:  0,
								Threshold: 1,
								Addrs:     []ids.ShortID{testShortIDAddrs[0]},
							},
						},
					},
				}
				return &tx
			},
			ctx:         ctx,
			rules:       banffRules,
			expectedErr: errExportNonDIONEOutputBanff.Error(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			executeTxVerifyTest(t, test)
		})
	}
}

// Note: this is a brittle test to ensure that the gas cost of a transaction does
// not change
func TestExportTxGasCost(t *testing.T) {
	dioneAssetID := ids.GenerateTestID()
	chainID := ids.GenerateTestID()
	aChainID := ids.GenerateTestID()
	networkID := uint32(5)
	exportAmount := uint64(5000000)

	tests := map[string]struct {
		UnsignedExportTx *UnsignedExportTx
		Keys             [][]*secp256k1.PrivateKey

		BaseFee         *big.Int
		ExpectedGasUsed uint64
		ExpectedFee     uint64
		FixedFee        bool
	}{
		"simple export 1wei BaseFee": {
			UnsignedExportTx: &UnsignedExportTx{
				NetworkID:        networkID,
				BlockchainID:     chainID,
				DestinationChain: aChainID,
				Ins: []DELTAInput{
					{
						Address: testEthAddrs[0],
						Amount:  exportAmount,
						AssetID: dioneAssetID,
						Nonce:   0,
					},
				},
				ExportedOutputs: []*dione.TransferableOutput{
					{
						Asset: dione.Asset{ID: dioneAssetID},
						Out: &secp256k1fx.TransferOutput{
							Amt: exportAmount,
							OutputOwners: secp256k1fx.OutputOwners{
								Locktime:  0,
								Threshold: 1,
								Addrs:     []ids.ShortID{testShortIDAddrs[0]},
							},
						},
					},
				},
			},
			Keys:            [][]*secp256k1.PrivateKey{{testKeys[0]}},
			ExpectedGasUsed: 1230,
			ExpectedFee:     1,
			BaseFee:         big.NewInt(1),
		},
		"simple export 1wei BaseFee + fixed fee": {
			UnsignedExportTx: &UnsignedExportTx{
				NetworkID:        networkID,
				BlockchainID:     chainID,
				DestinationChain: aChainID,
				Ins: []DELTAInput{
					{
						Address: testEthAddrs[0],
						Amount:  exportAmount,
						AssetID: dioneAssetID,
						Nonce:   0,
					},
				},
				ExportedOutputs: []*dione.TransferableOutput{
					{
						Asset: dione.Asset{ID: dioneAssetID},
						Out: &secp256k1fx.TransferOutput{
							Amt: exportAmount,
							OutputOwners: secp256k1fx.OutputOwners{
								Locktime:  0,
								Threshold: 1,
								Addrs:     []ids.ShortID{testShortIDAddrs[0]},
							},
						},
					},
				},
			},
			Keys:            [][]*secp256k1.PrivateKey{{testKeys[0]}},
			ExpectedGasUsed: 11230,
			ExpectedFee:     1,
			BaseFee:         big.NewInt(1),
			FixedFee:        true,
		},
		"simple export 25Gwei BaseFee": {
			UnsignedExportTx: &UnsignedExportTx{
				NetworkID:        networkID,
				BlockchainID:     chainID,
				DestinationChain: aChainID,
				Ins: []DELTAInput{
					{
						Address: testEthAddrs[0],
						Amount:  exportAmount,
						AssetID: dioneAssetID,
						Nonce:   0,
					},
				},
				ExportedOutputs: []*dione.TransferableOutput{
					{
						Asset: dione.Asset{ID: dioneAssetID},
						Out: &secp256k1fx.TransferOutput{
							Amt: exportAmount,
							OutputOwners: secp256k1fx.OutputOwners{
								Locktime:  0,
								Threshold: 1,
								Addrs:     []ids.ShortID{testShortIDAddrs[0]},
							},
						},
					},
				},
			},
			Keys:            [][]*secp256k1.PrivateKey{{testKeys[0]}},
			ExpectedGasUsed: 1230,
			ExpectedFee:     30750,
			BaseFee:         big.NewInt(25 * params.GWei),
		},
		"simple export 225Gwei BaseFee": {
			UnsignedExportTx: &UnsignedExportTx{
				NetworkID:        networkID,
				BlockchainID:     chainID,
				DestinationChain: aChainID,
				Ins: []DELTAInput{
					{
						Address: testEthAddrs[0],
						Amount:  exportAmount,
						AssetID: dioneAssetID,
						Nonce:   0,
					},
				},
				ExportedOutputs: []*dione.TransferableOutput{
					{
						Asset: dione.Asset{ID: dioneAssetID},
						Out: &secp256k1fx.TransferOutput{
							Amt: exportAmount,
							OutputOwners: secp256k1fx.OutputOwners{
								Locktime:  0,
								Threshold: 1,
								Addrs:     []ids.ShortID{testShortIDAddrs[0]},
							},
						},
					},
				},
			},
			Keys:            [][]*secp256k1.PrivateKey{{testKeys[0]}},
			ExpectedGasUsed: 1230,
			ExpectedFee:     276750,
			BaseFee:         big.NewInt(225 * params.GWei),
		},
		"complex export 25Gwei BaseFee": {
			UnsignedExportTx: &UnsignedExportTx{
				NetworkID:        networkID,
				BlockchainID:     chainID,
				DestinationChain: aChainID,
				Ins: []DELTAInput{
					{
						Address: testEthAddrs[0],
						Amount:  exportAmount,
						AssetID: dioneAssetID,
						Nonce:   0,
					},
					{
						Address: testEthAddrs[1],
						Amount:  exportAmount,
						AssetID: dioneAssetID,
						Nonce:   0,
					},
					{
						Address: testEthAddrs[2],
						Amount:  exportAmount,
						AssetID: dioneAssetID,
						Nonce:   0,
					},
				},
				ExportedOutputs: []*dione.TransferableOutput{
					{
						Asset: dione.Asset{ID: dioneAssetID},
						Out: &secp256k1fx.TransferOutput{
							Amt: exportAmount * 3,
							OutputOwners: secp256k1fx.OutputOwners{
								Locktime:  0,
								Threshold: 1,
								Addrs:     []ids.ShortID{testShortIDAddrs[0]},
							},
						},
					},
				},
			},
			Keys:            [][]*secp256k1.PrivateKey{{testKeys[0], testKeys[0], testKeys[0]}},
			ExpectedGasUsed: 3366,
			ExpectedFee:     84150,
			BaseFee:         big.NewInt(25 * params.GWei),
		},
		"complex export 225Gwei BaseFee": {
			UnsignedExportTx: &UnsignedExportTx{
				NetworkID:        networkID,
				BlockchainID:     chainID,
				DestinationChain: aChainID,
				Ins: []DELTAInput{
					{
						Address: testEthAddrs[0],
						Amount:  exportAmount,
						AssetID: dioneAssetID,
						Nonce:   0,
					},
					{
						Address: testEthAddrs[1],
						Amount:  exportAmount,
						AssetID: dioneAssetID,
						Nonce:   0,
					},
					{
						Address: testEthAddrs[2],
						Amount:  exportAmount,
						AssetID: dioneAssetID,
						Nonce:   0,
					},
				},
				ExportedOutputs: []*dione.TransferableOutput{
					{
						Asset: dione.Asset{ID: dioneAssetID},
						Out: &secp256k1fx.TransferOutput{
							Amt: exportAmount * 3,
							OutputOwners: secp256k1fx.OutputOwners{
								Locktime:  0,
								Threshold: 1,
								Addrs:     []ids.ShortID{testShortIDAddrs[0]},
							},
						},
					},
				},
			},
			Keys:            [][]*secp256k1.PrivateKey{{testKeys[0], testKeys[0], testKeys[0]}},
			ExpectedGasUsed: 3366,
			ExpectedFee:     757350,
			BaseFee:         big.NewInt(225 * params.GWei),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tx := &Tx{UnsignedAtomicTx: test.UnsignedExportTx}

			// Sign with the correct key
			if err := tx.Sign(Codec, test.Keys); err != nil {
				t.Fatal(err)
			}

			gasUsed, err := tx.GasUsed(test.FixedFee)
			if err != nil {
				t.Fatal(err)
			}
			if gasUsed != test.ExpectedGasUsed {
				t.Fatalf("Expected gasUsed to be %d, but found %d", test.ExpectedGasUsed, gasUsed)
			}

			fee, err := CalculateDynamicFee(gasUsed, test.BaseFee)
			if err != nil {
				t.Fatal(err)
			}
			if fee != test.ExpectedFee {
				t.Fatalf("Expected fee to be %d, but found %d", test.ExpectedFee, fee)
			}
		})
	}
}

func TestNewExportTx(t *testing.T) {
	tests := []struct {
		name                string
		genesis             string
		rules               params.Rules
		bal                 uint64
		expectedBurnedDIONE uint64
	}{
		{
			name:                "apricot phase 0",
			genesis:             genesisJSONApricotPhase0,
			rules:               apricotRulesPhase0,
			bal:                 44000000,
			expectedBurnedDIONE: 1000000,
		},
		{
			name:                "apricot phase 1",
			genesis:             genesisJSONApricotPhase1,
			rules:               apricotRulesPhase1,
			bal:                 44000000,
			expectedBurnedDIONE: 1000000,
		},
		{
			name:                "apricot phase 2",
			genesis:             genesisJSONApricotPhase2,
			rules:               apricotRulesPhase2,
			bal:                 43000000,
			expectedBurnedDIONE: 1000000,
		},
		{
			name:                "apricot phase 3",
			genesis:             genesisJSONApricotPhase3,
			rules:               apricotRulesPhase3,
			bal:                 44446500,
			expectedBurnedDIONE: 276750,
		},
		{
			name:                "apricot phase 4",
			genesis:             genesisJSONApricotPhase4,
			rules:               apricotRulesPhase4,
			bal:                 44446500,
			expectedBurnedDIONE: 276750,
		},
		{
			name:                "apricot phase 5",
			genesis:             genesisJSONApricotPhase5,
			rules:               apricotRulesPhase5,
			bal:                 39946500,
			expectedBurnedDIONE: 2526750,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issuer, vm, _, sharedMemory, _ := GenesisVM(t, true, test.genesis, "", "")

			defer func() {
				if err := vm.Shutdown(context.Background()); err != nil {
					t.Fatal(err)
				}
			}()

			parent := vm.LastAcceptedBlockInternal().(*Block)
			importAmount := uint64(50000000)
			utxoID := dione.UTXOID{TxID: ids.GenerateTestID()}

			utxo := &dione.UTXO{
				UTXOID: utxoID,
				Asset:  dione.Asset{ID: vm.ctx.DIONEAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: importAmount,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{testKeys[0].PublicKey().Address()},
					},
				},
			}
			utxoBytes, err := vm.codec.Marshal(codecVersion, utxo)
			if err != nil {
				t.Fatal(err)
			}

			aChainSharedMemory := sharedMemory.NewSharedMemory(vm.ctx.AChainID)
			inputID := utxo.InputID()
			if err := aChainSharedMemory.Apply(map[ids.ID]*atomic.Requests{vm.ctx.ChainID: {PutRequests: []*atomic.Element{{
				Key:   inputID[:],
				Value: utxoBytes,
				Traits: [][]byte{
					testKeys[0].PublicKey().Address().Bytes(),
				},
			}}}}); err != nil {
				t.Fatal(err)
			}

			tx, err := vm.newImportTx(vm.ctx.AChainID, testEthAddrs[0], initialBaseFee, []*secp256k1.PrivateKey{testKeys[0]})
			if err != nil {
				t.Fatal(err)
			}

			if err := vm.issueTx(tx, true /*=local*/); err != nil {
				t.Fatal(err)
			}

			<-issuer

			blk, err := vm.BuildBlock(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			if err := blk.Verify(context.Background()); err != nil {
				t.Fatal(err)
			}

			if err := vm.SetPreference(context.Background(), blk.ID()); err != nil {
				t.Fatal(err)
			}

			if err := blk.Accept(context.Background()); err != nil {
				t.Fatal(err)
			}

			parent = vm.LastAcceptedBlockInternal().(*Block)
			exportAmount := uint64(5000000)

			tx, err = vm.newExportTx(vm.ctx.DIONEAssetID, exportAmount, vm.ctx.AChainID, testShortIDAddrs[0], initialBaseFee, []*secp256k1.PrivateKey{testKeys[0]})
			if err != nil {
				t.Fatal(err)
			}

			exportTx := tx.UnsignedAtomicTx

			if err := exportTx.SemanticVerify(vm, tx, parent, parent.ethBlock.BaseFee(), test.rules); err != nil {
				t.Fatal("newExportTx created an invalid transaction", err)
			}

			burnedDIONE, err := exportTx.Burned(vm.ctx.DIONEAssetID)
			if err != nil {
				t.Fatal(err)
			}
			if burnedDIONE != test.expectedBurnedDIONE {
				t.Fatalf("burned wrong amount of DIONE - expected %d burned %d", test.expectedBurnedDIONE, burnedDIONE)
			}

			commitBatch, err := vm.db.CommitBatch()
			if err != nil {
				t.Fatalf("Failed to create commit batch for VM due to %s", err)
			}
			chainID, atomicRequests, err := exportTx.AtomicOps()

			if err != nil {
				t.Fatalf("Failed to accept export transaction due to: %s", err)
			}

			if err := vm.ctx.SharedMemory.Apply(map[ids.ID]*atomic.Requests{chainID: {PutRequests: atomicRequests.PutRequests}}, commitBatch); err != nil {
				t.Fatal(err)
			}

			sdb, err := vm.blockChain.State()
			if err != nil {
				t.Fatal(err)
			}
			err = exportTx.DELTAStateTransfer(vm.ctx, sdb)
			if err != nil {
				t.Fatal(err)
			}

			addr := GetEthAddress(testKeys[0])
			if sdb.GetBalance(addr).Cmp(new(big.Int).SetUint64(test.bal*units.Dione)) != 0 {
				t.Fatalf("address balance %s equal %s not %s", addr.String(), sdb.GetBalance(addr), new(big.Int).SetUint64(test.bal*units.Dione))
			}
		})
	}
}

func TestNewExportTxMulticoin(t *testing.T) {
	tests := []struct {
		name    string
		genesis string
		rules   params.Rules
		bal     uint64
		balmc   uint64
	}{
		{
			name:    "apricot phase 0",
			genesis: genesisJSONApricotPhase0,
			rules:   apricotRulesPhase0,
			bal:     49000000,
			balmc:   25000000,
		},
		{
			name:    "apricot phase 1",
			genesis: genesisJSONApricotPhase1,
			rules:   apricotRulesPhase1,
			bal:     49000000,
			balmc:   25000000,
		},
		{
			name:    "apricot phase 2",
			genesis: genesisJSONApricotPhase2,
			rules:   apricotRulesPhase2,
			bal:     48000000,
			balmc:   25000000,
		},
		{
			name:    "apricot phase 3",
			genesis: genesisJSONApricotPhase3,
			rules:   apricotRulesPhase3,
			bal:     48947900,
			balmc:   25000000,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issuer, vm, _, sharedMemory, _ := GenesisVM(t, true, test.genesis, "", "")

			defer func() {
				if err := vm.Shutdown(context.Background()); err != nil {
					t.Fatal(err)
				}
			}()

			parent := vm.LastAcceptedBlockInternal().(*Block)
			importAmount := uint64(50000000)
			utxoID := dione.UTXOID{TxID: ids.GenerateTestID()}

			utxo := &dione.UTXO{
				UTXOID: utxoID,
				Asset:  dione.Asset{ID: vm.ctx.DIONEAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: importAmount,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{testKeys[0].PublicKey().Address()},
					},
				},
			}
			utxoBytes, err := vm.codec.Marshal(codecVersion, utxo)
			if err != nil {
				t.Fatal(err)
			}

			inputID := utxo.InputID()

			tid := ids.GenerateTestID()
			importAmount2 := uint64(30000000)
			utxoID2 := dione.UTXOID{TxID: ids.GenerateTestID()}
			utxo2 := &dione.UTXO{
				UTXOID: utxoID2,
				Asset:  dione.Asset{ID: tid},
				Out: &secp256k1fx.TransferOutput{
					Amt: importAmount2,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{testKeys[0].PublicKey().Address()},
					},
				},
			}
			utxoBytes2, err := vm.codec.Marshal(codecVersion, utxo2)
			if err != nil {
				t.Fatal(err)
			}

			aChainSharedMemory := sharedMemory.NewSharedMemory(vm.ctx.AChainID)
			inputID2 := utxo2.InputID()
			if err := aChainSharedMemory.Apply(map[ids.ID]*atomic.Requests{vm.ctx.ChainID: {PutRequests: []*atomic.Element{
				{
					Key:   inputID[:],
					Value: utxoBytes,
					Traits: [][]byte{
						testKeys[0].PublicKey().Address().Bytes(),
					},
				},
				{
					Key:   inputID2[:],
					Value: utxoBytes2,
					Traits: [][]byte{
						testKeys[0].PublicKey().Address().Bytes(),
					},
				},
			}}}); err != nil {
				t.Fatal(err)
			}

			tx, err := vm.newImportTx(vm.ctx.AChainID, testEthAddrs[0], initialBaseFee, []*secp256k1.PrivateKey{testKeys[0]})
			if err != nil {
				t.Fatal(err)
			}

			if err := vm.issueTx(tx, false); err != nil {
				t.Fatal(err)
			}

			<-issuer

			blk, err := vm.BuildBlock(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			if err := blk.Verify(context.Background()); err != nil {
				t.Fatal(err)
			}

			if err := vm.SetPreference(context.Background(), blk.ID()); err != nil {
				t.Fatal(err)
			}

			if err := blk.Accept(context.Background()); err != nil {
				t.Fatal(err)
			}

			parent = vm.LastAcceptedBlockInternal().(*Block)
			exportAmount := uint64(5000000)

			testKeys0Addr := GetEthAddress(testKeys[0])
			exportId, err := ids.ToShortID(testKeys0Addr[:])
			if err != nil {
				t.Fatal(err)
			}

			tx, err = vm.newExportTx(tid, exportAmount, vm.ctx.AChainID, exportId, initialBaseFee, []*secp256k1.PrivateKey{testKeys[0]})
			if err != nil {
				t.Fatal(err)
			}

			exportTx := tx.UnsignedAtomicTx

			if err := exportTx.SemanticVerify(vm, tx, parent, parent.ethBlock.BaseFee(), test.rules); err != nil {
				t.Fatal("newExportTx created an invalid transaction", err)
			}

			commitBatch, err := vm.db.CommitBatch()
			if err != nil {
				t.Fatalf("Failed to create commit batch for VM due to %s", err)
			}
			chainID, atomicRequests, err := exportTx.AtomicOps()

			if err != nil {
				t.Fatalf("Failed to accept export transaction due to: %s", err)
			}

			if err := vm.ctx.SharedMemory.Apply(map[ids.ID]*atomic.Requests{chainID: {PutRequests: atomicRequests.PutRequests}}, commitBatch); err != nil {
				t.Fatal(err)
			}

			stdb, err := vm.blockChain.State()
			if err != nil {
				t.Fatal(err)
			}
			err = exportTx.DELTAStateTransfer(vm.ctx, stdb)
			if err != nil {
				t.Fatal(err)
			}

			addr := GetEthAddress(testKeys[0])
			if stdb.GetBalance(addr).Cmp(new(big.Int).SetUint64(test.bal*units.Dione)) != 0 {
				t.Fatalf("address balance %s equal %s not %s", addr.String(), stdb.GetBalance(addr), new(big.Int).SetUint64(test.bal*units.Dione))
			}
			if stdb.GetBalanceMultiCoin(addr, common.BytesToHash(tid[:])).Cmp(new(big.Int).SetUint64(test.balmc)) != 0 {
				t.Fatalf("address balance multicoin %s equal %s not %s", addr.String(), stdb.GetBalanceMultiCoin(addr, common.BytesToHash(tid[:])), new(big.Int).SetUint64(test.balmc))
			}
		})
	}
}
