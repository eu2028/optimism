package multithreaded

import (
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum-optimism/optimism/cannon/mipsevm"
)

// Define stats interface
type StatsTracker interface {
	trackLL(step uint64)
	trackSCSuccess(step uint64)
	annotateDebugInfo(debugInfo *mipsevm.DebugInfo)
}

// Noop implementation for when debug is disabled
type noopStatsTracker struct{}

func (s *noopStatsTracker) trackSCSuccess(step uint64) {}

func (s *noopStatsTracker) annotateDebugInfo(debugInfo *mipsevm.DebugInfo) {}

func (s *noopStatsTracker) trackLL(step uint64) {}

var _ StatsTracker = (*noopStatsTracker)(nil)

// Actual implementation
type statsTrackerImpl struct {
	// State
	lastLLOpStep uint64
	// Stats
	maxStepsBetweenLLAndSC uint64
}

func (s *statsTrackerImpl) annotateDebugInfo(debugInfo *mipsevm.DebugInfo) {
	debugInfo.MaxStepsBetweenLLAndSC = hexutil.Uint64(s.maxStepsBetweenLLAndSC)
}

func (s *statsTrackerImpl) trackLL(step uint64) {
	s.lastLLOpStep = step
}

func (s *statsTrackerImpl) trackSCSuccess(step uint64) {
	diff := step - s.lastLLOpStep
	if diff > s.maxStepsBetweenLLAndSC {
		s.maxStepsBetweenLLAndSC = diff
	}
}

var _ StatsTracker = (*statsTrackerImpl)(nil)
