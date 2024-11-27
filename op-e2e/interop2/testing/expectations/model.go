package expectations

import (
	"context"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
	"github.com/stretchr/testify/require"

	gethCore "github.com/ethereum/go-ethereum/core"
)

type BehaviorModel struct {
	InvalidPayloadExpectedError        error
	InvalidPayloadExecutionExpectation func(context.Context, interfaces.Test, error)
	NoError                            func(context.Context, interfaces.Test, error)
}

func GetBehaviorModel(mempoolFiltering bool) *BehaviorModel {
	model := &BehaviorModel{
		NoError: func(ctx context.Context, t interfaces.Test, err error) {
			require.NoError(t, err)
		},
	}

	if mempoolFiltering {
		model.InvalidPayloadExpectedError = gethCore.ErrTxFilteredOut
	} else {
		model.InvalidPayloadExpectedError = nil
	}

	model.InvalidPayloadExecutionExpectation = func(ctx context.Context, t interfaces.Test, err error) {
		if mempoolFiltering {
			require.ErrorContains(t, err, gethCore.ErrTxFilteredOut.Error())
		} else {
			require.ErrorIs(t, err, ctx.Err())
			require.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
		}
	}
	return model
}
