// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package delta

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/DioneProtocol/odysseygo/api"
	"github.com/DioneProtocol/odysseygo/ids"
	"github.com/DioneProtocol/odysseygo/utils/crypto/secp256k1"
	"github.com/DioneProtocol/odysseygo/utils/formatting"
	"github.com/DioneProtocol/odysseygo/utils/formatting/address"
	"github.com/DioneProtocol/odysseygo/utils/json"
	"github.com/DioneProtocol/odysseygo/utils/rpc"
)

// Interface compliance
var _ Client = (*client)(nil)

// Client interface for interacting with DELTA [chain]
type Client interface {
	IssueTx(ctx context.Context, txBytes []byte, options ...rpc.Option) (ids.ID, error)
	GetAtomicTxStatus(ctx context.Context, txID ids.ID, options ...rpc.Option) (Status, error)
	GetAtomicTx(ctx context.Context, txID ids.ID, options ...rpc.Option) ([]byte, error)
	GetAtomicUTXOs(ctx context.Context, addrs []ids.ShortID, sourceChain string, limit uint32, startAddress ids.ShortID, startUTXOID ids.ID, options ...rpc.Option) ([][]byte, ids.ShortID, ids.ID, error)
	ExportKey(ctx context.Context, userPass api.UserPass, addr common.Address, options ...rpc.Option) (*secp256k1.PrivateKey, string, error)
	ImportKey(ctx context.Context, userPass api.UserPass, privateKey *secp256k1.PrivateKey, options ...rpc.Option) (common.Address, error)
	Import(ctx context.Context, userPass api.UserPass, to common.Address, sourceChain string, options ...rpc.Option) (ids.ID, error)
	ExportDIONE(ctx context.Context, userPass api.UserPass, amount uint64, to ids.ShortID, targetChain string, options ...rpc.Option) (ids.ID, error)
	Export(ctx context.Context, userPass api.UserPass, amount uint64, to ids.ShortID, targetChain string, assetID string, options ...rpc.Option) (ids.ID, error)
	StartCPUProfiler(ctx context.Context, options ...rpc.Option) error
	StopCPUProfiler(ctx context.Context, options ...rpc.Option) error
	MemoryProfile(ctx context.Context, options ...rpc.Option) error
	LockProfile(ctx context.Context, options ...rpc.Option) error
	SetLogLevel(ctx context.Context, level log.Lvl, options ...rpc.Option) error
	GetVMConfig(ctx context.Context, options ...rpc.Option) (*Config, error)
}

// Client implementation for interacting with DELTA [chain]
type client struct {
	requester      rpc.EndpointRequester
	adminRequester rpc.EndpointRequester
}

// NewClient returns a Client for interacting with DELTA [chain]
func NewClient(uri, chain string) Client {
	return &client{
		requester:      rpc.NewEndpointRequester(fmt.Sprintf("%s/ext/bc/%s/dione", uri, chain)),
		adminRequester: rpc.NewEndpointRequester(fmt.Sprintf("%s/ext/bc/%s/admin", uri, chain)),
	}
}

// NewDChainClient returns a Client for interacting with the D Chain
func NewDChainClient(uri string) Client {
	return NewClient(uri, "D")
}

// IssueTx issues a transaction to a node and returns the TxID
func (c *client) IssueTx(ctx context.Context, txBytes []byte, options ...rpc.Option) (ids.ID, error) {
	res := &api.JSONTxID{}
	txStr, err := formatting.Encode(formatting.Hex, txBytes)
	if err != nil {
		return res.TxID, fmt.Errorf("problem hex encoding bytes: %w", err)
	}
	err = c.requester.SendRequest(ctx, "dione.issueTx", &api.FormattedTx{
		Tx:       txStr,
		Encoding: formatting.Hex,
	}, res, options...)
	return res.TxID, err
}

// GetAtomicTxStatus returns the status of [txID]
func (c *client) GetAtomicTxStatus(ctx context.Context, txID ids.ID, options ...rpc.Option) (Status, error) {
	res := &GetAtomicTxStatusReply{}
	err := c.requester.SendRequest(ctx, "dione.getAtomicTxStatus", &api.JSONTxID{
		TxID: txID,
	}, res, options...)
	return res.Status, err
}

// GetAtomicTx returns the byte representation of [txID]
func (c *client) GetAtomicTx(ctx context.Context, txID ids.ID, options ...rpc.Option) ([]byte, error) {
	res := &api.FormattedTx{}
	err := c.requester.SendRequest(ctx, "dione.getAtomicTx", &api.GetTxArgs{
		TxID:     txID,
		Encoding: formatting.Hex,
	}, res, options...)
	if err != nil {
		return nil, err
	}

	return formatting.Decode(formatting.Hex, res.Tx)
}

// GetAtomicUTXOs returns the byte representation of the atomic UTXOs controlled by [addresses]
// from [sourceChain]
func (c *client) GetAtomicUTXOs(ctx context.Context, addrs []ids.ShortID, sourceChain string, limit uint32, startAddress ids.ShortID, startUTXOID ids.ID, options ...rpc.Option) ([][]byte, ids.ShortID, ids.ID, error) {
	res := &api.GetUTXOsReply{}
	err := c.requester.SendRequest(ctx, "dione.getUTXOs", &api.GetUTXOsArgs{
		Addresses:   ids.ShortIDsToStrings(addrs),
		SourceChain: sourceChain,
		Limit:       json.Uint32(limit),
		StartIndex: api.Index{
			Address: startAddress.String(),
			UTXO:    startUTXOID.String(),
		},
		Encoding: formatting.Hex,
	}, res, options...)
	if err != nil {
		return nil, ids.ShortID{}, ids.Empty, err
	}

	utxos := make([][]byte, len(res.UTXOs))
	for i, utxo := range res.UTXOs {
		utxoBytes, err := formatting.Decode(res.Encoding, utxo)
		if err != nil {
			return nil, ids.ShortID{}, ids.Empty, err
		}
		utxos[i] = utxoBytes
	}
	endAddr, err := address.ParseToID(res.EndIndex.Address)
	if err != nil {
		return nil, ids.ShortID{}, ids.Empty, err
	}
	endUTXOID, err := ids.FromString(res.EndIndex.UTXO)
	return utxos, endAddr, endUTXOID, err
}

// ExportKey returns the private key corresponding to [addr] controlled by [user]
// in both Odyssey standard format and hex format
func (c *client) ExportKey(ctx context.Context, user api.UserPass, addr common.Address, options ...rpc.Option) (*secp256k1.PrivateKey, string, error) {
	res := &ExportKeyReply{}
	err := c.requester.SendRequest(ctx, "dione.exportKey", &ExportKeyArgs{
		UserPass: user,
		Address:  addr.Hex(),
	}, res, options...)
	return res.PrivateKey, res.PrivateKeyHex, err
}

// ImportKey imports [privateKey] to [user]
func (c *client) ImportKey(ctx context.Context, user api.UserPass, privateKey *secp256k1.PrivateKey, options ...rpc.Option) (common.Address, error) {
	res := &api.JSONAddress{}
	err := c.requester.SendRequest(ctx, "dione.importKey", &ImportKeyArgs{
		UserPass:   user,
		PrivateKey: privateKey,
	}, res, options...)
	if err != nil {
		return common.Address{}, err
	}
	return ParseEthAddress(res.Address)
}

// Import sends an import transaction to import funds from [sourceChain] and
// returns the ID of the newly created transaction
func (c *client) Import(ctx context.Context, user api.UserPass, to common.Address, sourceChain string, options ...rpc.Option) (ids.ID, error) {
	res := &api.JSONTxID{}
	err := c.requester.SendRequest(ctx, "dione.import", &ImportArgs{
		UserPass:    user,
		To:          to,
		SourceChain: sourceChain,
	}, res, options...)
	return res.TxID, err
}

// ExportDIONE sends DIONE from this chain to the address specified by [to].
// Returns the ID of the newly created atomic transaction
func (c *client) ExportDIONE(
	ctx context.Context,
	user api.UserPass,
	amount uint64,
	to ids.ShortID,
	targetChain string,
	options ...rpc.Option,
) (ids.ID, error) {
	return c.Export(ctx, user, amount, to, targetChain, "DIONE", options...)
}

// Export sends an asset from this chain to the O/D-Chain.
// After this tx is accepted, the DIONE must be imported to the O/D-chain with an importTx.
// Returns the ID of the newly created atomic transaction
func (c *client) Export(
	ctx context.Context,
	user api.UserPass,
	amount uint64,
	to ids.ShortID,
	targetChain string,
	assetID string,
	options ...rpc.Option,
) (ids.ID, error) {
	res := &api.JSONTxID{}
	err := c.requester.SendRequest(ctx, "dione.export", &ExportArgs{
		ExportDIONEArgs: ExportDIONEArgs{
			UserPass:    user,
			Amount:      json.Uint64(amount),
			TargetChain: targetChain,
			To:          to.String(),
		},
		AssetID: assetID,
	}, res, options...)
	return res.TxID, err
}

func (c *client) StartCPUProfiler(ctx context.Context, options ...rpc.Option) error {
	return c.adminRequester.SendRequest(ctx, "admin.startCPUProfiler", struct{}{}, &api.EmptyReply{}, options...)
}

func (c *client) StopCPUProfiler(ctx context.Context, options ...rpc.Option) error {
	return c.adminRequester.SendRequest(ctx, "admin.stopCPUProfiler", struct{}{}, &api.EmptyReply{}, options...)
}

func (c *client) MemoryProfile(ctx context.Context, options ...rpc.Option) error {
	return c.adminRequester.SendRequest(ctx, "admin.memoryProfile", struct{}{}, &api.EmptyReply{}, options...)
}

func (c *client) LockProfile(ctx context.Context, options ...rpc.Option) error {
	return c.adminRequester.SendRequest(ctx, "admin.lockProfile", struct{}{}, &api.EmptyReply{}, options...)
}

// SetLogLevel dynamically sets the log level for the D Chain
func (c *client) SetLogLevel(ctx context.Context, level log.Lvl, options ...rpc.Option) error {
	return c.adminRequester.SendRequest(ctx, "admin.setLogLevel", &SetLogLevelArgs{
		Level: level.String(),
	}, &api.EmptyReply{}, options...)
}

// GetVMConfig returns the current config of the VM
func (c *client) GetVMConfig(ctx context.Context, options ...rpc.Option) (*Config, error) {
	res := &ConfigReply{}
	err := c.adminRequester.SendRequest(ctx, "admin.getVMConfig", struct{}{}, res, options...)
	return res.Config, err
}