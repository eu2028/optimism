package syncnode

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-node/rollup/interop/managed"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

type SyncNodeCollection interface {
	Load(ctx context.Context, logger log.Logger) ([]SyncNodeSetup, error)
	Check() error
}

type SyncNodeSetup interface {
	Setup(ctx context.Context, logger log.Logger) (SyncNode, error)
}

type SyncSource interface {
	BlockRefByNumber(ctx context.Context, number uint64) (eth.BlockRef, error)
	FetchReceipts(ctx context.Context, blockHash common.Hash) (gethtypes.Receipts, error)
	ChainID(ctx context.Context) (types.ChainID, error)
	// String identifies the sync source
	String() string
}

type SyncControl interface {
	SignalNextL1(ctx context.Context, l1Ref eth.BlockRef, l2Ref eth.BlockRef) (managed.DerivedPair, error)
	AnchorPoint(ctx context.Context) (managed.DerivedPair, error)
}

type SyncNode interface {
	SyncSource
	SyncControl
}
