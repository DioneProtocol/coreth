package params

import (
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
)

var _ StakeGovernanceGetter = &stakeGovernanceGetter{}

type StakeGovernanceGetter interface {
	GetStakingValue(stateGetter) uint64
}

type stakeGovernanceGetter struct {
	stakingGovernanceConfigSlot common.Hash
	stakingGovernanceAddress    common.Address
}

func NewStakingGovernanceGetter(stakingGovernanceConfigSlot common.Hash, stakingGovernanceAddress common.Address) StakeGovernanceGetter {
	return &stakeGovernanceGetter{
		stakingGovernanceConfigSlot: stakingGovernanceConfigSlot,
		stakingGovernanceAddress:    stakingGovernanceAddress,
	}
}

func (s *stakeGovernanceGetter) GetStakingValue(state stateGetter) uint64 {
	return s.getUint64ForGovernanceConfig(state, s.stakingGovernanceConfigSlot)
}

func (s *stakeGovernanceGetter) getUint64ForGovernanceConfig(state stateGetter, slot common.Hash) uint64 {
	hash := state.GetState(s.stakingGovernanceAddress, slot)
	return binary.BigEndian.Uint64(hash[24:])
}
