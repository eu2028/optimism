// testdata/src/b/b.go
package b

import "testing"

type tester struct {
	t *testing.T // want "avoid using \\*testing.T directly"
}

func code(t *testing.T) { // want "avoid using \\*testing.T directly"
	t.Fail()
}

func code2(t *testing.T) { // want "avoid using \\*testing.T directly" code2:"TestingTBUnsafe"
	helper(t)
}

func helper(t *testing.T) { // want "avoid using \\*testing.T directly" helper:"TestingTBUnsafe"
	t.Parallel()
}
