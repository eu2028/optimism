package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ethereum-optimism/optimism/kurtosis-devnet/pkg/kurtosis/sources/deployer"
)

func main() {
	// Parse command line flags
	enclave := flag.String("enclave", "", "Kurtosis enclave name")
	flag.Parse()

	if *enclave == "" {
		fmt.Fprintln(os.Stderr, "Error: enclave parameter is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse the deployer state
	state, err := deployer.ParseDeployerData(*enclave)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing deployer state: %v\n", err)
		os.Exit(1)
	}

	// Convert to JSON and print
	output, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}
