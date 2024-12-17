package syncnode

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/ethereum-optimism/optimism/op-node/rollup/interop/managed"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

type RPCSyncNode struct {
	name string
	cl   client.RPC
}

func NewRPCSyncNode(name string, cl client.RPC) *RPCSyncNode {
	return &RPCSyncNode{
		name: name,
		cl:   cl,
	}
}

var _ SyncSource = (*RPCSyncNode)(nil)
var _ SyncControl = (*RPCSyncNode)(nil)
var _ SyncNode = (*RPCSyncNode)(nil)

func (rs *RPCSyncNode) BlockRefByNumber(ctx context.Context, number uint64) (eth.BlockRef, error) {
	var out *eth.BlockRef
	err := rs.cl.CallContext(ctx, &out, "interop_blockRefByNumber", number)
	if err != nil {
		var jsonErr rpc.Error
		if errors.As(err, &jsonErr) {
			if jsonErr.ErrorCode() == 0 { // TODO
				return eth.BlockRef{}, ethereum.NotFound
			}
		}
		return eth.BlockRef{}, err
	}
	return *out, nil
}

func (rs *RPCSyncNode) FetchReceipts(ctx context.Context, blockHash common.Hash) (gethtypes.Receipts, error) {
	var out gethtypes.Receipts
	err := rs.cl.CallContext(ctx, &out, "interop_fetchReceipts", blockHash)
	if err != nil {
		var jsonErr rpc.Error
		if errors.As(err, &jsonErr) {
			if jsonErr.ErrorCode() == 0 { // TODO
				return nil, ethereum.NotFound
			}
		}
		return nil, err
	}
	return out, nil
}

func (rs *RPCSyncNode) ChainID(ctx context.Context) (types.ChainID, error) {
	var chainID types.ChainID
	err := rs.cl.CallContext(ctx, &chainID, "interop_chainID")
	return chainID, err
}

func (rs *RPCSyncNode) String() string {
	return rs.name
}

func (rs *RPCSyncNode) SignalNextL1(ctx context.Context, l1Ref eth.BlockRef, l2Ref eth.BlockRef) (managed.DerivedPair, error) {
	var ret managed.DerivedPair
	err := rs.cl.CallContext(
		ctx,
		&ret,
		"interop_signalNextL1",
		l1Ref,
		l2Ref,
	)
	return ret, err
}

func (rs *RPCSyncNode) AnchorPoint(ctx context.Context) (managed.DerivedPair, error) {
	var ret managed.DerivedPair
	err := rs.cl.CallContext(
		ctx,
		&ret,
		"interop_anchorPoint",
	)
	return ret, err
}
