package providers

import (
	"fmt"
	"reflect"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/providers/e2e_backends"
)

func provideE2E[T any](t interfaces.Test, spec interfaces.TestSpec[T]) (T, error) {
	typ := reflect.TypeFor[T]()

	switch typ {
	case reflect.TypeFor[interfaces.SuperSystem]():
		// TODO: yikes, fix it please, this is a mess
		spec := spec.(interfaces.TestSpec[interfaces.SuperSystem])
		switch spec := spec.(type) {
		case *interfaces.SuperSystemSpec:
			s, err := e2e_backends.NewSpecifiedSuperSystem(t, spec)
			return s.(T), err
		case *interfaces.EmptyTestSpec[interfaces.SuperSystem]:
			s, err := e2e_backends.NewSpecifiedSuperSystem(t, &interfaces.SuperSystemSpec{})
			return s.(T), err
		default:
			var void T
			return void, fmt.Errorf("unsupported test spec type: %T", spec)
		}
	}

	var void T
	return void, fmt.Errorf("unsupported test type: %s", typ.String())
}
