// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package params

import (
	"math/big"

	"github.com/DioneProtocol/odysseygo/utils/units"
	"github.com/ethereum/go-ethereum/common"
)

// Minimum Gas Price
const (
	// MinGasPrice is the number of nDIONE required per gas unit for a
	// transaction to be valid, measured in wei
	LaunchMinGasPrice        int64 = 2_380_952_380_952_38
	ApricotPhase1MinGasPrice int64 = 2_380_952_380_952_38

	OdysseyAtomicTxFee = 5 * units.Dione

	ApricotPhase1GasLimit uint64 = 8_000_000
	CortinaGasLimit       uint64 = 15_000_000

	ApricotPhase3ExtraDataSize            uint64 = 80
	ApricotPhase3MinBaseFee               int64  = 2_380_952_380_952_38
	ApricotPhase3MaxBaseFee               int64  = 7_142_857_142_857_14
	ApricotPhase3InitialBaseFee           int64  = 2_380_952_380_952_38
	ApricotPhase3TargetGas                uint64 = 10_000_000
	ApricotPhase4MinBaseFee               int64  = 2_380_952_380_952_38
	ApricotPhase4MaxBaseFee               int64  = 7_142_857_142_857_14
	ApricotPhase4BaseFeeChangeDenominator uint64 = 12
	ApricotPhase5TargetGas                uint64 = 15_000_000
	ApricotPhase5BaseFeeChangeDenominator uint64 = 36
	ApricotPhase7MinBaseFee               int64  = 23_809_523_809_524

	LpAddressDefault         string = "0x0000000000000000000000000000000000000001"
	GovernanceAddressDefault string = "0x0000000000000000000000000000000000000002"
	LpAddressMainnet         string = "0xD72C3d7957950197EcAa68d41E2E6803b61874E3"
	GovernanceAddressMannet  string = "0xD21A82BB789dCa05271711902F9E7FDDE158Bfd7"

	// The base cost to charge per atomic transaction. Added in Apricot Phase 5.
	AtomicTxBaseCost uint64 = 21_000
)

// Constants for message sizes
const (
	MaxCodeHashesPerRequest = 5
)

var (
	// The atomic gas limit specifies the maximum amount of gas that can be consumed by the atomic
	// transactions included in a block and is enforced as of ApricotPhase5. Prior to ApricotPhase5,
	// a block included a single atomic transaction. As of ApricotPhase5, each block can include a set
	// of atomic transactions where the cumulative atomic gas consumed is capped by the atomic gas limit,
	// similar to the block gas limit.
	//
	// This value must always remain <= MaxUint64.
	AtomicGasLimit *big.Int = big.NewInt(100_000)

	LpAllocation               *big.Int = big.NewInt(25_000)  // 25%
	GovernanceAllocation       *big.Int = big.NewInt(50_000)  // 50%
	OrionAllocation            *big.Int = big.NewInt(5_000)   // 5%
	MaxOrionAllocation         *big.Int = big.NewInt(25_000)  // 25%
	PriorityFeeOrionAllocation *big.Int = big.NewInt(5_000)   // 5%
	AllocationDenominator      *big.Int = big.NewInt(100_000) // 100%

	orionContractAddress         = common.HexToAddress("0x0710400000000000000000000000000000000000")
	orionLastUpdateTimestampSlot = common.HexToHash("0x0000000000000000000000000000000000000001")
	orionNodesSlot               = common.HexToHash("0x0000000000000000000000000000000000000002")

	OrionGetter = NewOrionGetter(orionContractAddress, orionLastUpdateTimestampSlot, orionNodesSlot)
)
