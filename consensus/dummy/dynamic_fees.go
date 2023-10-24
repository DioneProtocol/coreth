// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package dummy

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/DioneProtocol/odysseygo/utils/wrappers"
	"github.com/DioneProtocol/coreth/core/types"
	"github.com/DioneProtocol/coreth/params"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

var (
	OdyPhase3MinBaseFee = big.NewInt(params.OdyPhase3MinBaseFee)
	OdyPhase3MaxBaseFee = big.NewInt(params.OdyPhase3MaxBaseFee)
	OdyPhase4MinBaseFee = big.NewInt(params.OdyPhase4MinBaseFee)
	OdyPhase4MaxBaseFee = big.NewInt(params.OdyPhase4MaxBaseFee)

	OdyPhase4BaseFeeChangeDenominator = new(big.Int).SetUint64(params.OdyPhase4BaseFeeChangeDenominator)
	OdyPhase5BaseFeeChangeDenominator = new(big.Int).SetUint64(params.OdyPhase5BaseFeeChangeDenominator)

	OdyPhase3BlockGasFee      uint64 = 1_000_000
	OdyPhase4MinBlockGasCost         = new(big.Int).Set(common.Big0)
	OdyPhase4MaxBlockGasCost         = big.NewInt(1_000_000)
	OdyPhase4BlockGasCostStep        = big.NewInt(50_000)
	OdyPhase4TargetBlockRate  uint64 = 2 // in seconds
	OdyPhase5BlockGasCostStep        = big.NewInt(200_000)
	rollupWindow                  uint64 = 10
)

// CalcBaseFee takes the previous header and the timestamp of its child block
// and calculates the expected base fee as well as the encoding of the past
// pricing information for the child block.
// CalcBaseFee should only be called if [timestamp] >= [config.OdyPhase3Timestamp]
func CalcBaseFee(config *params.ChainConfig, parent *types.Header, timestamp uint64) ([]byte, *big.Int, error) {
	// If the current block is the first EIP-1559 block, or it is the genesis block
	// return the initial slice and initial base fee.
	var (
		isOdyPhase3 = config.IsOdyPhase3(parent.Time)
		isOdyPhase4 = config.IsOdyPhase4(parent.Time)
		isOdyPhase5 = config.IsOdyPhase5(parent.Time)
	)
	if !isOdyPhase3 || parent.Number.Cmp(common.Big0) == 0 {
		initialSlice := make([]byte, params.OdyPhase3ExtraDataSize)
		initialBaseFee := big.NewInt(params.OdyPhase3InitialBaseFee)
		return initialSlice, initialBaseFee, nil
	}
	if uint64(len(parent.Extra)) != params.OdyPhase3ExtraDataSize {
		return nil, nil, fmt.Errorf("expected length of parent extra data to be %d, but found %d", params.OdyPhase3ExtraDataSize, len(parent.Extra))
	}

	if timestamp < parent.Time {
		return nil, nil, fmt.Errorf("cannot calculate base fee for timestamp (%d) prior to parent timestamp (%d)", timestamp, parent.Time)
	}
	roll := timestamp - parent.Time

	// roll the window over by the difference between the timestamps to generate
	// the new rollup window.
	newRollupWindow, err := rollLongWindow(parent.Extra, int(roll))
	if err != nil {
		return nil, nil, err
	}

	// If OP5, use a less responsive [BaseFeeChangeDenominator] and a higher gas
	// block limit
	var (
		baseFee                  = new(big.Int).Set(parent.BaseFee)
		baseFeeChangeDenominator = OdyPhase4BaseFeeChangeDenominator
		parentGasTarget          = params.OdyPhase3TargetGas
	)
	if isOdyPhase5 {
		baseFeeChangeDenominator = OdyPhase5BaseFeeChangeDenominator
		parentGasTarget = params.OdyPhase5TargetGas
	}
	parentGasTargetBig := new(big.Int).SetUint64(parentGasTarget)

	// Add in the gas used by the parent block in the correct place
	// If the parent consumed gas within the rollup window, add the consumed
	// gas in.
	if roll < rollupWindow {
		var blockGasCost, parentExtraStateGasUsed uint64
		switch {
		case isOdyPhase5:
			// [blockGasCost] has been removed in OP5, so it is left as 0.

			// At the start of a new network, the parent
			// may not have a populated [ExtDataGasUsed].
			if parent.ExtDataGasUsed != nil {
				parentExtraStateGasUsed = parent.ExtDataGasUsed.Uint64()
			}
		case isOdyPhase4:
			// The [blockGasCost] is paid by the effective tips in the block using
			// the block's value of [baseFee].
			blockGasCost = calcBlockGasCost(
				OdyPhase4TargetBlockRate,
				OdyPhase4MinBlockGasCost,
				OdyPhase4MaxBlockGasCost,
				OdyPhase4BlockGasCostStep,
				parent.BlockGasCost,
				parent.Time, timestamp,
			).Uint64()

			// On the boundary of OP3 and OP4 or at the start of a new network, the parent
			// may not have a populated [ExtDataGasUsed].
			if parent.ExtDataGasUsed != nil {
				parentExtraStateGasUsed = parent.ExtDataGasUsed.Uint64()
			}
		default:
			blockGasCost = OdyPhase3BlockGasFee
		}

		// Compute the new state of the gas rolling window.
		addedGas, overflow := math.SafeAdd(parent.GasUsed, parentExtraStateGasUsed)
		if overflow {
			addedGas = math.MaxUint64
		}

		// Only add the [blockGasCost] to the gas used if it isn't OP5
		if !isOdyPhase5 {
			addedGas, overflow = math.SafeAdd(addedGas, blockGasCost)
			if overflow {
				addedGas = math.MaxUint64
			}
		}

		slot := rollupWindow - 1 - roll
		start := slot * wrappers.LongLen
		updateLongWindow(newRollupWindow, start, addedGas)
	}

	// Calculate the amount of gas consumed within the rollup window.
	totalGas := sumLongWindow(newRollupWindow, int(rollupWindow))

	if totalGas == parentGasTarget {
		return newRollupWindow, baseFee, nil
	}

	if totalGas > parentGasTarget {
		// If the parent block used more gas than its target, the baseFee should increase.
		gasUsedDelta := new(big.Int).SetUint64(totalGas - parentGasTarget)
		x := new(big.Int).Mul(parent.BaseFee, gasUsedDelta)
		y := x.Div(x, parentGasTargetBig)
		baseFeeDelta := math.BigMax(
			x.Div(y, baseFeeChangeDenominator),
			common.Big1,
		)

		baseFee.Add(baseFee, baseFeeDelta)
	} else {
		// Otherwise if the parent block used less gas than its target, the baseFee should decrease.
		gasUsedDelta := new(big.Int).SetUint64(parentGasTarget - totalGas)
		x := new(big.Int).Mul(parent.BaseFee, gasUsedDelta)
		y := x.Div(x, parentGasTargetBig)
		baseFeeDelta := math.BigMax(
			x.Div(y, baseFeeChangeDenominator),
			common.Big1,
		)

		// If [roll] is greater than [rollupWindow], apply the state transition to the base fee to account
		// for the interval during which no blocks were produced.
		// We use roll/rollupWindow, so that the transition is applied for every [rollupWindow] seconds
		// that has elapsed between the parent and this block.
		if roll > rollupWindow {
			// Note: roll/rollupWindow must be greater than 1 since we've checked that roll > rollupWindow
			baseFeeDelta = baseFeeDelta.Mul(baseFeeDelta, new(big.Int).SetUint64(roll/rollupWindow))
		}
		baseFee.Sub(baseFee, baseFeeDelta)
	}

	// Ensure that the base fee does not increase/decrease outside of the bounds
	switch {
	case isOdyPhase5:
		baseFee = selectBigWithinBounds(OdyPhase4MinBaseFee, baseFee, nil)
	case isOdyPhase4:
		baseFee = selectBigWithinBounds(OdyPhase4MinBaseFee, baseFee, OdyPhase4MaxBaseFee)
	default:
		baseFee = selectBigWithinBounds(OdyPhase3MinBaseFee, baseFee, OdyPhase3MaxBaseFee)
	}

	return newRollupWindow, baseFee, nil
}

// EstiamteNextBaseFee attempts to estimate the next base fee based on a block with [parent] being built at
// [timestamp].
// If [timestamp] is less than the timestamp of [parent], then it uses the same timestamp as parent.
// Warning: This function should only be used in estimation and should not be used when calculating the canonical
// base fee for a subsequent block.
func EstimateNextBaseFee(config *params.ChainConfig, parent *types.Header, timestamp uint64) ([]byte, *big.Int, error) {
	if timestamp < parent.Time {
		timestamp = parent.Time
	}
	return CalcBaseFee(config, parent, timestamp)
}

// selectBigWithinBounds returns [value] if it is within the bounds:
// lowerBound <= value <= upperBound or the bound at either end if [value]
// is outside of the defined boundaries.
func selectBigWithinBounds(lowerBound, value, upperBound *big.Int) *big.Int {
	switch {
	case lowerBound != nil && value.Cmp(lowerBound) < 0:
		return new(big.Int).Set(lowerBound)
	case upperBound != nil && value.Cmp(upperBound) > 0:
		return new(big.Int).Set(upperBound)
	default:
		return value
	}
}

// rollWindow rolls the longs within [consumptionWindow] over by [roll] places.
// For example, if there are 4 longs encoded in a 32 byte slice, rollWindow would
// have the following effect:
// Original:
// [1, 2, 3, 4]
// Roll = 0
// [1, 2, 3, 4]
// Roll = 1
// [2, 3, 4, 0]
// Roll = 2
// [3, 4, 0, 0]
// Roll = 3
// [4, 0, 0, 0]
// Roll >= 4
// [0, 0, 0, 0]
// Assumes that [roll] is greater than or equal to 0
func rollWindow(consumptionWindow []byte, size, roll int) ([]byte, error) {
	if len(consumptionWindow)%size != 0 {
		return nil, fmt.Errorf("expected consumption window length (%d) to be a multiple of size (%d)", len(consumptionWindow), size)
	}

	// Note: make allocates a zeroed array, so we are guaranteed
	// that what we do not copy into, will be set to 0
	res := make([]byte, len(consumptionWindow))
	bound := roll * size
	if bound > len(consumptionWindow) {
		return res, nil
	}
	copy(res[:], consumptionWindow[roll*size:])
	return res, nil
}

func rollLongWindow(consumptionWindow []byte, roll int) ([]byte, error) {
	// Passes in [wrappers.LongLen] as the size of the individual value to be rolled over
	// so that it can be used to roll an array of long values.
	return rollWindow(consumptionWindow, wrappers.LongLen, roll)
}

// sumLongWindow sums [numLongs] encoded in [window]. Assumes that the length of [window]
// is sufficient to contain [numLongs] or else this function panics.
// If an overflow occurs, while summing the contents, the maximum uint64 value is returned.
func sumLongWindow(window []byte, numLongs int) uint64 {
	var (
		sum      uint64 = 0
		overflow bool
	)
	for i := 0; i < numLongs; i++ {
		// If an overflow occurs while summing the elements of the window, return the maximum
		// uint64 value immediately.
		sum, overflow = math.SafeAdd(sum, binary.BigEndian.Uint64(window[wrappers.LongLen*i:]))
		if overflow {
			return math.MaxUint64
		}
	}
	return sum
}

// updateLongWindow adds [gasConsumed] in at index within [window].
// Assumes that [index] has already been validated.
// If an overflow occurs, the maximum uint64 value is used.
func updateLongWindow(window []byte, start uint64, gasConsumed uint64) {
	prevGasConsumed := binary.BigEndian.Uint64(window[start:])

	totalGasConsumed, overflow := math.SafeAdd(prevGasConsumed, gasConsumed)
	if overflow {
		totalGasConsumed = math.MaxUint64
	}
	binary.BigEndian.PutUint64(window[start:], totalGasConsumed)
}

// calcBlockGasCost calculates the required block gas cost. If [parentTime]
// > [currentTime], the timeElapsed will be treated as 0.
func calcBlockGasCost(
	targetBlockRate uint64,
	minBlockGasCost *big.Int,
	maxBlockGasCost *big.Int,
	blockGasCostStep *big.Int,
	parentBlockGasCost *big.Int,
	parentTime, currentTime uint64,
) *big.Int {
	// Handle OP3/OP4 boundary by returning the minimum value as the boundary.
	if parentBlockGasCost == nil {
		return new(big.Int).Set(minBlockGasCost)
	}

	// Treat an invalid parent/current time combination as 0 elapsed time.
	var timeElapsed uint64
	if parentTime <= currentTime {
		timeElapsed = currentTime - parentTime
	}

	var blockGasCost *big.Int
	if timeElapsed < targetBlockRate {
		blockGasCostDelta := new(big.Int).Mul(blockGasCostStep, new(big.Int).SetUint64(targetBlockRate-timeElapsed))
		blockGasCost = new(big.Int).Add(parentBlockGasCost, blockGasCostDelta)
	} else {
		blockGasCostDelta := new(big.Int).Mul(blockGasCostStep, new(big.Int).SetUint64(timeElapsed-targetBlockRate))
		blockGasCost = new(big.Int).Sub(parentBlockGasCost, blockGasCostDelta)
	}

	blockGasCost = selectBigWithinBounds(minBlockGasCost, blockGasCost, maxBlockGasCost)
	if !blockGasCost.IsUint64() {
		blockGasCost = new(big.Int).SetUint64(math.MaxUint64)
	}
	return blockGasCost
}

// MinRequiredTip is the estimated minimum tip a transaction would have
// needed to pay to be included in a given block (assuming it paid a tip
// proportional to its gas usage). In reality, there is no minimum tip that
// is enforced by the consensus engine and high tip paying transactions can
// subsidize the inclusion of low tip paying transactions. The only
// correctness check performed is that the sum of all tips is >= the
// required block fee.
//
// This function will return nil for all return values prior to Ody Phase 4.
func MinRequiredTip(config *params.ChainConfig, header *types.Header) (*big.Int, error) {
	if !config.IsOdyPhase4(header.Time) {
		return nil, nil
	}
	if header.BaseFee == nil {
		return nil, errBaseFeeNil
	}
	if header.BlockGasCost == nil {
		return nil, errBlockGasCostNil
	}
	if header.ExtDataGasUsed == nil {
		return nil, errExtDataGasUsedNil
	}

	// minTip = requiredBlockFee/blockGasUsage
	requiredBlockFee := new(big.Int).Mul(
		header.BlockGasCost,
		header.BaseFee,
	)
	blockGasUsage := new(big.Int).Add(
		new(big.Int).SetUint64(header.GasUsed),
		header.ExtDataGasUsed,
	)
	return new(big.Int).Div(requiredBlockFee, blockGasUsage), nil
}
