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

// A CachingReceiptsProvider caches successful receipt fetches from the inner
// ReceiptsProvider. It also avoids duplicate in-flight requests per block hash.
type CachingReceiptsProvider struct {
	inner ReceiptsProvider
	cache *caching.LRUCache[common.Hash, types.Receipts]

	// lock fetching process for each block hash to avoid duplicate requests
	fetching       map[common.Hash]*sync.Mutex
	fetchingMu     sync.Mutex // only protects map
	batchLockingMu sync.Mutex // ensures only one group of locks are acquired at once
}

func NewCachingReceiptsProvider(inner ReceiptsProvider, m caching.Metrics, cacheSize int) *CachingReceiptsProvider {
	return &CachingReceiptsProvider{
		inner:    inner,
		cache:    caching.NewLRUCache[common.Hash, types.Receipts](m, "receipts", cacheSize),
		fetching: make(map[common.Hash]*sync.Mutex),
	}
}

func NewCachingRPCReceiptsProvider(client rpcClient, log log.Logger, config RPCReceiptsConfig, m caching.Metrics, cacheSize int) *CachingReceiptsProvider {
	return NewCachingReceiptsProvider(NewRPCReceiptsFetcher(client, log, config), m, cacheSize)
}

// FetchReceipts fetches receipts for the given block and transaction hashes
// it expects that the inner FetchReceipts implementation handles validation
func (p *CachingReceiptsProvider) FetchReceipts(ctx context.Context, blockInfo eth.BlockInfo, txHashes []common.Hash) (types.Receipts, error) {
	block := eth.ToBlockID(blockInfo)
	if r, ok := p.cache.Get(block.Hash); ok {
		return r, nil
	}

	mu := p.getOrCreateFetchingLock(block.Hash)
	mu.Lock()
	defer mu.Unlock()
	// Other routine might have fetched in the meantime
	if r, ok := p.cache.Get(block.Hash); ok {
		// we might have created a new lock above while the old
		// fetching job completed.
		p.deleteFetchingLock(block.Hash)
		return r, nil
	}

	r, err := p.inner.FetchReceipts(ctx, blockInfo, txHashes)
	if err != nil {
		return nil, err
	}

	p.cache.Add(block.Hash, r)
	// result now in cache, can delete fetching lock
	p.deleteFetchingLock(block.Hash)
	return r, nil
}

// BatchFetchReceipts fetches receipts for the given blocks and transaction hashes
// it expects that the inner BatchFetchReceipts implementation handles validation
// it functions by scanning the cache for any parts of the batch that are already cached
// and then only fetching the parts that are not cached.
// it takes a lock for each block that is not immediately cached to avoid duplicate requests, and then
// double-checks the cache after acquiring the lock to avoid duplicate requests.
func (p *CachingReceiptsProvider) BatchFetchReceipts(ctx context.Context, blockInfos []eth.BlockInfo, txHashes [][]common.Hash) ([]types.Receipts, error) {
	blockHashes := make([]common.Hash, len(blockInfos))
	results := make([]types.Receipts, len(blockInfos))
	innerBlockInfos := []eth.BlockInfo{}
	innerTxHashes := [][]common.Hash{}
	innerResultIndex := []int{}
	for i := range blockInfos {
		if r, ok := p.cache.Get(blockInfos[i].Hash()); ok {
			results[i] = r
			continue
		}
		// record information which be passed to the inner provider
		// or used for caching or result loading
		// a second check for the cache will be done after acquiring the lock
		blockHashes[i] = eth.ToBlockID(blockInfos[i]).Hash
		innerBlockInfos = append(innerBlockInfos, blockInfos[i])
		innerTxHashes = append(innerTxHashes, txHashes[i])
		innerResultIndex = append(innerResultIndex, i)
	}
	// the entire batch could be constructed from the cache, return early
	if len(innerBlockInfos) == 0 {
		return results, nil
	}

	// create fetching locks for the missing blocks
	locks := p.getOrCreateFetchingLocks(blockHashes)
	unlock := p.takeBatchLocks(locks)
	defer unlock()

	// it is possible other routines fetched the results while we were waiting for the locks
	// so do a second check for the results in the cache, now that we are locked
	finalBlockHashes := []common.Hash{}
	finalInnerBlockInfos := []eth.BlockInfo{}
	finalInnerTxHashes := [][]common.Hash{}
	finalInnerResultIndex := []int{}
	// one final chance to get the result from the cache
	// and potentially reserve it from the batch request
	for i := range innerBlockInfos {
		if r, ok := p.cache.Get(blockHashes[i]); ok {
			results[innerResultIndex[i]] = r
			p.deleteFetchingLock(blockHashes[i])
		} else {
			// record final information for the inner provider
			finalBlockHashes = append(finalBlockHashes, blockHashes[i])
			finalInnerBlockInfos = append(finalInnerBlockInfos, innerBlockInfos[i])
			finalInnerTxHashes = append(finalInnerTxHashes, innerTxHashes[i])
			finalInnerResultIndex = append(finalInnerResultIndex, innerResultIndex[i])
		}
	}
	// if there is no more work after the second check, return early
	if len(finalInnerBlockInfos) == 0 {
		return results, nil
	}

	newResults, err := p.inner.BatchFetchReceipts(ctx, finalInnerBlockInfos, finalInnerTxHashes)
	if err != nil {
		return nil, err
	}

	// save all the new results to the cache and to the results
	for i := range newResults {
		// save the result to the cache
		p.cache.Add(finalBlockHashes[i], newResults[i])
		// save the result to the outer results
		results[finalInnerResultIndex[i]] = newResults[i]
	}
	p.deleteFetchingLocks(finalBlockHashes)
	return results, nil
}

func (p *CachingReceiptsProvider) isInnerNil() bool {
	return p.inner == nil
}

func (p *CachingReceiptsProvider) getOrCreateFetchingLock(blockHash common.Hash) *sync.Mutex {
	p.fetchingMu.Lock()
	defer p.fetchingMu.Unlock()
	if mu, ok := p.fetching[blockHash]; ok {
		return mu
	}
	mu := new(sync.Mutex)
	p.fetching[blockHash] = mu
	return mu
}

func (p *CachingReceiptsProvider) getOrCreateFetchingLocks(blockHashes []common.Hash) []*sync.Mutex {
	p.fetchingMu.Lock()
	defer p.fetchingMu.Unlock()
	locks := make([]*sync.Mutex, len(blockHashes))
	for i, blockHash := range blockHashes {
		if mu, ok := p.fetching[blockHash]; ok {
			locks[i] = mu
		} else {
			mu := new(sync.Mutex)
			p.fetching[blockHash] = mu
			locks[i] = mu
		}
	}
	return locks
}

// takeBatchLocks serializes batch lock-taking to prevent deadlocks
// it returns a function to release all the locks at once
func (p *CachingReceiptsProvider) takeBatchLocks(locks []*sync.Mutex) func() {
	p.batchLockingMu.Lock()
	defer p.batchLockingMu.Unlock()

	for _, lock := range locks {
		lock.Lock()
	}
	return func() {
		for _, lock := range locks {
			lock.Unlock()
		}
	}
}

func (p *CachingReceiptsProvider) deleteFetchingLocks(blockHashes []common.Hash) {
	p.fetchingMu.Lock()
	defer p.fetchingMu.Unlock()
	for _, blockHash := range blockHashes {
		delete(p.fetching, blockHash)
	}
}

func (p *CachingReceiptsProvider) deleteFetchingLock(blockHash common.Hash) {
	p.fetchingMu.Lock()
	defer p.fetchingMu.Unlock()
	delete(p.fetching, blockHash)
}
