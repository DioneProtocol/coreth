// (c) 2019-2020, Ava Labs, Inc.
//
// This file is a derived work, based on the go-ethereum library whose original
// notices appear below.
//
// It is distributed under a license compatible with the licensing terms of the
// original code from which it is derived.
//
// Much love to the original authors for their work.
// **********
// Copyright 2017 The go-ethereum Authors
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

package params

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/DioneProtocol/odysseygo/ids"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var _ OrionNodesGetter = &orionNodesGetter{}

type stateGetter interface {
	GetState(addr common.Address, hash common.Hash) common.Hash
}

type OrionNodesGetter interface {
	GetLastUpdateTimestamp(stateGetter) uint64
	GetNodesList(stateGetter) []ids.NodeID
}

type orionNodesGetter struct {
	contract       common.Address
	lastUpdateSlot common.Hash
	sizeSlot       common.Hash
	listStartSlot  *big.Int
}

func NewOrionGetter(contract common.Address, lastUpdateSlot, orionsListSlot common.Hash) OrionNodesGetter {
	listStartSlot := crypto.Keccak256Hash(orionsListSlot[:])
	return &orionNodesGetter{
		contract:       contract,
		lastUpdateSlot: lastUpdateSlot,
		sizeSlot:       orionsListSlot,
		listStartSlot:  listStartSlot.Big(),
	}
}

func (o *orionNodesGetter) getUint64(state stateGetter, slot common.Hash) uint64 {
	hash := state.GetState(o.contract, slot)
	return binary.BigEndian.Uint64(hash[24:])
}

func (o *orionNodesGetter) GetLastUpdateTimestamp(state stateGetter) uint64 {
	return o.getUint64(state, o.lastUpdateSlot)
}

func (o *orionNodesGetter) GetNodesList(state stateGetter) []ids.NodeID {
	size := o.getUint64(state, o.sizeSlot)
	nodeIDs := make([]ids.NodeID, 0, size)

	for i := uint64(0); i < size; i++ {
		nodeIDslot := new(big.Int).Add(o.listStartSlot, new(big.Int).SetUint64(i))
		nodeIDHash := state.GetState(o.contract, common.BigToHash(nodeIDslot))
		fmt.Println(nodeIDHash)
		nodeID := ids.NodeID(nodeIDHash[:20])
		nodeIDs = append(nodeIDs, nodeID)
	}

	return nodeIDs
}
