// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package params

import (
	"math/big"

	"github.com/DioneProtocol/odysseygo/utils/units"
)

// Minimum Gas Price
const ( 
	// MinGasPrice is the number of nDIONE required per gas unit for a
	// transaction to be valid, measured in wei
	LaunchMinGasPrice        int64 = 470_000_000_000
	OdyPhase1MinGasPrice int64 = 225_000_000_000

	OdysseyAtomicTxFee = units.MilliDione

	OdyPhase1GasLimit uint64 = 8_000_000
	CortinaGasLimit       uint64 = 15_000_000

	OdyPhase3ExtraDataSize            uint64 = 80
	OdyPhase3MinBaseFee               int64  = 75_000_000_000
	OdyPhase3MaxBaseFee               int64  = 225_000_000_000
	OdyPhase3InitialBaseFee           int64  = 225_000_000_000
	OdyPhase3TargetGas                uint64 = 10_000_000
	OdyPhase4MinBaseFee               int64  = 25_000_000_000
	OdyPhase4MaxBaseFee               int64  = 1_000_000_000_000
	OdyPhase4BaseFeeChangeDenominator uint64 = 12
	OdyPhase5TargetGas                uint64 = 15_000_000
	OdyPhase5BaseFeeChangeDenominator uint64 = 36

	// The base cost to charge per atomic transaction. Added in Ody Phase 5.
	AtomicTxBaseCost uint64 = 500_000
)

// Constants for message sizes
const (
	MaxCodeHashesPerRequest = 5
)

var (
	// The atomic gas limit specifies the maximum amount of gas that can be consumed by the atomic
	// transactions included in a block and is enforced as of OdyPhase5. Prior to OdyPhase5,
	// a block included a single atomic transaction. As of OdyPhase5, each block can include a set
	// of atomic transactions where the cumulative atomic gas consumed is capped by the atomic gas limit,
	// similar to the block gas limit.
	//
	// This value must always remain <= MaxUint64.
	AtomicGasLimit *big.Int = big.NewInt(1_000_000)
)
