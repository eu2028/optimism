package kurtosis

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DeploymentAddresses maps contract names to their addresses
type DeploymentAddresses map[string]string

// DeploymentStateAddresses maps chain IDs to their contract addresses
type DeploymentStateAddresses map[string]DeploymentAddresses

// StateFile represents the structure of the state.json file
type StateFile struct {
	OpChainDeployments []map[string]interface{} `json:"opChainDeployments"`
}

// downloadArtifact downloads a kurtosis artifact to a temporary directory
func downloadArtifact(enclave, artifact, destDir string) error {
	cmd := exec.Command("kurtosis", "files", "download", enclave, artifact, destDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download artifact %s: %w", artifact, err)
	}
	return nil
}

// hexToDecimal converts a hex string (with or without 0x prefix) to a decimal string
func hexToDecimal(hex string) (string, error) {
	// Remove 0x prefix if present
	hex = strings.TrimPrefix(hex, "0x")

	// Parse hex string to big.Int
	n := new(big.Int)
	if _, ok := n.SetString(hex, 16); !ok {
		return "", fmt.Errorf("invalid hex string: %s", hex)
	}

	// Convert to decimal string
	return n.String(), nil
}

// parseStateFile parses the state.json file and extracts addresses
func parseStateFile(r io.Reader) (DeploymentStateAddresses, error) {
	var state StateFile
	if err := json.NewDecoder(r).Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode state file: %w", err)
	}

	result := make(DeploymentStateAddresses)

	for _, deployment := range state.OpChainDeployments {
		// Get the chain ID
		idValue, ok := deployment["id"]
		if !ok {
			continue
		}
		hexID, ok := idValue.(string)
		if !ok {
			continue
		}

		// Convert hex ID to decimal
		id, err := hexToDecimal(hexID)
		if err != nil {
			continue
		}

		addresses := make(DeploymentAddresses)

		// Look for address fields in the deployment map
		for key, value := range deployment {
			if strings.HasSuffix(key, "Address") {
				key = strings.TrimSuffix(key, "Address")
				addresses[key] = value.(string)
			}
		}

		if len(addresses) > 0 {
			result[id] = addresses
		}
	}

	return result, nil
}

// ParseDeployerData downloads and parses the op-deployer state
func ParseDeployerData(enclave string) (DeploymentStateAddresses, error) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "op-deployer-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download the artifact
	if err := downloadArtifact(enclave, "op-deployer-configs", tmpDir); err != nil {
		return nil, err
	}

	// Open and parse the state file
	stateFile := filepath.Join(tmpDir, "state.json")
	f, err := os.Open(stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open state file: %w", err)
	}
	defer f.Close()

	return parseStateFile(f)
}
