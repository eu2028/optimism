package kurtosis

import (
	"os"
	"strings"
	"testing"
)

func TestParseStateFile(t *testing.T) {
	stateJSON := `{
		"opChainDeployments": [
			{
				"id": "0x000000000000000000000000000000000000000000000000000000000020d5e4",
				"L1CrossDomainMessengerAddress": "0x123",
				"L1StandardBridgeAddress":       "0x456",
				"L2OutputOracleAddress":         "0x789"
			},
			{
				"id": "0x000000000000000000000000000000000000000000000000000000000020d5e5",
				"L1CrossDomainMessengerAddress": "0xabc",
				"L1StandardBridgeAddress":       "0xdef",
				"someOtherField": 123,
				"L2OutputOracleAddress":         "0xghi"
			}
		]
	}`

	result, err := parseStateFile(strings.NewReader(stateJSON))
	if err != nil {
		t.Fatalf("Failed to parse state file: %v", err)
	}

	tests := []struct {
		chainID  string
		expected DeploymentAddresses
	}{
		{
			chainID: "2151908",
			expected: DeploymentAddresses{
				"L1CrossDomainMessenger": "0x123",
				"L1StandardBridge":       "0x456",
				"L2OutputOracle":         "0x789",
			},
		},
		{
			chainID: "2151909",
			expected: DeploymentAddresses{
				"L1CrossDomainMessenger": "0xabc",
				"L1StandardBridge":       "0xdef",
				"L2OutputOracle":         "0xghi",
			},
		},
	}

	for _, tt := range tests {
		chain, ok := result[tt.chainID]
		if !ok {
			t.Fatalf("Chain %s not found in result", tt.chainID)
		}

		for key, expected := range tt.expected {
			if actual := chain[key]; actual != expected {
				t.Errorf("Chain %s, %s: expected %s, got %s", tt.chainID, key, expected, actual)
			}
		}
	}
}

func TestParseStateFileErrors(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "empty json",
			json:    "",
			wantErr: true,
		},
		{
			name:    "invalid json",
			json:    "{invalid",
			wantErr: true,
		},
		{
			name: "missing deployments",
			json: `{
				"otherField": []
			}`,
			wantErr: false,
		},
		{
			name: "invalid address type",
			json: `{
				"opChainDeployments": [
					{
						"id": "3151909",
						"data": {
							"L1CrossDomainMessengerAddress": 123
						}
					}
				]
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseStateFile(strings.NewReader(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStateFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDownloadArtifact(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-artifact-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with invalid enclave
	err = downloadArtifact("invalid-enclave", "invalid-artifact", tmpDir)
	if err == nil {
		t.Error("Expected error for invalid enclave, got nil")
	}
}
