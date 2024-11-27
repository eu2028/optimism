package providers

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
)

type BackendType int

const (
	E2E BackendType = iota
)

// TODO: make this configurable
var usedBackend BackendType = E2E

func Provide[T any](t interfaces.Test, spec interfaces.TestSpec[T]) (T, error) {
	t.Helper()

	switch usedBackend {
	case E2E:
		return provideE2E[T](t, spec)
	}

	var void T
	return void, fmt.Errorf("unsupported backend type: %d", usedBackend)
}
