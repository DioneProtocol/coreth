package delta

import (
	"context"
	"math/big"
	"testing"
	"time"

	engCommon "github.com/DioneProtocol/odysseygo/snow/engine/common"

	"github.com/DioneProtocol/coreth/consensus/dummy"
	"github.com/DioneProtocol/coreth/core"
	"github.com/DioneProtocol/coreth/core/types"
	"github.com/DioneProtocol/coreth/params"
	"github.com/DioneProtocol/odysseygo/ids"
	"github.com/DioneProtocol/odysseygo/snow/choices"
	"github.com/DioneProtocol/odysseygo/utils/crypto/secp256k1"
	"github.com/DioneProtocol/odysseygo/utils/units"
	"github.com/DioneProtocol/odysseygo/vms/components/chain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ params.OrionNodesGetter = &testOrionGetter{}

type testOrionGetter struct{}

func (t *testOrionGetter) GetLastUpdateTimestamp(params.StateGetter) uint64 {
	return 1
}

func (t *testOrionGetter) GetNodesList(params.StateGetter) []ids.NodeID {
	return []ids.NodeID{ids.NodeID{}}
}

func (t *testOrionGetter) NodesAmount(s params.StateGetter) *big.Int {
	return new(big.Int).SetInt64(int64(len(t.GetNodesList(s))))
}

func setupVM(t *testing.T) (chan engCommon.Message, *VM) {
	importAmount := 5000 * units.Dione
	issuer, vm, _, _, _ := GenesisVMWithUTXOs(t, true, genesisJSONCancun, "{\"pruning-enabled\":true}", "", map[ids.ShortID]uint64{
		testShortIDAddrs[0]: importAmount,
	})

	newTxPoolHeadChan := make(chan core.NewTxPoolReorgEvent, 1)
	vm.txPool.SubscribeNewReorgEvent(newTxPoolHeadChan)

	importTx, err := vm.newImportTx(vm.ctx.AChainID, testEthAddrs[0], initialBaseFee, []*secp256k1.PrivateKey{testKeys[0]})
	if err != nil {
		t.Fatal(err)
	}

	if err := vm.mempool.AddLocalTx(importTx); err != nil {
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

	if status := blk.Status(); status != choices.Processing {
		t.Fatalf("Expected status of built block to be %s, but found %s", choices.Processing, status)
	}

	if err := vm.SetPreference(context.Background(), blk.ID()); err != nil {
		t.Fatal(err)
	}

	if err := blk.Accept(context.Background()); err != nil {
		t.Fatal(err)
	}

	newHead := <-newTxPoolHeadChan
	if newHead.Head.Hash() != common.Hash(blk.ID()) {
		t.Fatalf("Expected new block to match")
	}

	time.Sleep(time.Second * time.Duration(dummy.ApricotPhase4TargetBlockRate))

	return issuer, vm
}

func executeTx(t *testing.T, vm *VM, issuer chan engCommon.Message, tx *types.Transaction) *types.Block {
	errs := vm.txPool.AddRemotesSync([]*types.Transaction{tx})
	for i, err := range errs {
		if err != nil {
			t.Fatalf("Failed to add tx at index %d: %s", i, err)
		}
	}

	<-issuer

	blk, err := vm.BuildBlock(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if err := blk.Verify(context.Background()); err != nil {
		t.Fatal(err)
	}

	if status := blk.Status(); status != choices.Processing {
		t.Fatalf("Expected status of built block to be %s, but found %s", choices.Processing, status)
	}

	if err := blk.Accept(context.Background()); err != nil {
		t.Fatal(err)
	}

	if status := blk.Status(); status != choices.Accepted {
		t.Fatalf("Expected status of accepted block to be %s, but found %s", choices.Accepted, status)
	}

	lastAcceptedID, err := vm.LastAccepted(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if lastAcceptedID != blk.ID() {
		t.Fatalf("Expected last accepted blockID to be the accepted block: %s, but found %s", blk.ID(), lastAcceptedID)
	}

	ethBlock := blk.(*chain.BlockWrapper).Block.(*Block).ethBlock
	return ethBlock
}

func TestDioneFees(t *testing.T) {
	issuer, vm := setupVM(t)

	defer func() {
		if err := vm.Shutdown(context.Background()); err != nil {
			t.Fatal(err)
		}
	}()

	testGetter := &testOrionGetter{}
	params.OrionGetter = testGetter
	bcState, err := vm.blockChain.State()
	require.NoError(t, err)

	rules := vm.currentRules()
	require.True(t, rules.IsCancun)

	lpBalanceBefore := bcState.GetBalance(rules.LpAddress)
	governanceBalanceBefore := bcState.GetBalance(rules.GovernanceAddress)

	baseFee := params.ApricotPhase4MinBaseFee
	gasPrice := new(big.Int).Mul(big.NewInt(baseFee), big.NewInt(3))
	tx := types.NewTransaction(0, testEthAddrs[0], common.Big0, 21000, gasPrice, nil)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(vm.chainID), testKeys[0].ToECDSA())
	if err != nil {
		t.Fatal(err)
	}

	ethBlock := executeTx(t, vm, issuer, signedTx)
	receipts := vm.blockChain.GetReceiptsByHash(ethBlock.Hash())

	require.Equal(t, len(receipts), 1)
	receipt := receipts[0]

	gasUsed := new(big.Int).SetUint64(receipt.GasUsed)
	gasPrice = receipt.EffectiveGasPrice
	txFee := new(big.Int).Mul(gasUsed, gasPrice)

	bcState, err = vm.blockChain.State()
	require.NoError(t, err)

	expectedBaseFee := big.NewInt(4999999999999998000) // ~5 Dione
	priorityFee := new(big.Int).Sub(txFee, expectedBaseFee)

	lpBalance := bcState.GetBalance(rules.LpAddress)
	lpBalanceDiff := new(big.Int).Sub(lpBalance, lpBalanceBefore)
	expectedLpBalance := new(big.Int).Div(new(big.Int).Mul(expectedBaseFee, rules.LpAllocation), rules.AllocationDenominator)
	assert.Equal(t, expectedLpBalance, lpBalanceDiff)

	nodesAmount := testGetter.NodesAmount(bcState)
	governanceBalance := bcState.GetBalance(rules.GovernanceAddress)
	governanceBalanceDiff := new(big.Int).Sub(governanceBalance, governanceBalanceBefore)
	governanceAllocation := new(big.Int).Sub(rules.GovernanceAllocation, new(big.Int).Mul(nodesAmount, rules.OrionAllocation))
	expectedGovernanceBalance := new(big.Int).Div(new(big.Int).Mul(expectedBaseFee, governanceAllocation), rules.AllocationDenominator)
	assert.Equal(t, expectedGovernanceBalance, governanceBalanceDiff)

	orionNodeFee := ethBlock.OrionNodeFee()
	orionNodePriorityFee := new(big.Int).Div(new(big.Int).Mul(priorityFee, rules.PriorityFeeOrionAllocation), rules.AllocationDenominator)
	orionNodeGovernance := new(big.Int).Div(new(big.Int).Mul(expectedBaseFee, new(big.Int).Mul(nodesAmount, rules.OrionAllocation)), rules.AllocationDenominator)
	expectedOrionFee := new(big.Int).Add(orionNodePriorityFee, orionNodeGovernance)

	assert.Equal(t, expectedOrionFee, orionNodeFee)
}
