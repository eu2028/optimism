package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ethereum-optimism/optimism/kurtosis-devnet/pkg/kurtosis/sources/inspect"
)

func main() {
	// Parse command line flags
	enclave := flag.String("enclave", "", "Name of the Kurtosis enclave")
	flag.Parse()

	if *enclave == "" {
		fmt.Fprintln(os.Stderr, "Error: --enclave flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Run kurtosis inspect command
	cmd := exec.Command("kurtosis", "enclave", "inspect", *enclave)
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running kurtosis inspect: %v\n", err)
		os.Exit(1)
	}

	// Parse the inspect output
	result, err := inspect.ParseInspectOutput(strings.NewReader(string(output)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing inspect output: %v\n", err)
		os.Exit(1)
	}

	// Encode as JSON and write to stdout
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}
