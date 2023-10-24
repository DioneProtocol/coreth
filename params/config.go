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
// Copyright 2016 The go-ethereum Authors
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
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/DioneProtocol/coreth/precompile"
	"github.com/DioneProtocol/coreth/utils"
	"github.com/ethereum/go-ethereum/common"
)

// Odyssey ChainIDs
var (
	// OdysseyMainnetChainID ...
	OdysseyMainnetChainID = big.NewInt(153)
	// OdysseyOdytChainID ...
	OdysseyOdytChainID = big.NewInt(13)
	// OdysseyLocalChainID ...
	OdysseyLocalChainID = big.NewInt(12)

	errNonGenesisForkByHeight = errors.New("coreth only supports forking by height at the genesis block")
)

var (
	// OdysseyMainnetChainConfig is the configuration for Odyssey Main Network
	OdysseyMainnetChainConfig = &ChainConfig{
		ChainID:                         OdysseyMainnetChainID,
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    big.NewInt(0),
		DAOForkSupport:                  true,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePre6BlockTimestamp:  	 utils.NewUint64(0),
		OdyPhase6BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePost6BlockTimestamp: 	 utils.NewUint64(0),
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             utils.NewUint64(0),
		CortinaBlockTimestamp:           utils.NewUint64(0),
	}

	// OdysseyOdytChainConfig is the configuration for the Odyt Test Network
	OdysseyOdytChainConfig = &ChainConfig{
		ChainID:                         OdysseyOdytChainID,
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    big.NewInt(0),
		DAOForkSupport:                  true,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePre6BlockTimestamp:  	 utils.NewUint64(0),
		OdyPhase6BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePost6BlockTimestamp: 	 utils.NewUint64(0),
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             utils.NewUint64(0),
		CortinaBlockTimestamp:           utils.NewUint64(0),
	}

	// OdysseyLocalChainConfig is the configuration for the Odyssey Local Network
	OdysseyLocalChainConfig = &ChainConfig{
		ChainID:                         OdysseyLocalChainID,
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    big.NewInt(0),
		DAOForkSupport:                  true,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePre6BlockTimestamp:  	 utils.NewUint64(0),
		OdyPhase6BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePost6BlockTimestamp: 	 utils.NewUint64(0),
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             utils.NewUint64(0),
		CortinaBlockTimestamp:           utils.NewUint64(0),
		DUpgradeBlockTimestamp:          utils.NewUint64(0),
	}

	TestChainConfig = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePre6BlockTimestamp:  	 utils.NewUint64(0),
		OdyPhase6BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePost6BlockTimestamp: 	 utils.NewUint64(0),
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             utils.NewUint64(0),
		CortinaBlockTimestamp:           utils.NewUint64(0),
		DUpgradeBlockTimestamp:          utils.NewUint64(0),
	}

	TestLaunchConfig = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 nil,
		OdyPhase2BlockTimestamp:     	 nil,
		OdyPhase3BlockTimestamp:     	 nil,
		OdyPhase4BlockTimestamp:     	 nil,
		OdyPhase5BlockTimestamp:     	 nil,
		OdyPhasePre6BlockTimestamp:  	 nil,
		OdyPhase6BlockTimestamp:     	 nil,
		OdyPhasePost6BlockTimestamp: 	 nil,
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             nil,
		CortinaBlockTimestamp:           nil,
		DUpgradeBlockTimestamp:          nil,
	}

	TestOdyPhase1Config = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 nil,
		OdyPhase3BlockTimestamp:     	 nil,
		OdyPhase4BlockTimestamp:     	 nil,
		OdyPhase5BlockTimestamp:     	 nil,
		OdyPhasePre6BlockTimestamp:  	 nil,
		OdyPhase6BlockTimestamp:     	 nil,
		OdyPhasePost6BlockTimestamp: 	 nil,
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             nil,
		CortinaBlockTimestamp:           nil,
		DUpgradeBlockTimestamp:          nil,
	}

	TestOdyPhase2Config = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 nil,
		OdyPhase4BlockTimestamp:     	 nil,
		OdyPhase5BlockTimestamp:     	 nil,
		OdyPhasePre6BlockTimestamp:  	 nil,
		OdyPhase6BlockTimestamp:     	 nil,
		OdyPhasePost6BlockTimestamp: 	 nil,
		BanffBlockTimestamp:             nil,
		CortinaBlockTimestamp:           nil,
		DUpgradeBlockTimestamp:          nil,
	}

	TestOdyPhase3Config = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 nil,
		OdyPhase5BlockTimestamp:     	 nil,
		OdyPhasePre6BlockTimestamp:  	 nil,
		OdyPhase6BlockTimestamp:     	 nil,
		OdyPhasePost6BlockTimestamp: 	 nil,
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             nil,
		CortinaBlockTimestamp:           nil,
		DUpgradeBlockTimestamp:          nil,
	}

	TestOdyPhase4Config = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:    	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:   	     nil,
		OdyPhasePre6BlockTimestamp:  	 nil,
		OdyPhase6BlockTimestamp:     	 nil,
		OdyPhasePost6BlockTimestamp: 	 nil,
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             nil,
		CortinaBlockTimestamp:           nil,
		DUpgradeBlockTimestamp:          nil,
	}

	TestOdyPhase5Config = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePre6BlockTimestamp:  	 nil,
		OdyPhase6BlockTimestamp:     	 nil,
		OdyPhasePost6BlockTimestamp: 	 nil,
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             nil,
		CortinaBlockTimestamp:           nil,
		DUpgradeBlockTimestamp:          nil,
	}

	TestOdyPhasePre6Config = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePre6BlockTimestamp:  	 utils.NewUint64(0),
		OdyPhase6BlockTimestamp:     	 nil,
		OdyPhasePost6BlockTimestamp: 	 nil,
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             nil,
		CortinaBlockTimestamp:           nil,
		DUpgradeBlockTimestamp:          nil,
	}

	TestOdyPhase6Config = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePre6BlockTimestamp:  	 utils.NewUint64(0),
		OdyPhase6BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePost6BlockTimestamp: 	 nil,
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             nil,
		CortinaBlockTimestamp:           nil,
		DUpgradeBlockTimestamp:          nil,
	}

	TestOdyPhasePost6Config = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePre6BlockTimestamp:  	 utils.NewUint64(0),
		OdyPhase6BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePost6BlockTimestamp: 	 utils.NewUint64(0),
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             nil,
		CortinaBlockTimestamp:           nil,
		DUpgradeBlockTimestamp:          nil,
	}

	TestOdyPhase7Config = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePre6BlockTimestamp:  	 utils.NewUint64(0),
		OdyPhase6BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePost6BlockTimestamp: 	 utils.NewUint64(0),
		OdyPhase7BlockTimestamp:     	 utils.NewUint64(0),
		BanffBlockTimestamp:             nil,
		CortinaBlockTimestamp:           nil,
		DUpgradeBlockTimestamp:          nil,
	}

	TestBanffChainConfig = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePre6BlockTimestamp:  	 utils.NewUint64(0),
		OdyPhase6BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePost6BlockTimestamp: 	 utils.NewUint64(0),
		BanffBlockTimestamp:             utils.NewUint64(0),
		OdyPhase7BlockTimestamp:     	 nil,
		CortinaBlockTimestamp:           nil,
		DUpgradeBlockTimestamp:          nil,
	}

	TestCortinaChainConfig = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePre6BlockTimestamp:  	 utils.NewUint64(0),
		OdyPhase6BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePost6BlockTimestamp: 	 utils.NewUint64(0),
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             utils.NewUint64(0),
		CortinaBlockTimestamp:           utils.NewUint64(0),
		DUpgradeBlockTimestamp:          nil,
	}

	TestDUpgradeChainConfig = &ChainConfig{
		OdysseyContext:                OdysseyContext{common.Hash{1}},
		ChainID:                         big.NewInt(1),
		HomesteadBlock:                  big.NewInt(0),
		DAOForkBlock:                    nil,
		DAOForkSupport:                  false,
		EIP150Block:                     big.NewInt(0),
		EIP155Block:                     big.NewInt(0),
		EIP158Block:                     big.NewInt(0),
		ByzantiumBlock:                  big.NewInt(0),
		ConstantinopleBlock:             big.NewInt(0),
		PetersburgBlock:                 big.NewInt(0),
		IstanbulBlock:                   big.NewInt(0),
		MuirGlacierBlock:                big.NewInt(0),
		OdyPhase1BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase2BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase3BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase4BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhase5BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePre6BlockTimestamp:  	 utils.NewUint64(0),
		OdyPhase6BlockTimestamp:     	 utils.NewUint64(0),
		OdyPhasePost6BlockTimestamp: 	 utils.NewUint64(0),
		OdyPhase7BlockTimestamp:     	 nil,
		BanffBlockTimestamp:             utils.NewUint64(0),
		CortinaBlockTimestamp:           utils.NewUint64(0),
	}

	TestRules = TestChainConfig.OdysseyRules(new(big.Int), 0)
)

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	OdysseyContext `json:"-"` // Odyssey specific context set during VM initialization. Not serialized.

	ChainID *big.Int `json:"chainId"` // chainId identifies the current chain and is used for replay protection

	HomesteadBlock *big.Int `json:"homesteadBlock,omitempty"` // Homestead switch block (nil = no fork, 0 = already homestead)

	DAOForkBlock   *big.Int `json:"daoForkBlock,omitempty"`   // TheDAO hard-fork switch block (nil = no fork)
	DAOForkSupport bool     `json:"daoForkSupport,omitempty"` // Whether the nodes supports or opposes the DAO hard-fork

	// EIP150 implements the Gas price changes (https://github.com/ethereum/EIPs/issues/150)
	EIP150Block *big.Int `json:"eip150Block,omitempty"` // EIP150 HF block (nil = no fork)
	EIP155Block *big.Int `json:"eip155Block,omitempty"` // EIP155 HF block
	EIP158Block *big.Int `json:"eip158Block,omitempty"` // EIP158 HF block

	ByzantiumBlock      *big.Int `json:"byzantiumBlock,omitempty"`      // Byzantium switch block (nil = no fork, 0 = already on byzantium)
	ConstantinopleBlock *big.Int `json:"constantinopleBlock,omitempty"` // Constantinople switch block (nil = no fork, 0 = already activated)
	PetersburgBlock     *big.Int `json:"petersburgBlock,omitempty"`     // Petersburg switch block (nil = same as Constantinople)
	IstanbulBlock       *big.Int `json:"istanbulBlock,omitempty"`       // Istanbul switch block (nil = no fork, 0 = already on istanbul)
	MuirGlacierBlock    *big.Int `json:"muirGlacierBlock,omitempty"`    // Eip-2384 (bomb delay) switch block (nil = no fork, 0 = already activated)

	// Odyssey Network Upgrades
	OdyPhase1BlockTimestamp *uint64 `json:"odyPhase1BlockTimestamp,omitempty"` // Ody Phase 1 Block Timestamp (nil = no fork, 0 = already activated)
	// Ody Phase 2 Block Timestamp (nil = no fork, 0 = already activated)
	// Ody Phase 2 includes a modified version of the Berlin Hard Fork from Ethereum
	OdyPhase2BlockTimestamp *uint64 `json:"odyPhase2BlockTimestamp,omitempty"`
	// Ody Phase 3 introduces dynamic fees and a modified version of the London Hard Fork from Ethereum (nil = no fork, 0 = already activated)
	OdyPhase3BlockTimestamp *uint64 `json:"odyPhase3BlockTimestamp,omitempty"`
	// Ody Phase 4 introduces the notion of a block fee to the dynamic fee algorithm (nil = no fork, 0 = already activated)
	OdyPhase4BlockTimestamp *uint64 `json:"odyPhase4BlockTimestamp,omitempty"`
	// Ody Phase 5 introduces a batch of atomic transactions with a maximum atomic gas limit per block. (nil = no fork, 0 = already activated)
	OdyPhase5BlockTimestamp *uint64 `json:"odyPhase5BlockTimestamp,omitempty"`
	// Ody Phase Pre-6 deprecates the NativeAssetCall precompile (soft). (nil = no fork, 0 = already activated)
	OdyPhasePre6BlockTimestamp *uint64 `json:"odyPhasePre6BlockTimestamp,omitempty"`
	// Ody Phase 6 deprecates the NativeAssetBalance and NativeAssetCall precompiles. (nil = no fork, 0 = already activated)
	OdyPhase6BlockTimestamp *uint64 `json:"odyPhase6BlockTimestamp,omitempty"`
	// Ody Phase Post-6 deprecates the NativeAssetCall precompile (soft). (nil = no fork, 0 = already activated)
	OdyPhasePost6BlockTimestamp *uint64 `json:"odyPhasePost6BlockTimestamp,omitempty"`
	// Banff restricts import/export transactions to DIONE. (nil = no fork, 0 = already activated)
	BanffBlockTimestamp *uint64 `json:"banffBlockTimestamp,omitempty"`
	// Cortina increases the block gas limit to 15M. (nil = no fork, 0 = already activated)
	CortinaBlockTimestamp *uint64 `json:"cortinaBlockTimestamp,omitempty"`
	// DUpgrade activates the Shanghai upgrade from Ethereum. (nil = no fork, 0 = already activated)
	DUpgradeBlockTimestamp *uint64 `json:"dUpgradeBlockTimestamp,omitempty"`
	// Cancun activates the Cancun upgrade from Ethereum. (nil = no fork, 0 = already activated)
	CancunTime *uint64 `json:"cancunTime,omitempty"`
	// Ody Phase 7 Enables new rewarding calculation based on the provided timestamp. (nil = no fork, 0 = already activated)
	OdyPhase7BlockTimestamp *uint64 `json:"odyPhase7BlockTimestamp,omitempty"`
}

// OdysseyContext provides Odyssey specific context directly into the EVM.
type OdysseyContext struct {
	BlockchainID common.Hash
}

// Description returns a human-readable description of ChainConfig.
func (c *ChainConfig) Description() string {
	var banner string

	banner += fmt.Sprintf("Chain ID:  %v\n", c.ChainID)
	banner += "Consensus: Dummy Consensus Engine\n\n"

	// Create a list of forks with a short description of them. Forks that only
	// makes sense for mainnet should be optional at printing to avoid bloating
	// the output for testnets and private networks.
	banner += "Hard Forks:\n"
	banner += fmt.Sprintf(" - Homestead:                   #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/homestead.md)\n", c.HomesteadBlock)
	if c.DAOForkBlock != nil {
		banner += fmt.Sprintf(" - DAO Fork:                    #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/dao-fork.md)\n", c.DAOForkBlock)
	}
	banner += fmt.Sprintf(" - Tangerine Whistle (EIP 150): #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/tangerine-whistle.md)\n", c.EIP150Block)
	banner += fmt.Sprintf(" - Spurious Dragon/1 (EIP 155): #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/spurious-dragon.md)\n", c.EIP155Block)
	banner += fmt.Sprintf(" - Spurious Dragon/2 (EIP 158): #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/spurious-dragon.md)\n", c.EIP155Block)
	banner += fmt.Sprintf(" - Byzantium:                   #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/byzantium.md)\n", c.ByzantiumBlock)
	banner += fmt.Sprintf(" - Constantinople:              #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/constantinople.md)\n", c.ConstantinopleBlock)
	banner += fmt.Sprintf(" - Petersburg:                  #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/petersburg.md)\n", c.PetersburgBlock)
	banner += fmt.Sprintf(" - Istanbul:                    #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/istanbul.md)\n", c.IstanbulBlock)
	if c.MuirGlacierBlock != nil {
		banner += fmt.Sprintf(" - Muir Glacier:                #%-8v (https://github.com/ethereum/execution-specs/blob/master/network-upgrades/mainnet-upgrades/muir-glacier.md)\n", c.MuirGlacierBlock)
	}
	banner += fmt.Sprintf(" - Ody Phase 1 Timestamp:        #%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.3.0)\n", c.OdyPhase1BlockTimestamp)
	banner += fmt.Sprintf(" - Ody Phase 2 Timestamp:        #%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.4.0)\n", c.OdyPhase2BlockTimestamp)
	banner += fmt.Sprintf(" - Ody Phase 3 Timestamp:        #%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.5.0)\n", c.OdyPhase3BlockTimestamp)
	banner += fmt.Sprintf(" - Ody Phase 4 Timestamp:        #%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.6.0)\n", c.OdyPhase4BlockTimestamp)
	banner += fmt.Sprintf(" - Ody Phase 5 Timestamp:        #%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.7.0)\n", c.OdyPhase5BlockTimestamp)
	banner += fmt.Sprintf(" - Ody Phase P6 Timestamp        #%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.8.0)\n", c.OdyPhasePre6BlockTimestamp)
	banner += fmt.Sprintf(" - Ody Phase 6 Timestamp:        #%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.8.0)\n", c.OdyPhase6BlockTimestamp)
	banner += fmt.Sprintf(" - Ody Phase Post-6 Timestamp:   #%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.8.0\n", c.OdyPhasePost6BlockTimestamp)
	banner += fmt.Sprintf(" - Banff Timestamp:              #%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.9.0)\n", c.BanffBlockTimestamp)
	banner += fmt.Sprintf(" - Cortina Timestamp:            #%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.10.0)\n", c.CortinaBlockTimestamp)
	banner += fmt.Sprintf(" - DUpgrade Timestamp:           #%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.11.0)\n", c.DUpgradeBlockTimestamp)
	banner += fmt.Sprintf(" - Cancun Timestamp:             #%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.11.0)\n", c.DUpgradeBlockTimestamp)
	banner += fmt.Sprintf(" - Ody Phase 7 Timestamp:  		#%-8v (https://github.com/DioneProtocol/odysseygo/releases/tag/v1.12.0\n", c.OdyPhase7BlockTimestamp)
	banner += "\n"
	return banner
}

// IsHomestead returns whether num is either equal to the homestead block or greater.
func (c *ChainConfig) IsHomestead(num *big.Int) bool {
	return utils.IsBlockForked(c.HomesteadBlock, num)
}

// IsDAOFork returns whether num is either equal to the DAO fork block or greater.
func (c *ChainConfig) IsDAOFork(num *big.Int) bool {
	return utils.IsBlockForked(c.DAOForkBlock, num)
}

// IsEIP150 returns whether num is either equal to the EIP150 fork block or greater.
func (c *ChainConfig) IsEIP150(num *big.Int) bool {
	return utils.IsBlockForked(c.EIP150Block, num)
}

// IsEIP155 returns whether num is either equal to the EIP155 fork block or greater.
func (c *ChainConfig) IsEIP155(num *big.Int) bool {
	return utils.IsBlockForked(c.EIP155Block, num)
}

// IsEIP158 returns whether num is either equal to the EIP158 fork block or greater.
func (c *ChainConfig) IsEIP158(num *big.Int) bool {
	return utils.IsBlockForked(c.EIP158Block, num)
}

// IsByzantium returns whether num is either equal to the Byzantium fork block or greater.
func (c *ChainConfig) IsByzantium(num *big.Int) bool {
	return utils.IsBlockForked(c.ByzantiumBlock, num)
}

// IsConstantinople returns whether num is either equal to the Constantinople fork block or greater.
func (c *ChainConfig) IsConstantinople(num *big.Int) bool {
	return utils.IsBlockForked(c.ConstantinopleBlock, num)
}

// IsMuirGlacier returns whether num is either equal to the Muir Glacier (EIP-2384) fork block or greater.
func (c *ChainConfig) IsMuirGlacier(num *big.Int) bool {
	return utils.IsBlockForked(c.MuirGlacierBlock, num)
}

// IsPetersburg returns whether num is either
// - equal to or greater than the PetersburgBlock fork block,
// - OR is nil, and Constantinople is active
func (c *ChainConfig) IsPetersburg(num *big.Int) bool {
	return utils.IsBlockForked(c.PetersburgBlock, num) || c.PetersburgBlock == nil && utils.IsBlockForked(c.ConstantinopleBlock, num)
}

// IsIstanbul returns whether num is either equal to the Istanbul fork block or greater.
func (c *ChainConfig) IsIstanbul(num *big.Int) bool {
	return utils.IsBlockForked(c.IstanbulBlock, num)
}

// Odyssey Upgrades:

// IsOdyPhase1 returns whether [time] represents a block
// with a timestamp after the Ody Phase 1 upgrade time.
func (c *ChainConfig) IsOdyPhase1(time uint64) bool {
	return utils.IsTimestampForked(c.OdyPhase1BlockTimestamp, time)
}

// IsOdyPhase2 returns whether [time] represents a block
// with a timestamp after the Ody Phase 2 upgrade time.
func (c *ChainConfig) IsOdyPhase2(time uint64) bool {
	return utils.IsTimestampForked(c.OdyPhase2BlockTimestamp, time)
}

// IsOdyPhase3 returns whether [time] represents a block
// with a timestamp after the Ody Phase 3 upgrade time.
func (c *ChainConfig) IsOdyPhase3(time uint64) bool {
	return utils.IsTimestampForked(c.OdyPhase3BlockTimestamp, time)
}

// IsOdyPhase4 returns whether [time] represents a block
// with a timestamp after the Ody Phase 4 upgrade time.
func (c *ChainConfig) IsOdyPhase4(time uint64) bool {
	return utils.IsTimestampForked(c.OdyPhase4BlockTimestamp, time)
}

// IsOdyPhase5 returns whether [time] represents a block
// with a timestamp after the Ody Phase 5 upgrade time.
func (c *ChainConfig) IsOdyPhase5(time uint64) bool {
	return utils.IsTimestampForked(c.OdyPhase5BlockTimestamp, time)
}

// IsOdyPhasePre6 returns whether [time] represents a block
// with a timestamp after the Ody Phase Pre 6 upgrade time.
func (c *ChainConfig) IsOdyPhasePre6(time uint64) bool {
	return utils.IsTimestampForked(c.OdyPhasePre6BlockTimestamp, time)
}

// IsOdyPhase6 returns whether [time] represents a block
// with a timestamp after the Ody Phase 6 upgrade time.
func (c *ChainConfig) IsOdyPhase6(time uint64) bool {
	return utils.IsTimestampForked(c.OdyPhase6BlockTimestamp, time)
}

// IsOdyPhasePost6 returns whether [time] represents a block
// with a timestamp after the Ody Phase 6 Post upgrade time.
func (c *ChainConfig) IsOdyPhasePost6(time uint64) bool {
	return utils.IsTimestampForked(c.OdyPhasePost6BlockTimestamp, time)
}

// IsOdyPhas7 returns whether [time] represents a block
// with a timestamp after the Ody Phase 7 upgrade time.
func (c *ChainConfig) IsOdyPhase7(time uint64) bool {
	return utils.IsTimestampForked(c.OdyPhase7BlockTimestamp, time)
}

// IsBanff returns whether [time] represents a block
// with a timestamp after the Banff upgrade time.
func (c *ChainConfig) IsBanff(time uint64) bool {
	return utils.IsTimestampForked(c.BanffBlockTimestamp, time)
}

// IsCortina returns whether [time] represents a block
// with a timestamp after the Cortina upgrade time.
func (c *ChainConfig) IsCortina(time uint64) bool {
	return utils.IsTimestampForked(c.CortinaBlockTimestamp, time)
}

// IsDUpgrade returns whether [time] represents a block
// with a timestamp after the DUpgrade upgrade time.
func (c *ChainConfig) IsDUpgrade(time uint64) bool {
	return utils.IsTimestampForked(c.DUpgradeBlockTimestamp, time)
}

// IsCancun returns whether [time] represents a block
// with a timestamp after the Cancun upgrade time.
func (c *ChainConfig) IsCancun(time uint64) bool {
	return utils.IsTimestampForked(c.CancunTime, time)
}

// CheckCompatible checks whether scheduled fork transitions have been imported
// with a mismatching chain configuration.
func (c *ChainConfig) CheckCompatible(newcfg *ChainConfig, height uint64, time uint64) *ConfigCompatError {
	var (
		bhead = new(big.Int).SetUint64(height)
		btime = time
	)
	// Iterate checkCompatible to find the lowest conflict.
	var lasterr *ConfigCompatError
	for {
		err := c.checkCompatible(newcfg, bhead, btime)
		if err == nil || (lasterr != nil && err.RewindToBlock == lasterr.RewindToBlock && err.RewindToTime == lasterr.RewindToTime) {
			break
		}
		lasterr = err

		if err.RewindToTime > 0 {
			btime = err.RewindToTime
		} else {
			bhead.SetUint64(err.RewindToBlock)
		}
	}
	return lasterr
}

// CheckConfigForkOrder checks that we don't "skip" any forks, geth isn't pluggable enough
// to guarantee that forks can be implemented in a different order than on official networks
func (c *ChainConfig) CheckConfigForkOrder() error {
	type fork struct {
		name      string
		block     *big.Int // some go-ethereum forks use block numbers
		timestamp *uint64  // Odyssey forks use timestamps
		optional  bool     // if true, the fork may be nil and next fork is still allowed
	}
	var lastFork fork
	for _, cur := range []fork{
		{name: "homesteadBlock", block: c.HomesteadBlock},
		{name: "daoForkBlock", block: c.DAOForkBlock, optional: true},
		{name: "eip150Block", block: c.EIP150Block},
		{name: "eip155Block", block: c.EIP155Block},
		{name: "eip158Block", block: c.EIP158Block},
		{name: "byzantiumBlock", block: c.ByzantiumBlock},
		{name: "constantinopleBlock", block: c.ConstantinopleBlock},
		{name: "petersburgBlock", block: c.PetersburgBlock},
		{name: "istanbulBlock", block: c.IstanbulBlock},
		{name: "muirGlacierBlock", block: c.MuirGlacierBlock, optional: true},
	} {
		if cur.block != nil && common.Big0.Cmp(cur.block) != 0 {
			return errNonGenesisForkByHeight
		}
		if lastFork.name != "" {
			// Next one must be higher number
			if lastFork.block == nil && cur.block != nil {
				return fmt.Errorf("unsupported fork ordering: %v not enabled, but %v enabled at %v",
					lastFork.name, cur.name, cur.block)
			}
			if lastFork.block != nil && cur.block != nil {
				if lastFork.block.Cmp(cur.block) > 0 {
					return fmt.Errorf("unsupported fork ordering: %v enabled at %v, but %v enabled at %v",
						lastFork.name, lastFork.block, cur.name, cur.block)
				}
			}
		}
		// If it was optional and not set, then ignore it
		if !cur.optional || cur.block != nil {
			lastFork = cur
		}
	}

	// Note: OdyPhase1 and OdyPhase2 override the rules set by block number
	// hard forks. In Odyssey, hard forks must take place via block timestamps instead
	// of block numbers since blocks are produced asynchronously. Therefore, we do not
	// check that the block timestamps for Ody Phase1 and Phase2 in the same way as for
	// the block number forks since it would not be a meaningful comparison.
	// Instead, we check only that Ody Phases are enabled in order.
	lastFork = fork{}
	for _, cur := range []fork{
		{name: "odyPhase1BlockTimestamp", timestamp: c.OdyPhase1BlockTimestamp},
		{name: "odyPhase2BlockTimestamp", timestamp: c.OdyPhase2BlockTimestamp},
		{name: "odyPhase3BlockTimestamp", timestamp: c.OdyPhase3BlockTimestamp},
		{name: "odyPhase4BlockTimestamp", timestamp: c.OdyPhase4BlockTimestamp},
		{name: "odyPhase5BlockTimestamp", timestamp: c.OdyPhase5BlockTimestamp},
		{name: "odyPhasePre6BlockTimestamp", timestamp: c.OdyPhasePre6BlockTimestamp},
		{name: "odyPhase6BlockTimestamp", timestamp: c.OdyPhase6BlockTimestamp},
		{name: "odyPhasePost6BlockTimestamp", timestamp: c.OdyPhasePost6BlockTimestamp},
		{name: "odyPhase7BlockTimestamp", timestamp: c.OdyPhase7BlockTimestamp},
		{name: "banffBlockTimestamp", timestamp: c.BanffBlockTimestamp},
		{name: "cortinaBlockTimestamp", timestamp: c.CortinaBlockTimestamp},
		{name: "dUpgradeBlockTimestamp", timestamp: c.DUpgradeBlockTimestamp},
		{name: "cancunTime", timestamp: c.CancunTime},
	} {
		if lastFork.name != "" {
			// Next one must be higher number
			if lastFork.timestamp == nil && cur.timestamp != nil {
				return fmt.Errorf("unsupported fork ordering: %v not enabled, but %v enabled at %v",
					lastFork.name, cur.name, cur.timestamp)
			}
			if lastFork.timestamp != nil && cur.timestamp != nil {
				if *lastFork.timestamp > *cur.timestamp {
					return fmt.Errorf("unsupported fork ordering: %v enabled at %v, but %v enabled at %v",
						lastFork.name, lastFork.timestamp, cur.name, cur.timestamp)
				}
			}
		}
		// If it was optional and not set, then ignore it
		if !cur.optional || cur.timestamp != nil {
			lastFork = cur
		}
	}
	// TODO(aaronbuchwald) check that odyssey block timestamps are at least possible with the other rule set changes
	// additional change: require that block number hard forks are either 0 or nil since they should not
	// be enabled at a specific block number.

	return nil
}

func (c *ChainConfig) checkCompatible(newcfg *ChainConfig, height *big.Int, time uint64) *ConfigCompatError {
	if isForkBlockIncompatible(c.HomesteadBlock, newcfg.HomesteadBlock, height) {
		return newBlockCompatError("Homestead fork block", c.HomesteadBlock, newcfg.HomesteadBlock)
	}
	if isForkBlockIncompatible(c.DAOForkBlock, newcfg.DAOForkBlock, height) {
		return newBlockCompatError("DAO fork block", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if c.IsDAOFork(height) && c.DAOForkSupport != newcfg.DAOForkSupport {
		return newBlockCompatError("DAO fork support flag", c.DAOForkBlock, newcfg.DAOForkBlock)
	}
	if isForkBlockIncompatible(c.EIP150Block, newcfg.EIP150Block, height) {
		return newBlockCompatError("EIP150 fork block", c.EIP150Block, newcfg.EIP150Block)
	}
	if isForkBlockIncompatible(c.EIP155Block, newcfg.EIP155Block, height) {
		return newBlockCompatError("EIP155 fork block", c.EIP155Block, newcfg.EIP155Block)
	}
	if isForkBlockIncompatible(c.EIP158Block, newcfg.EIP158Block, height) {
		return newBlockCompatError("EIP158 fork block", c.EIP158Block, newcfg.EIP158Block)
	}
	if c.IsEIP158(height) && !configBlockEqual(c.ChainID, newcfg.ChainID) {
		return newBlockCompatError("EIP158 chain ID", c.EIP158Block, newcfg.EIP158Block)
	}
	if isForkBlockIncompatible(c.ByzantiumBlock, newcfg.ByzantiumBlock, height) {
		return newBlockCompatError("Byzantium fork block", c.ByzantiumBlock, newcfg.ByzantiumBlock)
	}
	if isForkBlockIncompatible(c.ConstantinopleBlock, newcfg.ConstantinopleBlock, height) {
		return newBlockCompatError("Constantinople fork block", c.ConstantinopleBlock, newcfg.ConstantinopleBlock)
	}
	if isForkBlockIncompatible(c.PetersburgBlock, newcfg.PetersburgBlock, height) {
		// the only case where we allow Petersburg to be set in the past is if it is equal to Constantinople
		// mainly to satisfy fork ordering requirements which state that Petersburg fork be set if Constantinople fork is set
		if isForkBlockIncompatible(c.ConstantinopleBlock, newcfg.PetersburgBlock, height) {
			return newBlockCompatError("Petersburg fork block", c.PetersburgBlock, newcfg.PetersburgBlock)
		}
	}
	if isForkBlockIncompatible(c.IstanbulBlock, newcfg.IstanbulBlock, height) {
		return newBlockCompatError("Istanbul fork block", c.IstanbulBlock, newcfg.IstanbulBlock)
	}
	if isForkBlockIncompatible(c.MuirGlacierBlock, newcfg.MuirGlacierBlock, height) {
		return newBlockCompatError("Muir Glacier fork block", c.MuirGlacierBlock, newcfg.MuirGlacierBlock)
	}
	if isForkTimestampIncompatible(c.OdyPhase1BlockTimestamp, newcfg.OdyPhase1BlockTimestamp, time) {
		return newTimestampCompatError("OdyPhase1 fork block timestamp", c.OdyPhase1BlockTimestamp, newcfg.OdyPhase1BlockTimestamp)
	}
	if isForkTimestampIncompatible(c.OdyPhase2BlockTimestamp, newcfg.OdyPhase2BlockTimestamp, time) {
		return newTimestampCompatError("OdyPhase2 fork block timestamp", c.OdyPhase2BlockTimestamp, newcfg.OdyPhase2BlockTimestamp)
	}
	if isForkTimestampIncompatible(c.OdyPhase3BlockTimestamp, newcfg.OdyPhase3BlockTimestamp, time) {
		return newTimestampCompatError("OdyPhase3 fork block timestamp", c.OdyPhase3BlockTimestamp, newcfg.OdyPhase3BlockTimestamp)
	}
	if isForkTimestampIncompatible(c.OdyPhase4BlockTimestamp, newcfg.OdyPhase4BlockTimestamp, time) {
		return newTimestampCompatError("OdyPhase4 fork block timestamp", c.OdyPhase4BlockTimestamp, newcfg.OdyPhase4BlockTimestamp)
	}
	if isForkTimestampIncompatible(c.OdyPhase5BlockTimestamp, newcfg.OdyPhase5BlockTimestamp, time) {
		return newTimestampCompatError("OdyPhase5 fork block timestamp", c.OdyPhase5BlockTimestamp, newcfg.OdyPhase5BlockTimestamp)
	}
	if isForkTimestampIncompatible(c.OdyPhasePre6BlockTimestamp, newcfg.OdyPhasePre6BlockTimestamp, time) {
		return newTimestampCompatError("OdyPhasePre6 fork block timestamp", c.OdyPhasePre6BlockTimestamp, newcfg.OdyPhasePre6BlockTimestamp)
	}
	if isForkTimestampIncompatible(c.OdyPhase6BlockTimestamp, newcfg.OdyPhase6BlockTimestamp, time) {
		return newTimestampCompatError("OdyPhase6 fork block timestamp", c.OdyPhase6BlockTimestamp, newcfg.OdyPhase6BlockTimestamp)
	}
	if isForkTimestampIncompatible(c.OdyPhasePost6BlockTimestamp, newcfg.OdyPhasePost6BlockTimestamp, time) {
		return newTimestampCompatError("OdyPhasePost6 fork block timestamp", c.OdyPhasePost6BlockTimestamp, newcfg.OdyPhasePost6BlockTimestamp)
	}
	if isForkTimestampIncompatible(c.OdyPhase7BlockTimestamp, newcfg.OdyPhase7BlockTimestamp, time) {
		return newTimestampCompatError("OdyPhase7 fork block timestamp", c.OdyPhase7BlockTimestamp, newcfg.OdyPhase7BlockTimestamp)
	}
	if isForkTimestampIncompatible(c.BanffBlockTimestamp, newcfg.BanffBlockTimestamp, time) {
		return newTimestampCompatError("Banff fork block timestamp", c.BanffBlockTimestamp, newcfg.BanffBlockTimestamp)
	}
	if isForkTimestampIncompatible(c.CortinaBlockTimestamp, newcfg.CortinaBlockTimestamp, time) {
		return newTimestampCompatError("Cortina fork block timestamp", c.CortinaBlockTimestamp, newcfg.CortinaBlockTimestamp)
	}
	if isForkTimestampIncompatible(c.DUpgradeBlockTimestamp, newcfg.DUpgradeBlockTimestamp, time) {
		return newTimestampCompatError("DUpgrade fork block timestamp", c.DUpgradeBlockTimestamp, newcfg.DUpgradeBlockTimestamp)
	}
	if isForkTimestampIncompatible(c.CancunTime, newcfg.CancunTime, time) {
		return newTimestampCompatError("Cancun fork block timestamp", c.DUpgradeBlockTimestamp, newcfg.DUpgradeBlockTimestamp)
	}

	return nil
}

// isForkBlockIncompatible returns true if a fork scheduled at s1 cannot be rescheduled to
// block s2 because head is already past the fork.
func isForkBlockIncompatible(s1, s2, head *big.Int) bool {
	return (utils.IsBlockForked(s1, head) || utils.IsBlockForked(s2, head)) && !configBlockEqual(s1, s2)
}

func configBlockEqual(x, y *big.Int) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return x.Cmp(y) == 0
}

// isForkTimestampIncompatible returns true if a fork scheduled at timestamp s1
// cannot be rescheduled to timestamp s2 because head is already past the fork.
func isForkTimestampIncompatible(s1, s2 *uint64, head uint64) bool {
	return (utils.IsTimestampForked(s1, head) || utils.IsTimestampForked(s2, head)) && !configTimestampEqual(s1, s2)
}

func configTimestampEqual(x, y *uint64) bool {
	if x == nil {
		return y == nil
	}
	if y == nil {
		return x == nil
	}
	return *x == *y
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string

	// block numbers of the stored and new configurations if block based forking
	StoredBlock, NewBlock *big.Int

	// timestamps of the stored and new configurations if time based forking
	StoredTime, NewTime *uint64

	// the block number to which the local chain must be rewound to correct the error
	RewindToBlock uint64

	// the timestamp to which the local chain must be rewound to correct the error
	RewindToTime uint64
}

func newBlockCompatError(what string, storedblock, newblock *big.Int) *ConfigCompatError {
	var rew *big.Int
	switch {
	case storedblock == nil:
		rew = newblock
	case newblock == nil || storedblock.Cmp(newblock) < 0:
		rew = storedblock
	default:
		rew = newblock
	}
	err := &ConfigCompatError{
		What:          what,
		StoredBlock:   storedblock,
		NewBlock:      newblock,
		RewindToBlock: 0,
	}
	if rew != nil && rew.Sign() > 0 {
		err.RewindToBlock = rew.Uint64() - 1
	}
	return err
}

func newTimestampCompatError(what string, storedtime, newtime *uint64) *ConfigCompatError {
	var rew *uint64
	switch {
	case storedtime == nil:
		rew = newtime
	case newtime == nil || *storedtime < *newtime:
		rew = storedtime
	default:
		rew = newtime
	}
	err := &ConfigCompatError{
		What:         what,
		StoredTime:   storedtime,
		NewTime:      newtime,
		RewindToTime: 0,
	}
	if rew != nil && *rew > 0 {
		err.RewindToTime = *rew - 1
	}
	return err
}

func (err *ConfigCompatError) Error() string {
	if err.StoredBlock != nil {
		return fmt.Sprintf("mismatching %s in database (have block %d, want block %d, rewindto block %d)", err.What, err.StoredBlock, err.NewBlock, err.RewindToBlock)
	}
	return fmt.Sprintf("mismatching %s in database (have timestamp %d, want timestamp %d, rewindto timestamp %d)", err.What, err.StoredTime, err.NewTime, err.RewindToTime)
}

// Rules wraps ChainConfig and is merely syntactic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainID                                                 *big.Int
	IsHomestead, IsEIP150, IsEIP155, IsEIP158               bool
	IsByzantium, IsConstantinople, IsPetersburg, IsIstanbul bool
	IsCancun                                                bool

	// Rules for Odyssey releases
	IsOdyPhase1, IsOdyPhase2, IsOdyPhase3, IsOdyPhase4, IsOdyPhase5 					bool
	IsOdyPhasePre6, IsOdyPhase6, IsOdyPhasePost6, IsOdyPhase7                           bool
	IsBanff                                                                             bool
	IsCortina                                                                           bool
	IsDUpgrade                                                                          bool

	// Precompiles maps addresses to stateful precompiled contracts that are enabled
	// for this rule set.
	// Note: none of these addresses should conflict with the address space used by
	// any existing precompiles.
	Precompiles map[common.Address]precompile.StatefulPrecompiledContract
}

// Rules ensures c's ChainID is not nil.
func (c *ChainConfig) rules(num *big.Int, timestamp uint64) Rules {
	chainID := c.ChainID
	if chainID == nil {
		chainID = new(big.Int)
	}
	return Rules{
		ChainID:          new(big.Int).Set(chainID),
		IsHomestead:      c.IsHomestead(num),
		IsEIP150:         c.IsEIP150(num),
		IsEIP155:         c.IsEIP155(num),
		IsEIP158:         c.IsEIP158(num),
		IsByzantium:      c.IsByzantium(num),
		IsConstantinople: c.IsConstantinople(num),
		IsPetersburg:     c.IsPetersburg(num),
		IsIstanbul:       c.IsIstanbul(num),
		IsCancun:         c.IsCancun(timestamp),
	}
}

// OdysseyRules returns the Odyssey modified rules to support Odyssey
// network upgrades
func (c *ChainConfig) OdysseyRules(blockNum *big.Int, timestamp uint64) Rules {
	rules := c.rules(blockNum, timestamp)

	rules.IsOdyPhase1 = c.IsOdyPhase1(timestamp)
	rules.IsOdyPhase2 = c.IsOdyPhase2(timestamp)
	rules.IsOdyPhase3 = c.IsOdyPhase3(timestamp)
	rules.IsOdyPhase4 = c.IsOdyPhase4(timestamp)
	rules.IsOdyPhase5 = c.IsOdyPhase5(timestamp)
	rules.IsOdyPhasePre6 = c.IsOdyPhasePre6(timestamp)
	rules.IsOdyPhase6 = c.IsOdyPhase6(timestamp)
	rules.IsOdyPhasePost6 = c.IsOdyPhasePost6(timestamp)
	rules.IsOdyPhase7 = c.IsOdyPhase7(timestamp)
	rules.IsBanff = c.IsBanff(timestamp)
	rules.IsCortina = c.IsCortina(timestamp)
	rules.IsDUpgrade = c.IsDUpgrade(timestamp)

	// Initialize the stateful precompiles that should be enabled at [blockTimestamp].
	rules.Precompiles = make(map[common.Address]precompile.StatefulPrecompiledContract)
	for _, config := range c.enabledStatefulPrecompiles() {
		if utils.IsTimestampForked(config.Timestamp(), timestamp) {
			rules.Precompiles[config.Address()] = config.Contract()
		}
	}

	return rules
}

// enabledStatefulPrecompiles returns a list of stateful precompile configs in the order that they are enabled
// by block timestamp.
// Note: the return value does not include the native precompiles [nativeAssetCall] and [nativeAssetBalance].
// These are handled in [evm.precompile] directly.
func (c *ChainConfig) enabledStatefulPrecompiles() []precompile.StatefulPrecompileConfig {
	statefulPrecompileConfigs := make([]precompile.StatefulPrecompileConfig, 0)

	return statefulPrecompileConfigs
}

// CheckConfigurePrecompiles checks if any of the precompiles specified in the chain config are enabled by the block
// transition from [parentTimestamp] to the timestamp set in [blockContext]. If this is the case, it calls [Configure]
// to apply the necessary state transitions for the upgrade.
// This function is called:
// - within genesis setup to configure the starting state for precompiles enabled at genesis,
// - during block processing to update the state before processing the given block.
func (c *ChainConfig) CheckConfigurePrecompiles(parentTimestamp *uint64, blockContext precompile.BlockContext, statedb precompile.StateDB) {
	// Iterate the enabled stateful precompiles and configure them if needed
	for _, config := range c.enabledStatefulPrecompiles() {
		precompile.CheckConfigure(c, parentTimestamp, blockContext, config, statedb)
	}
}
