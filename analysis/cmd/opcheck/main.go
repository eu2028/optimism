// Package opcheck provides a code checker for Optimism.
//
// Compared to opvet, which is geared towards CI, opcheck is more appropriate
// for deverlopers, and can be used to apply suggested fixes to the codebase.
// It's also more granular and provides parsable output if needed.
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/ethereum-optimism/optimism/analysis/passes/testingt"
)

func main() {
	multichecker.Main(
		testingt.Analyzer,
	)
}
