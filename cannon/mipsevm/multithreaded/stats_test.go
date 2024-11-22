package multithreaded

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/cannon/mipsevm"
)

func TestStatsTracker(t *testing.T) {
	cases := []struct {
		name       string
		operations []Operation
		expected   *mipsevm.DebugInfo
	}{
		{
			name:       "Successful RMW operation",
			operations: []Operation{ll(3), scSuccess(13)},
			expected:   &mipsevm.DebugInfo{RmwSuccessCount: 1, MaxStepsBetweenLLAndSC: 10},
		},
		{
			name:       "Failed RMW operation",
			operations: []Operation{ll(3), scFail(13)},
			expected:   &mipsevm.DebugInfo{RmwFailCount: 1, MaxStepsBetweenLLAndSC: 10},
		},
		{
			name:       "Failed isolated sc op",
			operations: []Operation{scFail(13)},
			expected:   &mipsevm.DebugInfo{RmwFailCount: 1},
		},
		{
			name:       "Failed isolated sc op preceded by successful sc op",
			operations: []Operation{ll(1), scSuccess(10), scFail(23)},
			expected:   &mipsevm.DebugInfo{RmwSuccessCount: 1, RmwFailCount: 1, MaxStepsBetweenLLAndSC: 9},
		},
		{
			name:       "Multiple RMW operations",
			operations: []Operation{ll(1), scSuccess(2), ll(3), scFail(5), ll(6), scSuccess(16), ll(18), scSuccess(20), ll(21), scFail(30)},
			expected:   &mipsevm.DebugInfo{RmwSuccessCount: 3, RmwFailCount: 2, MaxStepsBetweenLLAndSC: 10},
		},
		{
			name:       "Interleaved RMW operations",
			operations: []Operation{ll(5), ll(10), scSuccess(15), scFail(25)},
			expected:   &mipsevm.DebugInfo{RmwSuccessCount: 1, RmwFailCount: 1, MaxStepsBetweenLLAndSC: 5},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stats := NewStatsTracker()
			for _, op := range c.operations {
				op(stats)
			}

			// Validate expectations
			actual := &mipsevm.DebugInfo{}
			stats.annotateDebugInfo(actual)
			require.Equal(t, c.expected, actual)
		})
	}
}

type Operation func(tracker StatsTracker)

func ll(step uint64) Operation {
	return func(tracker StatsTracker) {
		tracker.trackLL(step)
	}
}

func scSuccess(step uint64) Operation {
	return func(tracker StatsTracker) {
		tracker.trackSCSuccess(step)
	}
}

func scFail(step uint64) Operation {
	return func(tracker StatsTracker) {
		tracker.trackSCFailure(step)
	}
}
