package mipsevm

import "github.com/ethereum/go-ethereum/common/hexutil"

type DebugInfo struct {
	Pages               int            `json:"pages"`
	MemoryUsed          hexutil.Uint64 `json:"memory_used"`
	NumPreimageRequests int            `json:"num_preimage_requests"`
	TotalPreimageSize   int            `json:"total_preimage_size"`
	//  Multithreading-related stats below
	RmwSuccessCount        int            `json:"rmw_success_count"`
	RmwFailCount           int            `json:"rmw_fail_count"`
	MaxStepsBetweenLLAndSC hexutil.Uint64 `json:"max_steps_between_ll_and_sc"`
}
