package multithreaded

import (
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/ethereum-optimism/optimism/cannon/mipsevm"
)

// Define stats interface
type StatsTracker interface {
	trackLL(step uint64)
	trackSCSuccess(step uint64)
	trackSCFailure(step uint64)
	annotateDebugInfo(debugInfo *mipsevm.DebugInfo)
}

// Noop implementation for when debug is disabled
type noopStatsTracker struct{}

func NoopStatsTracker() StatsTracker {
	return &noopStatsTracker{}
}

func (s *noopStatsTracker) trackLL(step uint64)                            {}
func (s *noopStatsTracker) trackSCSuccess(step uint64)                     {}
func (s *noopStatsTracker) trackSCFailure(step uint64)                     {}
func (s *noopStatsTracker) annotateDebugInfo(debugInfo *mipsevm.DebugInfo) {}

var _ StatsTracker = (*noopStatsTracker)(nil)

// Actual implementation
type statsTrackerImpl struct {
	// State
	lastLLOpStep uint64
	// Stats
	rmwSuccessCount int
	rmwFailCount    int
	// Note: Once a new LL operation is executed, we reset lastLLOpStep, losing track of previous RMW operations.
	// So, maxStepsBetweenLLAndSC is not complete and may miss longer ranges for failed rmw sequences.
	maxStepsBetweenLLAndSC uint64
}

func NewStatsTracker() StatsTracker {
	return &statsTrackerImpl{}
}

func (s *statsTrackerImpl) annotateDebugInfo(debugInfo *mipsevm.DebugInfo) {
	debugInfo.RmwSuccessCount = s.rmwSuccessCount
	debugInfo.RmwFailCount = s.rmwFailCount
	debugInfo.MaxStepsBetweenLLAndSC = hexutil.Uint64(s.maxStepsBetweenLLAndSC)
}

func (s *statsTrackerImpl) trackLL(step uint64) {
	s.lastLLOpStep = step
}

func (s *statsTrackerImpl) trackSCSuccess(step uint64) {
	s.rmwSuccessCount += 1
	diff := step - s.lastLLOpStep
	if diff > s.maxStepsBetweenLLAndSC {
		s.maxStepsBetweenLLAndSC = diff
	}
	// Reset ll op state
	s.lastLLOpStep = 0
}

func (s *statsTrackerImpl) trackSCFailure(step uint64) {
	s.rmwFailCount += 1

	diff := step - s.lastLLOpStep
	if s.lastLLOpStep > 0 && diff > s.maxStepsBetweenLLAndSC {
		s.maxStepsBetweenLLAndSC = diff
	}
}

var _ StatsTracker = (*statsTrackerImpl)(nil)
