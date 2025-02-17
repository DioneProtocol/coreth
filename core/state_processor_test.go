// (c) 2019-2021, Ava Labs, Inc.
//
// This file is a derived work, based on the go-ethereum library whose original
// notices appear below.
//
// It is distributed under a license compatible with the licensing terms of the
// original code from which it is derived.
//
// Much love to the original authors for their work.
// **********
// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"math/big"
	"testing"

	"github.com/DioneProtocol/coreth/consensus"
	"github.com/DioneProtocol/coreth/consensus/dummy"
	"github.com/DioneProtocol/coreth/core/rawdb"
	"github.com/DioneProtocol/coreth/core/types"
	"github.com/DioneProtocol/coreth/core/vm"
	"github.com/DioneProtocol/coreth/params"
	"github.com/DioneProtocol/coreth/trie"
	"github.com/DioneProtocol/coreth/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

var (
	config     = params.TestChainConfig
	signer     = types.LatestSigner(config)
	testKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	testAddr   = common.HexToAddress("0x71562b71999873DB5b286dF957af199Ec94617F7")
)

func makeTx(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *types.Transaction {
	tx, _ := types.SignTx(types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, data), signer, testKey)
	return tx
}

func makeContractTx(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *types.Transaction {
	tx, _ := types.SignTx(types.NewContractCreation(nonce, amount, gasLimit, gasPrice, data), signer, testKey)
	return tx
}

func mkDynamicTx(nonce uint64, to common.Address, gasLimit uint64, gasTipCap, gasFeeCap *big.Int) *types.Transaction {
	tx, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     big.NewInt(0),
	}), signer, testKey)
	return tx
}

func mkDynamicCreationTx(nonce uint64, gasLimit uint64, gasTipCap, gasFeeCap *big.Int, data []byte) *types.Transaction {
	tx, _ := types.SignTx(types.NewTx(&types.DynamicFeeTx{
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		Value:     big.NewInt(0),
		Data:      data,
	}), signer, testKey)
	return tx
}

func u64(val uint64) *uint64 { return &val }

// TestStateProcessorErrors tests the output from the 'core' errors
// as defined in core/error.go. These errors are generated when the
// blockchain imports bad blocks, meaning blocks which have valid headers but
// contain invalid transactions
func TestStateProcessorErrors(t *testing.T) {
	{ // Tests against a 'recent' chain definition
		var (
			db             = rawdb.NewMemoryDatabase()
			initialBalance = new(big.Int).Mul(big.NewInt(params.Ether), big.NewInt(20000)) // 20000 ether
			gspec          = &Genesis{
				Config: config,
				Alloc: GenesisAlloc{
					common.HexToAddress("0x71562b71999873DB5b286dF957af199Ec94617F7"): GenesisAccount{
						Balance: initialBalance,
						Nonce:   0,
					},
				},
				GasLimit: params.CortinaGasLimit,
			}
			blockchain, _ = NewBlockChain(db, DefaultCacheConfig, gspec, dummy.NewFaker(), vm.Config{}, common.Hash{}, false)
		)
		defer blockchain.Stop()
		bigNumber := new(big.Int).SetBytes(common.FromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"))
		tooBigNumber := new(big.Int).Set(bigNumber)
		tooBigNumber.Add(tooBigNumber, common.Big1)
		for i, tt := range []struct {
			txs  []*types.Transaction
			want string
		}{
			{ // ErrNonceTooLow
				txs: []*types.Transaction{
					makeTx(0, common.Address{}, big.NewInt(0), params.TxGas, big.NewInt(238095238095238), nil),
					makeTx(0, common.Address{}, big.NewInt(0), params.TxGas, big.NewInt(238095238095238), nil),
				},
				want: "could not apply tx 1 [0xa5720245b9ef3b19135cd06f63bd0f2f13f89a370fa838513b99d0dcca9d1075]: nonce too low: address 0x71562b71999873DB5b286dF957af199Ec94617F7, tx: 0 state: 1",
			},
			{ // ErrNonceTooHigh
				txs: []*types.Transaction{
					makeTx(100, common.Address{}, big.NewInt(0), params.TxGas, big.NewInt(238095238095238), nil),
				},
				want: "could not apply tx 0 [0x5836a3edf86505a68bf4e0deafa673f5de3feb5bce10ced96c39c95d892242c9]: nonce too high: address 0x71562b71999873DB5b286dF957af199Ec94617F7, tx: 100 state: 0",
			},
			{ // ErrGasLimitReached
				txs: []*types.Transaction{
					makeTx(0, common.Address{}, big.NewInt(0), 15000001, big.NewInt(238095238095238), nil),
				},
				want: "could not apply tx 0 [0x6a344abc58d6f44eca47d2441abcab97fcc66e80445ba3677c822677ddf42dc2]: gas limit reached",
			},
			{ // ErrInsufficientFundsForTransfer
				txs: []*types.Transaction{
					makeTx(0, common.Address{}, new(big.Int).Mul(big.NewInt(4000000000000000000), big.NewInt(10000)), params.TxGas, big.NewInt(238095238095238), nil),
				},
				want: "could not apply tx 0 [0xe6d7c11ce139ec07e158ca6a0e8b5747caac889f1cfd7342517c80813070c928]: insufficient funds for gas * price + value: address 0x71562b71999873DB5b286dF957af199Ec94617F7 have 20000000000000000000000 want 40004999999999999998000",
			},
			{ // ErrInsufficientFunds
				txs: []*types.Transaction{
					makeTx(0, common.Address{}, big.NewInt(0), params.TxGas, new(big.Int).Mul(big.NewInt(900000000000000000), big.NewInt(10000)), nil),
				},
				want: "could not apply tx 0 [0xf0c3f9896e66b4fb44531723eb1b399ebc36026d00881c39fc6aeb30d78969ff]: insufficient funds for gas * price + value: address 0x71562b71999873DB5b286dF957af199Ec94617F7 have 20000000000000000000000 want 189000000000000000000000000",
			},
			// ErrGasUintOverflow
			// One missing 'core' error is ErrGasUintOverflow: "gas uint64 overflow",
			// In order to trigger that one, we'd have to allocate a _huge_ chunk of data, such that the
			// multiplication len(data) +gas_per_byte overflows uint64. Not testable at the moment
			{ // ErrIntrinsicGas
				txs: []*types.Transaction{
					makeTx(0, common.Address{}, big.NewInt(0), params.TxGas-1000, big.NewInt(238095238095238), nil),
				},
				want: "could not apply tx 0 [0xdae8df2e1ec45886f9dfce6a717e646d9d938b46ac86130a1f2baccc95bf13f3]: intrinsic gas too low: have 20000, want 21000",
			},
			{ // ErrGasLimitReached
				txs: []*types.Transaction{
					makeTx(0, common.Address{}, big.NewInt(0), params.TxGas*762, big.NewInt(238095238095238), nil),
				},
				want: "could not apply tx 0 [0x7c1b58947efb1685a0ad2248209ee112016bcc27646d5dbe1c96f3801e4f13af]: gas limit reached",
			},
			{ // ErrFeeCapTooLow
				txs: []*types.Transaction{
					mkDynamicTx(0, common.Address{}, params.TxGas, big.NewInt(0), big.NewInt(0)),
				},
				want: "could not apply tx 0 [0xc4ab868fef0c82ae0387b742aee87907f2d0fc528fc6ea0a021459fb0fc4a4a8]: max fee per gas less than block base fee: address 0x71562b71999873DB5b286dF957af199Ec94617F7, maxFeePerGas: 0 baseFee: 238095238095238",
			},
			{ // ErrTipVeryHigh
				txs: []*types.Transaction{
					mkDynamicTx(0, common.Address{}, params.TxGas, tooBigNumber, big.NewInt(1)),
				},
				want: "could not apply tx 0 [0x15b8391b9981f266b32f3ab7da564bbeb3d6c21628364ea9b32a21139f89f712]: max priority fee per gas higher than 2^256-1: address 0x71562b71999873DB5b286dF957af199Ec94617F7, maxPriorityFeePerGas bit length: 257",
			},
			{ // ErrFeeCapVeryHigh
				txs: []*types.Transaction{
					mkDynamicTx(0, common.Address{}, params.TxGas, big.NewInt(1), tooBigNumber),
				},
				want: "could not apply tx 0 [0x48bc299b83fdb345c57478f239e89814bb3063eb4e4b49f3b6057a69255c16bd]: max fee per gas higher than 2^256-1: address 0x71562b71999873DB5b286dF957af199Ec94617F7, maxFeePerGas bit length: 257",
			},
			{ // ErrTipAboveFeeCap
				txs: []*types.Transaction{
					mkDynamicTx(0, common.Address{}, params.TxGas, big.NewInt(2), big.NewInt(1)),
				},
				want: "could not apply tx 0 [0xf987a31ff0c71895780a7612f965a0c8b056deb54e020bb44fa478092f14c9b4]: max priority fee per gas higher than max fee per gas: address 0x71562b71999873DB5b286dF957af199Ec94617F7, maxPriorityFeePerGas: 2, maxFeePerGas: 1",
			},
			{ // ErrInsufficientFunds
				// This test is designed to have the effective cost be covered by the balance, but
				// the extended requirement on FeeCap*gas < balance to fail
				txs: []*types.Transaction{
					mkDynamicTx(0, common.Address{}, params.TxGas, big.NewInt(1), new(big.Int).Mul(big.NewInt(200000000000000), big.NewInt(10000))),
				},
				want: "could not apply tx 0 [0x3d5917e97d1ff683dad229dedd792043e7a39782ac12597408167893fa539fd3]: insufficient funds for gas * price + value: address 0x71562b71999873DB5b286dF957af199Ec94617F7 have 20000000000000000000000 want 42000000000000000000000",
			},
			{ // Another ErrInsufficientFunds, this one to ensure that feecap/tip of max u256 is allowed
				txs: []*types.Transaction{
					mkDynamicTx(0, common.Address{}, params.TxGas, bigNumber, bigNumber),
				},
				want: "could not apply tx 0 [0xd82a0c2519acfeac9a948258c47e784acd20651d9d80f9a1c67b4137651c3a24]: insufficient funds for gas * price + value: address 0x71562b71999873DB5b286dF957af199Ec94617F7 have 20000000000000000000000 want 2431633873983640103894990685182446064918669677978451844828609264166175722438635000",
			},
		} {
			block := GenerateBadBlock(gspec.ToBlock(), dummy.NewFaker(), tt.txs, gspec.Config)
			_, err := blockchain.InsertChain(types.Blocks{block})
			if err == nil {
				t.Fatal("block imported without errors")
			}
			if have, want := err.Error(), tt.want; have != want {
				t.Errorf("test %d:\nhave \"%v\"\nwant \"%v\"\n", i, have, want)
			}
		}
	}

	// ErrTxTypeNotSupported, For this, we need an older chain
	{
		var (
			db    = rawdb.NewMemoryDatabase()
			gspec = &Genesis{
				Config: &params.ChainConfig{
					ChainID:                     big.NewInt(1),
					HomesteadBlock:              big.NewInt(0),
					EIP150Block:                 big.NewInt(0),
					EIP155Block:                 big.NewInt(0),
					EIP158Block:                 big.NewInt(0),
					ByzantiumBlock:              big.NewInt(0),
					ConstantinopleBlock:         big.NewInt(0),
					PetersburgBlock:             big.NewInt(0),
					IstanbulBlock:               big.NewInt(0),
					MuirGlacierBlock:            big.NewInt(0),
					ApricotPhase1BlockTimestamp: utils.NewUint64(0),
					ApricotPhase2BlockTimestamp: utils.NewUint64(0),
				},
				Alloc: GenesisAlloc{
					common.HexToAddress("0x71562b71999873DB5b286dF957af199Ec94617F7"): GenesisAccount{
						Balance: big.NewInt(1000000000000000000), // 1 ether
						Nonce:   0,
					},
				},
				GasLimit: params.ApricotPhase1GasLimit,
			}
			blockchain, _ = NewBlockChain(db, DefaultCacheConfig, gspec, dummy.NewFaker(), vm.Config{}, common.Hash{}, false)
		)
		defer blockchain.Stop()
		for i, tt := range []struct {
			txs  []*types.Transaction
			want string
		}{
			{ // ErrTxTypeNotSupported
				txs: []*types.Transaction{
					mkDynamicTx(0, common.Address{}, params.TxGas-1000, big.NewInt(0), big.NewInt(0)),
				},
				want: "could not apply tx 0 [0x88626ac0d53cb65308f2416103c62bb1f18b805573d4f96a3640bbbfff13c14f]: transaction type not supported",
			},
		} {
			block := GenerateBadBlock(gspec.ToBlock(), dummy.NewFaker(), tt.txs, gspec.Config)
			_, err := blockchain.InsertChain(types.Blocks{block})
			if err == nil {
				t.Fatal("block imported without errors")
			}
			if have, want := err.Error(), tt.want; have != want {
				t.Errorf("test %d:\nhave \"%v\"\nwant \"%v\"\n", i, have, want)
			}
		}
	}

	// ErrSenderNoEOA, for this we need the sender to have contract code
	{
		var (
			db    = rawdb.NewMemoryDatabase()
			gspec = &Genesis{
				Config: config,
				Alloc: GenesisAlloc{
					common.HexToAddress("0x71562b71999873DB5b286dF957af199Ec94617F7"): GenesisAccount{
						Balance: big.NewInt(1000000000000000000), // 1 ether
						Nonce:   0,
						Code:    common.FromHex("0xB0B0FACE"),
					},
				},
				GasLimit: params.CortinaGasLimit,
			}
			blockchain, _ = NewBlockChain(db, DefaultCacheConfig, gspec, dummy.NewFaker(), vm.Config{}, common.Hash{}, false)
		)
		defer blockchain.Stop()
		for i, tt := range []struct {
			txs  []*types.Transaction
			want string
		}{
			{ // ErrSenderNoEOA
				txs: []*types.Transaction{
					mkDynamicTx(0, common.Address{}, params.TxGas-1000, big.NewInt(0), big.NewInt(0)),
				},
				want: "could not apply tx 0 [0x88626ac0d53cb65308f2416103c62bb1f18b805573d4f96a3640bbbfff13c14f]: sender not an eoa: address 0x71562b71999873DB5b286dF957af199Ec94617F7, codehash: 0x9280914443471259d4570a8661015ae4a5b80186dbc619658fb494bebc3da3d1",
			},
		} {
			block := GenerateBadBlock(gspec.ToBlock(), dummy.NewFaker(), tt.txs, gspec.Config)
			_, err := blockchain.InsertChain(types.Blocks{block})
			if err == nil {
				t.Fatal("block imported without errors")
			}
			if have, want := err.Error(), tt.want; have != want {
				t.Errorf("test %d:\nhave \"%v\"\nwant \"%v\"\n", i, have, want)
			}
		}
	}

	// ErrMaxInitCodeSizeExceeded, for this we need extra Shanghai (DUpgrade/EIP-3860) enabled.
	{
		var (
			db             = rawdb.NewMemoryDatabase()
			initialBalance = new(big.Int).Mul(big.NewInt(params.Ether), big.NewInt(1000)) // 1000 ether
			gspec          = &Genesis{
				Config: &params.ChainConfig{
					ChainID:                         big.NewInt(1),
					HomesteadBlock:                  big.NewInt(0),
					DAOForkBlock:                    big.NewInt(0),
					DAOForkSupport:                  true,
					EIP150Block:                     big.NewInt(0),
					EIP155Block:                     big.NewInt(0),
					EIP158Block:                     big.NewInt(0),
					ByzantiumBlock:                  big.NewInt(0),
					ConstantinopleBlock:             big.NewInt(0),
					PetersburgBlock:                 big.NewInt(0),
					IstanbulBlock:                   big.NewInt(0),
					MuirGlacierBlock:                big.NewInt(0),
					ApricotPhase1BlockTimestamp:     utils.NewUint64(0),
					ApricotPhase2BlockTimestamp:     utils.NewUint64(0),
					ApricotPhase3BlockTimestamp:     utils.NewUint64(0),
					ApricotPhase4BlockTimestamp:     utils.NewUint64(0),
					ApricotPhase5BlockTimestamp:     utils.NewUint64(0),
					ApricotPhasePre6BlockTimestamp:  utils.NewUint64(0),
					ApricotPhase6BlockTimestamp:     utils.NewUint64(0),
					ApricotPhasePost6BlockTimestamp: utils.NewUint64(0),
					BanffBlockTimestamp:             utils.NewUint64(0),
					CortinaBlockTimestamp:           utils.NewUint64(0),
					DUpgradeBlockTimestamp:          utils.NewUint64(0),
				},
				Alloc: GenesisAlloc{
					common.HexToAddress("0x71562b71999873DB5b286dF957af199Ec94617F7"): GenesisAccount{
						Balance: initialBalance, // 1 ether
						Nonce:   0,
					},
				},
				GasLimit: params.CortinaGasLimit,
			}
			blockchain, _  = NewBlockChain(db, DefaultCacheConfig, gspec, dummy.NewFaker(), vm.Config{}, common.Hash{}, false)
			tooBigInitCode = [params.MaxInitCodeSize + 1]byte{}
			smallInitCode  = [320]byte{}
		)
		defer blockchain.Stop()
		for i, tt := range []struct {
			txs  []*types.Transaction
			want string
		}{
			{ // ErrMaxInitCodeSizeExceeded
				txs: []*types.Transaction{
					mkDynamicCreationTx(0, 500000, common.Big0, big.NewInt(params.ApricotPhase3InitialBaseFee), tooBigInitCode[:]),
				},
				want: "could not apply tx 0 [0xe09166942d9f37333c3b01b5060cd4e332f997226ab3fed0116248bef2f3eeb8]: max initcode size exceeded: code size 49153 limit 49152",
			},
			{ // ErrIntrinsicGas: Not enough gas to cover init code
				txs: []*types.Transaction{
					mkDynamicCreationTx(0, 54299, common.Big0, big.NewInt(params.ApricotPhase3InitialBaseFee), smallInitCode[:]),
				},
				want: "could not apply tx 0 [0x2f36e07010df04a5dba603efadb4ae7914c4689295dac18b2ef71ea935fe9dbf]: intrinsic gas too low: have 54299, want 54300",
			},
		} {
			block := GenerateBadBlock(gspec.ToBlock(), dummy.NewFaker(), tt.txs, gspec.Config)
			_, err := blockchain.InsertChain(types.Blocks{block})
			if err == nil {
				t.Fatal("block imported without errors")
			}
			if have, want := err.Error(), tt.want; have != want {
				t.Errorf("test %d:\nhave \"%v\"\nwant \"%v\"\n", i, have, want)
			}
		}
	}
}

// GenerateBadBlock constructs a "block" which contains the transactions. The transactions are not expected to be
// valid, and no proper post-state can be made. But from the perspective of the blockchain, the block is sufficiently
// valid to be considered for import:
// - valid pow (fake), ancestry, difficulty, gaslimit etc
func GenerateBadBlock(parent *types.Block, engine consensus.Engine, txs types.Transactions, config *params.ChainConfig) *types.Block {
	header := &types.Header{
		ParentHash: parent.Hash(),
		Coinbase:   parent.Coinbase(),
		Difficulty: engine.CalcDifficulty(&fakeChainReader{config}, parent.Time()+10, &types.Header{
			Number:     parent.Number(),
			Time:       parent.Time(),
			Difficulty: parent.Difficulty(),
			UncleHash:  parent.UncleHash(),
		}),
		GasLimit:  parent.GasLimit(),
		Number:    new(big.Int).Add(parent.Number(), common.Big1),
		Time:      parent.Time() + 10,
		UncleHash: types.EmptyUncleHash,
	}
	if config.IsApricotPhase3(header.Time) {
		header.Extra, header.BaseFee, _ = dummy.CalcBaseFee(config, parent.Header(), header.Time)
	}
	if config.IsApricotPhase4(header.Time) {
		header.BlockGasCost = big.NewInt(0)
		header.ExtDataGasUsed = big.NewInt(0)
	}
	var receipts []*types.Receipt
	// The post-state result doesn't need to be correct (this is a bad block), but we do need something there
	// Preferably something unique. So let's use a combo of blocknum + txhash
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(header.Number.Bytes())
	var cumulativeGas uint64
	for _, tx := range txs {
		txh := tx.Hash()
		hasher.Write(txh[:])
		receipt := types.NewReceipt(nil, false, cumulativeGas+tx.Gas())
		receipt.TxHash = tx.Hash()
		receipt.GasUsed = tx.Gas()
		receipts = append(receipts, receipt)
		cumulativeGas += tx.Gas()
	}
	header.Root = common.BytesToHash(hasher.Sum(nil))
	// Assemble and return the final block for sealing
	return types.NewBlock(header, txs, nil, receipts, trie.NewStackTrie(nil), nil, true)
}
