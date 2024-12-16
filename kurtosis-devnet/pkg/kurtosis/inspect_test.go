package kurtosis

import (
	"testing"
)

func TestParseInspectOutput(t *testing.T) {
	output := `Name:            interop-devnet
UUID:            1aca207b7bfd
Status:          RUNNING
Creation Time:   Mon, 16 Dec 2024 21:43:28 CET
Flags:

========================================= Files Artifacts =========================================
UUID           Name
24fa22fbbe9e   1-lighthouse-geth-0-63
018a906c5ea5   el_cl_genesis_data
7a52f4b6848f   final-genesis-timestamp
1dfce39e2be9   genesis-el-cl-env-file
49805cf85754   genesis_validators_root
02ea3e61386e   jwt_file
19d0b8addd06   keymanager_file
233da3830dd2   op-deployer-configs
15b859be0607   op-deployer-fund-script
27127fc07627   op_jwt_fileop-kurtosis-1
b6740ec44fb2   op_jwt_fileop-kurtosis-2
5ce33ff4e9ef   prysm-password
550585a62aa7   validator-ranges

========================================== User Services ==========================================
UUID           Name                                             Ports                                         Status
295ece6f10b0   cl-1-lighthouse-geth                             http: 4000/tcp -> http://127.0.0.1:56397      RUNNING
                                                                metrics: 5054/tcp -> http://127.0.0.1:56398
                                                                tcp-discovery: 9000/tcp -> 127.0.0.1:56399
                                                                udp-discovery: 9000/udp -> 127.0.0.1:50029
d8010602c8d9   el-1-geth-lighthouse                             engine-rpc: 8551/tcp -> 127.0.0.1:56384       RUNNING
                                                                metrics: 9001/tcp -> http://127.0.0.1:56385
                                                                rpc: 8545/tcp -> 127.0.0.1:56382
                                                                tcp-discovery: 30303/tcp -> 127.0.0.1:56381
                                                                udp-discovery: 30303/udp -> 127.0.0.1:50818
                                                                ws: 8546/tcp -> 127.0.0.1:56383`

	result, err := ParseInspectOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse inspect output: %v", err)
	}

	// Verify file artifacts
	expectedFiles := []string{
		"1-lighthouse-geth-0-63",
		"el_cl_genesis_data",
		"final-genesis-timestamp",
		"genesis-el-cl-env-file",
		"genesis_validators_root",
		"jwt_file",
		"keymanager_file",
		"op-deployer-configs",
		"op-deployer-fund-script",
		"op_jwt_fileop-kurtosis-1",
		"op_jwt_fileop-kurtosis-2",
		"prysm-password",
		"validator-ranges",
	}

	if len(result.FileArtifacts) != len(expectedFiles) {
		t.Errorf("Expected %d file artifacts, got %d", len(expectedFiles), len(result.FileArtifacts))
	}

	for i, file := range expectedFiles {
		if i >= len(result.FileArtifacts) {
			t.Errorf("Missing expected file artifact: %s", file)
			continue
		}
		if result.FileArtifacts[i] != file {
			t.Errorf("Expected file artifact %s, got %s", file, result.FileArtifacts[i])
		}
	}

	// Verify services and ports
	expectedServices := map[string]map[string]int{
		"cl-1-lighthouse-geth": {
			"http":          56397,
			"metrics":       56398,
			"tcp-discovery": 56399,
			"udp-discovery": 50029,
		},
		"el-1-geth-lighthouse": {
			"engine-rpc":    56384,
			"metrics":       56385,
			"rpc":           56382,
			"tcp-discovery": 56381,
			"udp-discovery": 50818,
			"ws":            56383,
		},
	}

	for service, expectedPorts := range expectedServices {
		ports, exists := result.UserServices[service]
		if !exists {
			t.Errorf("Expected service %s not found", service)
			continue
		}

		for portName, expectedPort := range expectedPorts {
			actualPort, exists := ports[portName]
			if !exists {
				t.Errorf("Expected port %s not found for service %s", portName, service)
				continue
			}
			if actualPort != expectedPort {
				t.Errorf("For service %s port %s: expected port %d, got %d",
					service, portName, expectedPort, actualPort)
			}
		}
	}
}
