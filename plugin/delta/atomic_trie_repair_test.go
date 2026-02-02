// (c) 2020-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package delta

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/DioneProtocol/coreth/core/types"
	"github.com/DioneProtocol/odysseygo/database/memdb"
	"github.com/DioneProtocol/odysseygo/database/versiondb"
	"github.com/DioneProtocol/odysseygo/ids"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

var (
	repairTestHeights = make(map[uint64]ids.ID)

	//go:embed bonus_blocks_repair_test.json
	repairTestBonusBlocksJson []byte

	repairTestBlocksParsed map[uint64]*types.Block = make(map[uint64]*types.Block)
)

func init() {
	repairTestBonusBlocks := map[uint64]string{
		102972: "Njm9TcLUXRojZk8YhEM6ksvfiPdC1TME4zJvGaDXgzMCyB6oB",
		103105: "BYqLB6xpqy7HsAgP2XNfGE8Ubg1uEzse5mBPTSJH9z5s8pvMa",
		103143: "AfWvJH3rB2fdHuPWQp6qYNCFVT29MooQPRigD88rKKwUDEDhq",
		103183: "2KPW9G5tiNF14tZNfG4SqHuQrtUYVZyxuof37aZ7AnTKrQdsHn",
		103197: "pE93VXY3N5QKfwsEFcM9i59UpPFgeZ8nxpJNaGaDQyDgsscNf",
		103203: "2czmtnBS44VCWNRFUM89h4Fe9m3ZeZVYyh7Pe3FhNqjRNgPXhZ",
		103208: "esx5J962LtYm2aSrskpLai5e4CMMsaS1dsu9iuLGJ3KWgSu2M",
		103209: "DK9NqAJGry1wAo767uuYc1dYXAjUhzwka6vi8d9tNheqzGUTd",
		103259: "i1HoerJ1axognkUKKL58FvF9aLrbZKtv7TdKLkT5kgzoeU1vB",
		103261: "2DpCuBaH94zKKFNY2XTs4GeJcwsEv6qT2DHc59S8tdg97GZpcJ",
		103266: "2ez4CA7w4HHr8SSobHQUAwFgj2giRNjNFUZK9JvrZFa1AuRj6X",
		103287: "2QBNMMFJmhVHaGF45GAPszKyj1gK6ToBERRxYvXtM7yfrdUGPK",
		103339: "2pSjfo7rkFCfZ2CqAxqfw8vqM2CU2nVLHrFZe3rwxz43gkVuGo",
		103346: "2SiSziHHqPjb1qkw7CdGYupokiYpd2b7mMqRiyszurctcA5AKr",
		103350: "2F5tSQbdTfhZxvkxZqdFp7KR3FrJPKEsDLQK7KtPhNXj1EZAh4",
		103358: "2tCe88ur6MLQcVgwE5XxoaHiTGtSrthwKN3SdbHE4kWiQ7MSTV",
		103437: "21o2fVTnzzmtgXqkV1yuQeze7YEQhR5JB31jVVD9oVUnaaV8qm",
		103472: "2nG4exd9eUoAGzELfksmBR8XDCKhohY1uDKRFzEXJG4M8p3qA7",
		103478: "63YLdYXfXc5tY3mwWLaDsbXzQHYmwWVxMP7HKbRh4Du3C2iM1",
		103493: "soPweZ8DGaoUMjrnzjH3V2bypa7ZvvfqBan4UCsMUxMP759gw",
		103514: "2dNkpQF4mooveyUDfBYQTBfsGDV4wkncQPpEw4kHKfSTSTo5x",
		103536: "PJTkRrHvKZ1m4AQdPND1MBpUXpCrGN4DDmXmJQAiUrsxPoLQX",
		103545: "22ck2Z7cC38hmBfX2v3jMWxun8eD8psNaicfYeokS67DxwmPTx",
		103547: "pTf7gfk1ksj7bqMrLyMCij8FBKth1uRqQrtfykMFeXhx5xnrL",
		103554: "9oZh4qyBCcVwSGyDoUzRAuausvPJN3xH6nopKS6bwYzMfLoQ2",
		103555: "MjExz2z1qhwugc1tAyiGxRsCq4GvJwKfyyS29nr4tRVB8ooic",
		103559: "cwJusfmn98TW3DjAbfLRN9utYR24KAQ82qpAXmVSvjHyJZuM2",
		103561: "2YgxGHns7Z2hMMHJsPCgVXuJaL7x1b3gnHbmSCfCdyAcYGr6mx",
		103563: "2AXxT3PSEnaYHNtBTnYrVTf24TtKDWjky9sqoFEhydrGXE9iKH",
		103564: "Ry2sfjFfGEnJxRkUGFSyZNn7GR3m4aKAf1scDW2uXSNQB568Y",
		103569: "21Jys8UNURmtckKSV89S2hntEWymJszrLQbdLaNcbXcxDAsQSa",
		103570: "sg6wAwFBsPQiS5Yfyh41cVkCRQbrrXsxXmeNyQ1xkunf2sdyv",
		103575: "z3BgePPpCXq1mRBRvUi28rYYxnEtJizkUEHnDBrcZeVA7MFVk",
		103577: "uK5Ff9iBfDtREpVv9NgCQ1STD1nzLJG3yrfibHG4mGvmybw6f",
		103578: "Qv5v5Ru8ArfnWKB1w6s4G5EYPh7TybHJtF6UsVwAkfvZFoqmj",
		103582: "7KCZKBpxovtX9opb7rMRie9WmW5YbZ8A4HwBBokJ9eSHpZPqx",
		103587: "2AfTQ2FXNj9bkSUQnud9pFXULx6EbF7cbbw6i3ayvc2QNhgxfF",
		103590: "2gTygYckZgFZfN5QQWPaPBD3nabqjidV55mwy1x1Nd4JmJAwaM",
		103591: "2cUPPHy1hspr2nAKpQrrAEisLKkaWSS9iF2wjNFyFRs8vnSkKK",
		103594: "5MptSdP6dBMPSwk9GJjeVe39deZJTRh9i82cgNibjeDffrrTf",
		103597: "2J8z7HNv4nwh82wqRGyEHqQeuw4wJ6mCDCSvUgusBu35asnshK",
		103598: "2i2FP6nJyvhX9FR15qN2D9AVoK5XKgBD2i2AQ7FoSpfowxvQDX",
		103603: "2v3smb35s4GLACsK4Zkd2RcLBLdWA4huqrvq8Y3VP4CVe8kfTM",
		103604: "b7XfDDLgwB12DfL7UTWZoxwBpkLPL5mdHtXngD94Y2RoeWXSh",
		103607: "PgaRk1UAoUvRybhnXsrLq5t6imWhEa6ksNjbN6hWgs4qPrSzm",
		103612: "2oueNTj4dUE2FFtGyPpawnmCCsy6EUQeVHVLZy8NHeQmkAciP4",
		103614: "2YHZ1KymFjiBhpXzgt6HXJhLSt5SV9UQ4tJuUNjfN1nQQdm5zz",
		103617: "amgH2C1s9H3Av7vSW4y7n7TXb9tKyKHENvrDXutgNN6nsejgc",
		103618: "fV8k1U8oQDmfVwK66kAwN73aSsWiWhm8quNpVnKmSznBycV2W",
		103621: "Nzs93kFTvcXanFUp9Y8VQkKYnzmH8xykxVNFJTkdyAEeuxWbP",
		103623: "2rAsBj3emqQa13CV8r5fTtHogs4sXnjvbbXVzcKPi3WmzhpK9D",
		103624: "2JbuExUGKW5mYz5KfXATwq1ibRDimgks9wEdYGNSC6Ttey1R4U",
		103627: "tLLijh7oKfvWT1yk9zRv4FQvuQ5DAiuvb5kHCNN9zh4mqkFMG",
		103628: "dWBsRYRwFrcyi3DPdLoHsL67QkZ5h86hwtVfP94ZBaY18EkmF",
		103629: "XMoEsew2DhSgQaydcJFJUQAQYP8BTNTYbEJZvtbrV2QsX7iE3",
		103630: "2db2wMbVAoCc5EUJrsBYWvNZDekqyY8uNpaaVapdBAQZ5oRaou",
		103633: "2QiHZwLhQ3xLuyyfcdo5yCUfoSqWDvRZox5ECU19HiswfroCGp",
	}

	for height, blkIDStr := range repairTestBonusBlocks {
		blkID, err := ids.FromString(blkIDStr)
		if err != nil {
			panic(err)
		}
		repairTestHeights[height] = blkID
	}

	var rlpMap map[uint64]string
	err := json.Unmarshal(repairTestBonusBlocksJson, &rlpMap)
	if err != nil {
		panic(err)
	}
	for height, rlpHex := range rlpMap {
		expectedHash, ok := repairTestHeights[height]
		if !ok {
			panic(fmt.Sprintf("missing bonus block at height %d", height))
		}
		var ethBlock types.Block
		if err := rlp.DecodeBytes(common.Hex2Bytes(rlpHex), &ethBlock); err != nil {
			panic(fmt.Sprintf("failed to decode bonus block at height %d: %s", height, err))
		}
		if ids.ID(ethBlock.Hash()) != expectedHash {
			panic(fmt.Sprintf("block ID mismatch at (%s != %s)", ids.ID(ethBlock.Hash()), expectedHash))
		}

		repairTestBlocksParsed[height] = &ethBlock
	}
	if len(repairTestBlocksParsed) != len(repairTestHeights) {
		panic("mismatched bonus block heights")
	}
}

type atomicTrieRepairTest struct {
	setup                   func(a *atomicTrie, db *versiondb.Database)
	expectedHeightsRepaired int
}

func TestAtomicTrieRepair(t *testing.T) {
	require := require.New(t)
	for name, test := range map[string]atomicTrieRepairTest{
		"needs repair": {
			setup:                   func(a *atomicTrie, db *versiondb.Database) {},
			expectedHeightsRepaired: len(repairTestBlocksParsed),
		},
		"should not be repaired twice": {
			setup: func(a *atomicTrie, db *versiondb.Database) {
				_, err := a.repairAtomicTrie(repairTestHeights, repairTestBlocksParsed)
				require.NoError(err)
				require.NoError(db.Commit())
			},
			expectedHeightsRepaired: 0,
		},
		"did not need repair": {
			setup: func(a *atomicTrie, db *versiondb.Database) {
				// simulates a node that has the bonus blocks in the atomic trie
				// but has not yet run the repair
				_, err := a.repairAtomicTrie(repairTestHeights, repairTestBlocksParsed)
				require.NoError(err)
				require.NoError(a.metadataDB.Delete(repairedKey))
				require.NoError(db.Commit())
			},
			expectedHeightsRepaired: len(repairTestBlocksParsed),
		},
	} {
		t.Run(name, test.test)
	}
}

func (test atomicTrieRepairTest) test(t *testing.T) {
	require := require.New(t)
	commitInterval := uint64(4096)

	// create an unrepaired atomic trie for setup
	db := versiondb.New(memdb.New())
	repo, err := NewAtomicTxRepository(db, Codec, 0, nil)
	require.NoError(err)
	atomicBackend, err := NewAtomicBackend(db, testSharedMemory(), nil, repo, 0, common.Hash{}, commitInterval)
	require.NoError(err)
	a := atomicBackend.AtomicTrie().(*atomicTrie)

	// make a commit at a height larger than all bonus blocks
	maxBonusBlockHeight := slices.Max(maps.Keys(repairTestBlocksParsed))
	commitHeight := nearestCommitHeight(maxBonusBlockHeight, commitInterval) + commitInterval
	err = a.commit(commitHeight, types.EmptyRootHash)
	require.NoError(err)
	require.NoError(db.Commit())

	// perform additional setup
	test.setup(a, db)

	// recreate the trie with the repair constructor to test the repair
	var heightsRepaired int
	atomicBackend, heightsRepaired, err = NewAtomicBackendWithBonusBlockRepair(
		db, testSharedMemory(), repairTestHeights, repairTestBlocksParsed,
		repo, commitHeight, common.Hash{}, commitInterval,
	)
	require.NoError(err)
	require.Equal(test.expectedHeightsRepaired, heightsRepaired)

	// call Abort to make sure the repair has called Commit
	db.Abort()
	// verify the trie is repaired
	verifyAtomicTrieIsAlreadyRepaired(require, db, repo, commitHeight, commitInterval)
}

func verifyAtomicTrieIsAlreadyRepaired(
	require *require.Assertions, db *versiondb.Database, repo *atomicTxRepository,
	commitHeight uint64, commitInterval uint64,
) {
	// create a map to track the expected items in the atomic trie.
	// note we serialize the atomic ops to bytes so we can compare nil
	// and empty slices as equal
	expectedKeys := 0
	expected := make(map[uint64]map[ids.ID][]byte)
	for height, block := range repairTestBlocksParsed {
		txs, err := ExtractAtomicTxs(block.ExtData(), false, Codec)
		require.NoError(err)

		requests := make(map[ids.ID][]byte)
		ops, err := mergeAtomicOps(txs)
		require.NoError(err)
		for id, op := range ops {
			bytes, err := Codec.Marshal(codecVersion, op)
			require.NoError(err)
			requests[id] = bytes
			expectedKeys++
		}
		expected[height] = requests
	}

	atomicBackend, heightsRepaired, err := NewAtomicBackendWithBonusBlockRepair(
		db, testSharedMemory(), repairTestHeights, repairTestBlocksParsed,
		repo, commitHeight, common.Hash{}, commitInterval,
	)
	require.NoError(err)
	a := atomicBackend.AtomicTrie().(*atomicTrie)
	require.NoError(err)
	require.Zero(heightsRepaired) // migration should not run a second time

	// iterate over the trie and check it contains the expected items
	root, err := a.Root(commitHeight)
	require.NoError(err)
	it, err := a.Iterator(root, nil)
	require.NoError(err)

	foundKeys := 0
	for it.Next() {
		bytes, err := a.codec.Marshal(codecVersion, it.AtomicOps())
		require.NoError(err)
		require.Equal(expected[it.BlockNumber()][it.BlockchainID()], bytes)
		foundKeys++
	}
	require.Equal(expectedKeys, foundKeys)
}
