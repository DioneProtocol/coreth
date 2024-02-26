// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package delta

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/DioneProtocol/coreth/params"
	"github.com/stretchr/testify/require"
)

func TestFeeCalculator(t *testing.T) {
	tests := []struct {
		baseFee                    uint64
		priorityFee                uint64
		nodesAmount                uint64
		lpAllocation               uint64
		governanceAllocation       uint64
		priorityFeeOrionAllocation uint64
		orionAllocation            uint64
		maxOrionAllocation         uint64

		expectedBaseFee              uint64
		expectedPriorityFee          uint64
		expectedOrionFee             uint64
		expectedLpAllocation         uint64
		expectedGovernanceAllocation uint64
	}{
		{
			baseFee:         1_000_000,
			expectedBaseFee: 1_000_000,
		},
		{
			priorityFee:         1_000_000,
			expectedPriorityFee: 1_000_000,
		},
		{
			baseFee:             2_000_000,
			expectedBaseFee:     2_000_000,
			priorityFee:         1_000_000,
			expectedPriorityFee: 1_000_000,
		},
		{
			baseFee:                      1_000_000,
			governanceAllocation:         50,
			expectedBaseFee:              500_000,
			expectedGovernanceAllocation: 500_000,
		},
		{
			baseFee:                      1_000_000,
			governanceAllocation:         50,
			lpAllocation:                 50,
			expectedGovernanceAllocation: 500_000,
			expectedLpAllocation:         500_000,
		},
		{
			baseFee:                      1_000_000,
			governanceAllocation:         50,
			orionAllocation:              0,
			nodesAmount:                  1,
			maxOrionAllocation:           100,
			expectedBaseFee:              500_000,
			expectedOrionFee:             0,
			expectedGovernanceAllocation: 500_000,
		},
		{
			baseFee:                      1_000_000,
			governanceAllocation:         50,
			orionAllocation:              5,
			nodesAmount:                  1,
			maxOrionAllocation:           100,
			expectedBaseFee:              500_000,
			expectedOrionFee:             50_000,
			expectedGovernanceAllocation: 450_000,
		},
		{
			baseFee:              1_000_000,
			governanceAllocation: 50,
			orionAllocation:      5,
			nodesAmount:          10,
			maxOrionAllocation:   100,
			expectedBaseFee:      500_000,
			expectedOrionFee:     50_000,
		},
		{
			baseFee:                      1_000_000,
			governanceAllocation:         50,
			orionAllocation:              5,
			nodesAmount:                  10,
			maxOrionAllocation:           25,
			expectedBaseFee:              500_000,
			expectedOrionFee:             25_000,
			expectedGovernanceAllocation: 250_000,
		},
		{
			baseFee:                      1_000_000,
			governanceAllocation:         50,
			orionAllocation:              5,
			nodesAmount:                  20,
			maxOrionAllocation:           25,
			expectedBaseFee:              500_000,
			expectedOrionFee:             12_500,
			expectedGovernanceAllocation: 250_000,
		},
		{
			baseFee:                      1_000_000,
			governanceAllocation:         50,
			orionAllocation:              100,
			nodesAmount:                  20,
			maxOrionAllocation:           25,
			expectedBaseFee:              500_000,
			expectedOrionFee:             12_500,
			expectedGovernanceAllocation: 250_000,
		},
		{
			baseFee:              1_000_000,
			lpAllocation:         25,
			expectedBaseFee:      750_000,
			expectedLpAllocation: 250_000,
		},
		{
			baseFee:                      1_000_000,
			lpAllocation:                 25,
			governanceAllocation:         50,
			expectedBaseFee:              250_000,
			expectedLpAllocation:         250_000,
			expectedGovernanceAllocation: 500_000,
		},
		{
			baseFee:                      1_000_000,
			orionAllocation:              5,
			nodesAmount:                  1,
			maxOrionAllocation:           100,
			lpAllocation:                 25,
			governanceAllocation:         50,
			expectedBaseFee:              250_000,
			expectedLpAllocation:         250_000,
			expectedOrionFee:             50_000,
			expectedGovernanceAllocation: 450_000,
		},
		{
			priorityFee:                1_000_000,
			priorityFeeOrionAllocation: 50,
			expectedPriorityFee:        1_000_000,
		},
		{
			priorityFee:                1_000_000,
			priorityFeeOrionAllocation: 50,
			nodesAmount:                1,
			expectedOrionFee:           500_000,
			expectedPriorityFee:        500_000,
		},
		{
			priorityFee:                1_000_000,
			priorityFeeOrionAllocation: 50,
			nodesAmount:                2,
			expectedOrionFee:           250_000,
			expectedPriorityFee:        500_000,
		},
		{
			baseFee:                    1_000_000,
			priorityFee:                1_000_000,
			priorityFeeOrionAllocation: 50,
			nodesAmount:                2,
			lpAllocation:               25,
			expectedOrionFee:           250_000,
			expectedPriorityFee:        500_000,
			expectedBaseFee:            750_000,
			expectedLpAllocation:       250_000,
		},
		{
			baseFee:                      1_000_000,
			priorityFee:                  1_000_000,
			priorityFeeOrionAllocation:   50,
			nodesAmount:                  2,
			lpAllocation:                 25,
			governanceAllocation:         50,
			maxOrionAllocation:           100,
			expectedOrionFee:             250_000,
			expectedPriorityFee:          500_000,
			expectedBaseFee:              250_000,
			expectedLpAllocation:         250_000,
			expectedGovernanceAllocation: 500_000,
		},
		{
			baseFee:                      1_000_000,
			priorityFee:                  1_000_000,
			priorityFeeOrionAllocation:   50,
			nodesAmount:                  2,
			orionAllocation:              5,
			lpAllocation:                 25,
			governanceAllocation:         50,
			maxOrionAllocation:           100,
			expectedOrionFee:             300_000,
			expectedPriorityFee:          500_000,
			expectedBaseFee:              250_000,
			expectedLpAllocation:         250_000,
			expectedGovernanceAllocation: 400_000,
		},
		{
			baseFee:                      1_000_000,
			priorityFee:                  1_000_000,
			priorityFeeOrionAllocation:   50,
			nodesAmount:                  5,
			orionAllocation:              5,
			lpAllocation:                 25,
			governanceAllocation:         50,
			maxOrionAllocation:           100,
			expectedOrionFee:             150_000,
			expectedPriorityFee:          500_000,
			expectedBaseFee:              250_000,
			expectedLpAllocation:         250_000,
			expectedGovernanceAllocation: 250_000,
		},
		{
			// primary numbers
			baseFee:     1_002_577,
			priorityFee: 1_000_159,

			priorityFeeOrionAllocation: 50,
			nodesAmount:                5,
			orionAllocation:            5,
			lpAllocation:               25,
			governanceAllocation:       50,
			maxOrionAllocation:         100,

			expectedOrionFee:             150_143,
			expectedPriorityFee:          500_084,
			expectedBaseFee:              251_853,
			expectedLpAllocation:         250_644,
			expectedGovernanceAllocation: 249_439,
		},
	}

	for _, test := range tests {
		name := fmt.Sprintf("calculateFees(%d,%d,%d,%d,%d,%d,%d,%d)==(%d,%d,%d,%d,%d)",
			test.baseFee, test.priorityFee, test.nodesAmount, test.lpAllocation, test.governanceAllocation,
			test.priorityFee, test.orionAllocation, test.maxOrionAllocation, test.expectedBaseFee, test.expectedPriorityFee,
			test.expectedOrionFee, test.expectedLpAllocation, test.expectedGovernanceAllocation,
		)

		t.Run(name, func(t *testing.T) {
			rules := params.Rules{
				LpAllocation:               new(big.Int).SetUint64(test.lpAllocation),
				GovernanceAllocation:       new(big.Int).SetUint64(test.governanceAllocation),
				PriorityFeeOrionAllocation: new(big.Int).SetUint64(test.priorityFeeOrionAllocation),
				OrionAllocation:            new(big.Int).SetUint64(test.orionAllocation),
				MaxOrionAllocation:         new(big.Int).SetUint64(test.maxOrionAllocation),
				AllocationDenominator:      new(big.Int).SetUint64(100),
			}

			fees := CalculateFees(new(big.Int).SetUint64(test.baseFee), new(big.Int).SetUint64(test.priorityFee), test.nodesAmount, &rules)
			require.Equal(t, fees.PriorityFee.Uint64(), test.expectedPriorityFee, "Priority fee %d != %d", fees.PriorityFee, test.expectedPriorityFee)
			require.Equal(t, fees.OrionFee.Uint64(), test.expectedOrionFee, "Orion fee %d != %d", fees.OrionFee, test.expectedOrionFee)
			require.Equal(t, fees.LpAllocation.Uint64(), test.expectedLpAllocation, "Lp allocation %d != %d", fees.LpAllocation, test.expectedLpAllocation)
			require.Equal(t, fees.GovernanceAllocation.Uint64(), test.expectedGovernanceAllocation, "Governance allocation %d != %d", fees.GovernanceAllocation, test.expectedGovernanceAllocation)
			require.Equal(t, fees.BaseFee.Uint64(), test.expectedBaseFee, "Base fee %d != %d", fees.BaseFee, test.expectedBaseFee)
		})
	}
}
