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
// Copyright 2014 The go-ethereum Authors
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

package vm

import (
	"github.com/DioneProtocol/coreth/vmerrs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

var (
	BuiltinAddr = common.Address{
		1, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
)

// Config are the configuration options for the Interpreter
type Config struct {
	Tracer                  DELTALogger // Opcode logger
	NoBaseFee               bool        // Forces the EIP-1559 baseFee to 0 (needed for 0 price calls)
	EnablePreimageRecording bool        // Enables recording of SHA3/keccak preimages
	ExtraEips               []int       // Additional EIPS that are to be enabled

	// AllowUnfinalizedQueries allow unfinalized queries
	AllowUnfinalizedQueries bool
}

// ScopeContext contains the things that are per-call, such as stack and memory,
// but not transients like pc and gas
type ScopeContext struct {
	Memory   *Memory
	Stack    *Stack
	Contract *Contract
}

// DELTAInterpreter represents an DELTA interpreter
type DELTAInterpreter struct {
	delta *DELTA
	table *JumpTable

	hasher    crypto.KeccakState // Keccak256 hasher instance shared across opcodes
	hasherBuf common.Hash        // Keccak256 hasher result array shared aross opcodes

	readOnly   bool   // Whether to throw on stateful modifications
	returnData []byte // Last CALL's return data for subsequent reuse
}

// NewDELTAInterpreter returns a new instance of the Interpreter.
func NewDELTAInterpreter(delta *DELTA) *DELTAInterpreter {
	// If jump table was not initialised we set the default one.
	var table *JumpTable
	switch {
	case delta.chainRules.IsDUpgrade:
		table = &dUpgradeInstructionSet
	case delta.chainRules.IsApricotPhase3:
		table = &apricotPhase3InstructionSet
	case delta.chainRules.IsApricotPhase2:
		table = &apricotPhase2InstructionSet
	case delta.chainRules.IsApricotPhase1:
		table = &apricotPhase1InstructionSet
	case delta.chainRules.IsIstanbul:
		table = &istanbulInstructionSet
	case delta.chainRules.IsConstantinople:
		table = &constantinopleInstructionSet
	case delta.chainRules.IsByzantium:
		table = &byzantiumInstructionSet
	case delta.chainRules.IsEIP158:
		table = &spuriousDragonInstructionSet
	case delta.chainRules.IsEIP150:
		table = &tangerineWhistleInstructionSet
	case delta.chainRules.IsHomestead:
		table = &homesteadInstructionSet
	default:
		table = &frontierInstructionSet
	}
	var extraEips []int
	if len(delta.Config.ExtraEips) > 0 {
		// Deep-copy jumptable to prevent modification of opcodes in other tables
		table = copyJumpTable(table)
	}
	for _, eip := range delta.Config.ExtraEips {
		if err := EnableEIP(eip, table); err != nil {
			// Disable it, so caller can check if it's activated or not
			log.Error("EIP activation failed", "eip", eip, "error", err)
		} else {
			extraEips = append(extraEips, eip)
		}
	}
	delta.Config.ExtraEips = extraEips
	return &DELTAInterpreter{delta: delta, table: table}
}

// Run loops and evaluates the contract's code with the given input data and returns
// the return byte-slice and an error if one occurred.
//
// It's important to note that any errors returned by the interpreter should be
// considered a revert-and-consume-all-gas operation except for
// ErrExecutionReverted which means revert-and-keep-gas-left.
func (in *DELTAInterpreter) Run(contract *Contract, input []byte, readOnly bool) (ret []byte, err error) {
	// Deprecate special handling of [BuiltinAddr] as of ApricotPhase2.
	// In ApricotPhase2, the contract deployed in the genesis is overridden by a deprecated precompiled
	// contract which will return an error immediately if its ever called. Therefore, this function should
	// never be called after ApricotPhase2 with [BuiltinAddr] as the contract address.
	if !in.delta.chainRules.IsApricotPhase2 && contract.Address() == BuiltinAddr {
		self := AccountRef(contract.Caller())
		if _, ok := contract.caller.(*Contract); ok {
			contract = contract.AsDelegate()
		}
		contract.self = self
	}

	// Increment the call depth which is restricted to 1024
	in.delta.depth++
	defer func() { in.delta.depth-- }()

	// Make sure the readOnly is only set if we aren't in readOnly yet.
	// This also makes sure that the readOnly flag isn't removed for child calls.
	if readOnly && !in.readOnly {
		in.readOnly = true
		defer func() { in.readOnly = false }()
	}

	// Reset the previous call's return data. It's unimportant to preserve the old buffer
	// as every returning call will return new data anyway.
	in.returnData = nil

	// Don't bother with the execution if there's no code.
	// Note: this avoids invoking the tracer in any way for simple value
	// transfers to EOA accounts.
	if len(contract.Code) == 0 {
		return nil, nil
	}

	var (
		op          OpCode        // current opcode
		mem         = NewMemory() // bound memory
		stack       = newstack()  // local stack
		callContext = &ScopeContext{
			Memory:   mem,
			Stack:    stack,
			Contract: contract,
		}
		// For optimisation reason we're using uint64 as the program counter.
		// It's theoretically possible to go above 2^64. The YP defines the PC
		// to be uint256. Practically much less so feasible.
		pc   = uint64(0) // program counter
		cost uint64
		// copies used by tracer
		pcCopy  uint64 // needed for the deferred DELTALogger
		gasCopy uint64 // for DELTALogger to log gas remaining before execution
		logged  bool   // deferred DELTALogger should ignore already logged steps
		res     []byte // result of the opcode execution function
		debug   = in.delta.Config.Tracer != nil
	)

	// Don't move this deferred function, it's placed before the capturestate-deferred method,
	// so that it get's executed _after_: the capturestate needs the stacks before
	// they are returned to the pools
	defer func() {
		returnStack(stack)
	}()
	contract.Input = input

	if debug {
		defer func() {
			if err != nil {
				if !logged {
					in.delta.Config.Tracer.CaptureState(pcCopy, op, gasCopy, cost, callContext, in.returnData, in.delta.depth, err)
				} else {
					in.delta.Config.Tracer.CaptureFault(pcCopy, op, gasCopy, cost, callContext, in.delta.depth, err)
				}
			}
		}()
	}
	// The Interpreter main run loop (contextual). This loop runs until either an
	// explicit STOP, RETURN or SELFDESTRUCT is executed, an error occurred during
	// the execution of one of the operations or until the done flag is set by the
	// parent context.
	for {
		if debug {
			// Capture pre-execution values for tracing.
			logged, pcCopy, gasCopy = false, pc, contract.Gas
		}
		// Get the operation from the jump table and validate the stack to ensure there are
		// enough stack items available to perform the operation.
		op = contract.GetOp(pc)
		operation := in.table[op]
		cost = operation.constantGas // For tracing
		// Validate stack
		if sLen := stack.len(); sLen < operation.minStack {
			return nil, &ErrStackUnderflow{stackLen: sLen, required: operation.minStack}
		} else if sLen > operation.maxStack {
			return nil, &ErrStackOverflow{stackLen: sLen, limit: operation.maxStack}
		}
		if !contract.UseGas(cost) {
			return nil, vmerrs.ErrOutOfGas
		}

		if operation.dynamicGas != nil {
			// All ops with a dynamic memory usage also has a dynamic gas cost.
			var memorySize uint64
			// calculate the new memory size and expand the memory to fit
			// the operation
			// Memory check needs to be done prior to evaluating the dynamic gas portion,
			// to detect calculation overflows
			if operation.memorySize != nil {
				memSize, overflow := operation.memorySize(stack)
				if overflow {
					return nil, vmerrs.ErrGasUintOverflow
				}
				// memory is expanded in words of 32 bytes. Gas
				// is also calculated in words.
				if memorySize, overflow = math.SafeMul(toWordSize(memSize), 32); overflow {
					return nil, vmerrs.ErrGasUintOverflow
				}
			}
			// Consume the gas and return an error if not enough gas is available.
			// cost is explicitly set so that the capture state defer method can get the proper cost
			var dynamicCost uint64
			dynamicCost, err = operation.dynamicGas(in.delta, contract, stack, mem, memorySize)
			cost += dynamicCost // for tracing
			if err != nil || !contract.UseGas(dynamicCost) {
				return nil, vmerrs.ErrOutOfGas
			}
			// Do tracing before memory expansion
			if debug {
				in.delta.Config.Tracer.CaptureState(pc, op, gasCopy, cost, callContext, in.returnData, in.delta.depth, err)
				logged = true
			}
			if memorySize > 0 {
				mem.Resize(memorySize)
			}
		} else if debug {
			in.delta.Config.Tracer.CaptureState(pc, op, gasCopy, cost, callContext, in.returnData, in.delta.depth, err)
			logged = true
		}

		// execute the operation
		res, err = operation.execute(&pc, in, callContext)
		if err != nil {
			break
		}
		pc++
	}
	if err == errStopToken {
		err = nil // clear stop token error
	}

	return res, err
}
