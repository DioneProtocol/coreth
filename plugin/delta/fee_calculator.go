// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package delta

import (
	"math/big"

	"github.com/DioneProtocol/coreth/params"
)

type FeesDistribution struct {
	BaseFee              *big.Int
	PriorityFee          *big.Int
	LpAllocation         *big.Int
	GovernanceAllocation *big.Int
	OrionFee             *big.Int
}

func calculateToGovernanceAndOrion(totalBaseFee, orionAmount *big.Int, rules *params.Rules) (*big.Int, *big.Int) {
	governanceAndOrion := new(big.Int).Set(rules.GovernanceAllocation)
	governanceAndOrion.Mul(governanceAndOrion, totalBaseFee)
	governanceAndOrion.Div(governanceAndOrion, rules.AllocationDenominator)

	summaryOrionAllocation := new(big.Int).Set(rules.OrionAllocation)
	summaryOrionAllocation.Mul(summaryOrionAllocation, orionAmount)

	if summaryOrionAllocation.Cmp(rules.MaxOrionAllocation) > 0 {
		summaryOrionAllocation.Set(rules.MaxOrionAllocation)
	}

	if orionAmount.Sign() == 0 {
		return governanceAndOrion, new(big.Int)
	}

	summaryOrionAllocation.Mul(summaryOrionAllocation, totalBaseFee)
	summaryOrionAllocation.Div(summaryOrionAllocation, rules.AllocationDenominator)
	orionAllocation := new(big.Int).Div(summaryOrionAllocation, orionAmount)

	correctSummaryOrionAllocatoin := new(big.Int).Mul(orionAllocation, orionAmount)
	governanceAllocation := new(big.Int).Sub(governanceAndOrion, correctSummaryOrionAllocatoin)
	return governanceAllocation, correctSummaryOrionAllocatoin
}

func calculateToLp(totalBaseFee *big.Int, rules *params.Rules) *big.Int {
	lpAllocation := new(big.Int).Set(rules.LpAllocation)
	lpAllocation.Mul(lpAllocation, totalBaseFee)
	lpAllocation.Div(lpAllocation, rules.AllocationDenominator)
	return lpAllocation
}

func calculatePriorityFeeAndOrion(totalPriorityFee, orionAmount *big.Int, rules *params.Rules) (*big.Int, *big.Int) {
	summaryOrionAllocation := new(big.Int).Set(rules.PriorityFeeOrionAllocation)
	summaryOrionAllocation.Mul(summaryOrionAllocation, totalPriorityFee)
	summaryOrionAllocation.Div(summaryOrionAllocation, rules.AllocationDenominator)

	if orionAmount.Sign() == 0 {
		return totalPriorityFee, new(big.Int)
	}

	orionAllocation := new(big.Int).Div(summaryOrionAllocation, orionAmount)
	correctSummaryOrionAllocation := new(big.Int).Mul(orionAllocation, orionAmount)

	toPriorityFee := new(big.Int).Sub(totalPriorityFee, correctSummaryOrionAllocation)
	return toPriorityFee, correctSummaryOrionAllocation
}

func CalculateFees(totalBaseFee *big.Int, totalPriorityFee *big.Int, orionAmount uint64, rules *params.Rules) *FeesDistribution {
	totalBaseFee = new(big.Int).Set(totalBaseFee)
	totalPriorityFee = new(big.Int).Set(totalPriorityFee)

	orionAmountBigInt := new(big.Int).SetUint64(orionAmount)
	lpAllocation := calculateToLp(totalBaseFee, rules)
	governanceAllocation, orionFeeFromGovernance := calculateToGovernanceAndOrion(totalBaseFee, orionAmountBigInt, rules)
	totalPriorityFee, orionFeeFromPriorityFee := calculatePriorityFeeAndOrion(totalPriorityFee, orionAmountBigInt, rules)

	orionAllocation := new(big.Int).Set(orionFeeFromGovernance)
	orionAllocation.Add(orionAllocation, orionFeeFromPriorityFee)

	if orionAllocation.Sign() > 0 {
		orionAllocation.Div(orionAllocation, orionAmountBigInt)
	}

	totalBaseFee.Sub(totalBaseFee, lpAllocation)
	totalBaseFee.Sub(totalBaseFee, governanceAllocation)
	totalBaseFee.Sub(totalBaseFee, orionFeeFromGovernance)

	return &FeesDistribution{
		BaseFee:              totalBaseFee,
		PriorityFee:          totalPriorityFee,
		LpAllocation:         lpAllocation,
		GovernanceAllocation: governanceAllocation,
		OrionFee:             orionAllocation,
	}
}
