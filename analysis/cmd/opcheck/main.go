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
