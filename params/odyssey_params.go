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
	OdysseyPhase1MinGasPrice int64 = 225_000_000_000

	OdysseyAtomicTxFee = 50 * units.MilliDione

	OdysseyPhase1GasLimit uint64 = 8_000_000
	CortinaGasLimit       uint64 = 15_000_000

	OdysseyPhase1ExtraDataSize            uint64 = 80
	OdysseyPhase1InitialBaseFee           int64  = 225_000_000_000
	OdysseyPhase1MinBaseFee               int64  = 25_000_000_000
	OdysseyPhase1MaxBaseFee               int64  = 1_000_000_000_000
	OdysseyPhase1TargetGas                uint64 = 15_000_000
	OdysseyPhase1BaseFeeChangeDenominator uint64 = 36

	// The base cost to charge per atomic transaction. Added in Odyssey Phase 1.
	AtomicTxBaseCost uint64 = 10_000
)

// Constants for message sizes
const (
	MaxCodeHashesPerRequest = 5
)

var (
	// The atomic gas limit specifies the maximum amount of gas that can be consumed by the atomic
	// transactions included in a block and is enforced as of OdysseyPhase1. Prior to OdysseyPhase1,
	// a block included a single atomic transaction. As of OdysseyPhase1, each block can include a set
	// of atomic transactions where the cumulative atomic gas consumed is capped by the atomic gas limit,
	// similar to the block gas limit.
	//
	// This value must always remain <= MaxUint64.
	AtomicGasLimit *big.Int = big.NewInt(100_000)
)
