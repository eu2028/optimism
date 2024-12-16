package kurtosis

import (
	"strings"
	"testing"
)

func TestParseSpec(t *testing.T) {
	yamlContent := `
optimism_package:
  chains:
    - participants:
        - el_type: op-geth
      network_params:
        name: op-rollup-one
        network_id: "3151909"
      additional_services:
        - blockscout
    - participants:
        - el_type: op-geth
      network_params:
        name: op-rollup-two
        network_id: "3151910"
      additional_services:
        - blockscout
ethereum_package:
  participants:
    - el_type: geth
    - el_type: reth
  network_params:
    preset: minimal
    genesis_delay: 5
`

	result, err := ParseSpec(strings.NewReader(yamlContent))
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	expectedChains := []ChainSpec{
		{
			Name:      "op-rollup-one",
			NetworkID: "3151909",
		},
		{
			Name:      "op-rollup-two",
			NetworkID: "3151910",
		},
	}

	if len(result.Chains) != len(expectedChains) {
		t.Fatalf("Expected %d chains, got %d", len(expectedChains), len(result.Chains))
	}

	for i, expected := range expectedChains {
		actual := result.Chains[i]
		if actual.Name != expected.Name {
			t.Errorf("Chain %d: expected name %q, got %q", i, expected.Name, actual.Name)
		}
		if actual.NetworkID != expected.NetworkID {
			t.Errorf("Chain %d: expected network ID %q, got %q", i, expected.NetworkID, actual.NetworkID)
		}
	}
}

func TestParseSpecErrors(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "empty yaml",
			yaml: "",
		},
		{
			name:    "invalid yaml",
			yaml:    "invalid: [yaml: content",
			wantErr: true,
		},
		{
			name: "missing network params",
			yaml: `
optimism_package:
  chains:
    - participants:
        - el_type: op-geth
      additional_services:
        - blockscout`,
		},
		{
			name: "missing chains",
			yaml: `
optimism_package:
  other_field: value`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSpec(strings.NewReader(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSpec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
