package sources

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

type receiptsBatchCall = batching.IterativeBatchCall[common.Hash, *types.Receipt]

type BasicRPCReceiptsFetcher struct {
	client       rpcClient
	maxBatchSize int

	// calls caches uncompleted batch calls
	calls   map[receiptCacheKey]*receiptsBatchCall
	callsMu sync.Mutex
}

func NewBasicRPCReceiptsFetcher(client rpcClient, maxBatchSize int) *BasicRPCReceiptsFetcher {
	return &BasicRPCReceiptsFetcher{
		client:       client,
		maxBatchSize: maxBatchSize,
		calls:        make(map[receiptCacheKey]*receiptsBatchCall),
	}
}

// FetchReceipts fetches receipts for the given block and transaction hashes
// it does not validate receipts, and expects the caller to do so
func (f *BasicRPCReceiptsFetcher) FetchReceipts(ctx context.Context, blockInfo eth.BlockInfo, txHashes []common.Hash) (types.Receipts, error) {
	ret, err := f.FetchReceiptsRange(ctx, []blockTxHashes{{blockInfo, txHashes}})
	if err != nil {
		return nil, err
	}
	return ret[0].receipts, nil
}

func (f *BasicRPCReceiptsFetcher) FetchReceiptsRange(ctx context.Context, blocks []blockTxHashes) ([]blockReceipts, error) {
	key := newCacheKey(blocks)
	// merge all requested tx hashes into one batch call
	allTx := []common.Hash{}
	for _, b := range blocks {
		allTx = append(allTx, b.txHashes...)
	}
	call := f.getOrCreateBatchCall(key, allTx)

	// Fetch all receipts
	for {
		if err := call.Fetch(ctx); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}
	res, err := call.Result()
	if err != nil {
		return nil, err
	}
	// call successful, remove from cache
	f.deleteBatchCall(key)

	// single block, no need to sort results
	if len(blocks) == 1 {
		return []blockReceipts{{blocks[0].block, res}}, nil
	}

	// split the single batch result into block results
	blockRes := make([]blockReceipts, len(blocks))
	// also maintain pointers to each item so we can refer to them while sorting
	blockResMap := make(map[common.Hash]*blockReceipts, len(blocks))
	for i, b := range blocks {
		blockRes[i] = blockReceipts{
			block:    b.block,
			receipts: make(types.Receipts, len(b.txHashes))}
		blockResMap[b.block.Hash()] = &blockRes[i]
	}
	for _, r := range res {
		if _, ok := blockResMap[r.BlockHash]; !ok {
			return nil, fmt.Errorf("unexpected receipt for block %s", r.BlockHash)
		}
		blockResMap[r.BlockHash].receipts = append(blockResMap[r.BlockHash].receipts, r)
	}

	return blockRes, nil
}

func (f *BasicRPCReceiptsFetcher) getOrCreateBatchCall(key receiptCacheKey, txHashes []common.Hash) *receiptsBatchCall {
	f.callsMu.Lock()
	defer f.callsMu.Unlock()
	if call, ok := f.calls[key]; ok {
		return call
	}
	call := batching.NewIterativeBatchCall[common.Hash, *types.Receipt](
		txHashes,
		makeReceiptRequest,
		f.client.BatchCallContext,
		f.client.CallContext,
		f.maxBatchSize,
	)
	f.calls[key] = call
	return call
}

func (f *BasicRPCReceiptsFetcher) deleteBatchCall(key receiptCacheKey) {
	f.callsMu.Lock()
	defer f.callsMu.Unlock()
	delete(f.calls, key)
}

func makeReceiptRequest(txHash common.Hash) (*types.Receipt, rpc.BatchElem) {
	out := new(types.Receipt)
	return out, rpc.BatchElem{
		Method: "eth_getTransactionReceipt",
		Args:   []any{txHash},
		Result: &out, // receipt may become nil, double pointer is intentional
	}
}
