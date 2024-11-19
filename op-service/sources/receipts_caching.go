package sources

import (
	"context"
	"sync"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources/caching"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
)

type receiptCacheKey string

// newCacheKey creates a cache key from a list of blockTxHashes.
// it concatenates the block hashes with a separator ","
func newCacheKey(blocks []blockTxHashes) receiptCacheKey {
	ret := ""
	for i, b := range blocks {
		id := eth.ToBlockID(b.block)
		ret += id.Hash.String()
		// if there is more to come, append a separator
		if i < len(blocks)-1 {
			ret += ","
		}
	}
	return receiptCacheKey(ret)
}

// A CachingReceiptsProvider caches successful receipt fetches from the inner
// ReceiptsProvider. It also avoids duplicate in-flight requests per block hash.
type CachingReceiptsProvider struct {
	inner ReceiptsProvider
	cache *caching.LRUCache[receiptCacheKey, []blockReceipts]

	// lock fetching process for each block hash to avoid duplicate requests
	fetching   map[receiptCacheKey]*sync.Mutex
	fetchingMu sync.Mutex // only protects map
}

func NewCachingReceiptsProvider(inner ReceiptsProvider, m caching.Metrics, cacheSize int) *CachingReceiptsProvider {
	return &CachingReceiptsProvider{
		inner:    inner,
		cache:    caching.NewLRUCache[receiptCacheKey, []blockReceipts](m, "receipts", cacheSize),
		fetching: make(map[receiptCacheKey]*sync.Mutex),
	}
}

func NewCachingRPCReceiptsProvider(client rpcClient, log log.Logger, config RPCReceiptsConfig, m caching.Metrics, cacheSize int) *CachingReceiptsProvider {
	return NewCachingReceiptsProvider(NewRPCReceiptsFetcher(client, log, config), m, cacheSize)
}

func (p *CachingReceiptsProvider) getOrCreateFetchingLock(key receiptCacheKey) *sync.Mutex {
	p.fetchingMu.Lock()
	defer p.fetchingMu.Unlock()
	if mu, ok := p.fetching[key]; ok {
		return mu
	}
	mu := new(sync.Mutex)
	p.fetching[key] = mu
	return mu
}

func (p *CachingReceiptsProvider) deleteFetchingLock(key receiptCacheKey) {
	p.fetchingMu.Lock()
	defer p.fetchingMu.Unlock()
	delete(p.fetching, key)
}

// FetchReceipts fetches receipts for the given block and transaction hashes
// it uses FetchReceiptsRange internally, and unwraps the single result
func (p *CachingReceiptsProvider) FetchReceipts(ctx context.Context, blockInfo eth.BlockInfo, txHashes []common.Hash) (types.Receipts, error) {
	// wrap the single request into a list for the range fetch
	res, err := p.FetchReceiptsRange(ctx, []blockTxHashes{{blockInfo, txHashes}})
	if err != nil {
		return nil, err
	}
	// this is a single result, so unwrap it
	return res[0].receipts, nil
}

// FetchReceiptsRange fetches receipts for the given blocks and transaction hashes per block
// it expects that the inner FetchReceiptsRange implementation handles validation
func (p *CachingReceiptsProvider) FetchReceiptsRange(ctx context.Context, blockInfos []blockTxHashes) ([]blockReceipts, error) {
	key := newCacheKey(blockInfos)
	if r, ok := p.cache.Get(key); ok {
		return r, nil
	}

	mu := p.getOrCreateFetchingLock(key)
	mu.Lock()
	defer mu.Unlock()
	// Other routine might have fetched in the meantime
	if r, ok := p.cache.Get(key); ok {
		// we might have created a new lock above while the old
		// fetching job completed.
		p.deleteFetchingLock(key)
		return r, nil
	}

	// call the inner provider
	r, err := p.inner.FetchReceiptsRange(ctx, blockInfos)
	if err != nil {
		return nil, err
	}

	p.cache.Add(key, r)
	// result now in cache, can delete fetching lock
	p.deleteFetchingLock(key)
	return r, nil
}

func (p *CachingReceiptsProvider) isInnerNil() bool {
	return p.inner == nil
}
