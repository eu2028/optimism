package kurtosis

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// ChainSpec represents the network parameters for a chain
type ChainSpec struct {
	Name      string
	NetworkID string
}

// EnclaveSpec represents the parsed chain specifications from the YAML
type EnclaveSpec struct {
	Chains []ChainSpec
}

// NetworkParams represents the network parameters section in the YAML
type NetworkParams struct {
	Name      string `yaml:"name"`
	NetworkID string `yaml:"network_id"`
}

// ChainConfig represents a chain configuration in the YAML
type ChainConfig struct {
	NetworkParams NetworkParams `yaml:"network_params"`
}

// OptimismPackage represents the optimism_package section in the YAML
type OptimismPackage struct {
	Chains []ChainConfig `yaml:"chains"`
}

// YAMLSpec represents the root of the YAML document
type YAMLSpec struct {
	OptimismPackage OptimismPackage `yaml:"optimism_package"`
}

// ParseSpec parses a YAML document and returns the chain specifications
func ParseSpec(r io.Reader) (*EnclaveSpec, error) {
	var yamlSpec YAMLSpec
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&yamlSpec); err != nil {
		if err == io.EOF {
			// Return empty spec for empty input
			return &EnclaveSpec{
				Chains: make([]ChainSpec, 0),
			}, nil
		}
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	result := &EnclaveSpec{
		Chains: make([]ChainSpec, 0, len(yamlSpec.OptimismPackage.Chains)),
	}

	// Extract chain specifications
	for _, chain := range yamlSpec.OptimismPackage.Chains {
		result.Chains = append(result.Chains, ChainSpec{
			Name:      chain.NetworkParams.Name,
			NetworkID: chain.NetworkParams.NetworkID,
		})
	}

	return result, nil
}
