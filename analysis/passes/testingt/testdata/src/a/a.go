// testdata/src/a/a.go
package a

import "testing"

func Code(*testing.T) {} // want "avoid using \\*testing.T directly"

func Test(t *testing.T) { // want "avoid using \\*testing.T directly"
	t.Fail()
}
