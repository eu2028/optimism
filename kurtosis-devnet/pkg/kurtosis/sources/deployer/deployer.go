package deployer

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

// Wallet represents a wallet with optional private key and name
type Wallet struct {
	Address    string
	PrivateKey string
	Name       string
}

// WalletList holds a list of wallets
type WalletList []*Wallet

type DeployerData struct {
	Wallets WalletList
	State   DeploymentStateAddresses
}

// parseWalletsFile parses a JSON file containing wallet information
func parseWalletsFile(r io.Reader) (WalletList, error) {
	// Read all data from reader
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	// Unmarshal into a map first
	var rawData map[string]string
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("failed to decode wallet file: %w", err)
	}

	// Create a map to store wallets by name
	walletMap := make(map[string]Wallet)

	// Process each key-value pair
	for key, value := range rawData {
		if strings.HasSuffix(key, "Address") {
			name := strings.TrimSuffix(key, "Address")
			wallet := walletMap[name]
			wallet.Address = value
			wallet.Name = name
			walletMap[name] = wallet
		} else if strings.HasSuffix(key, "PrivateKey") {
			name := strings.TrimSuffix(key, "PrivateKey")
			wallet := walletMap[name]
			wallet.PrivateKey = value
			wallet.Name = name
			walletMap[name] = wallet
		}
	}

	// Convert map to list
	result := make(WalletList, 0, len(walletMap))

	for _, wallet := range walletMap {
		// Only include wallets that have at least an address
		if wallet.Address != "" {
			result = append(result, &wallet)
		}
	}

	return result, nil
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

// TODO: we should be getting this from somewhere in the enclave. It's ok-ish
// though, as this is hard-coded in ethereum-optimism anyway.
var knownWallets = []*Wallet{
	{
		Name:       "m/44'/60'/0'/0/0",
		Address:    "0x8943545177806ED17B9F23F0a21ee5948eCaa776",
		PrivateKey: "bcdf20249abf0ed6d944c0288fad489e33f66b3960d9e6229c1cd214ed3bbe31",
	},
	{
		Name:       "m/44'/60'/0'/0/1",
		Address:    "0xE25583099BA105D9ec0A67f5Ae86D90e50036425",
		PrivateKey: "39725efee3fb28614de3bacaffe4cc4bd8c436257e2c8bb887c4b5c4be45e76d",
	},
	{
		Name:       "m/44'/60'/0'/0/2",
		Address:    "0x614561D2d143621E126e87831AEF287678B442b8",
		PrivateKey: "53321db7c1e331d93a11a41d16f004d7ff63972ec8ec7c25db329728ceeb1710",
	},
	{
		Name:       "m/44'/60'/0'/0/3",
		Address:    "0xf93Ee4Cf8c6c40b329b0c0626F28333c132CF241",
		PrivateKey: "ab63b23eb7941c1251757e24b3d2350d2bc05c3c388d06f8fe6feafefb1e8c70",
	},
	{
		Name:       "m/44'/60'/0'/0/4",
		Address:    "0x802dCbE1B1A97554B4F50DB5119E37E8e7336417",
		PrivateKey: "5d2344259f42259f82d2c140aa66102ba89b57b4883ee441a8b312622bd42491",
	},
	{
		Name:       "m/44'/60'/0'/0/5",
		Address:    "0xAe95d8DA9244C37CaC0a3e16BA966a8e852Bb6D6",
		PrivateKey: "27515f805127bebad2fb9b183508bdacb8c763da16f54e0678b16e8f28ef3fff",
	},
	{
		Name:       "m/44'/60'/0'/0/6",
		Address:    "0x2c57d1CFC6d5f8E4182a56b4cf75421472eBAEa4",
		PrivateKey: "7ff1a4c1d57e5e784d327c4c7651e952350bc271f156afb3d00d20f5ef924856",
	},
	{
		Name:       "m/44'/60'/0'/0/7",
		Address:    "0x741bFE4802cE1C4b5b00F9Df2F5f179A1C89171A",
		PrivateKey: "3a91003acaf4c21b3953d94fa4a6db694fa69e5242b2e37be05dd82761058899",
	},
	{
		Name:       "m/44'/60'/0'/0/8",
		Address:    "0xc3913d4D8bAb4914328651C2EAE817C8b78E1f4c",
		PrivateKey: "bb1d0f125b4fb2bb173c318cdead45468474ca71474e2247776b2b4c0fa2d3f5",
	},
	{
		Name:       "m/44'/60'/0'/0/9",
		Address:    "0x65D08a056c17Ae13370565B04cF77D2AfA1cB9FA",
		PrivateKey: "850643a0224065ecce3882673c21f56bcf6eef86274cc21cadff15930b59fc8c",
	},
	{
		Name:       "m/44'/60'/0'/0/10",
		Address:    "0x3e95dFbBaF6B348396E6674C7871546dCC568e56",
		PrivateKey: "94eb3102993b41ec55c241060f47daa0f6372e2e3ad7e91612ae36c364042e44",
	},
	{
		Name:       "m/44'/60'/0'/0/11",
		Address:    "0x5918b2e647464d4743601a865753e64C8059Dc4F",
		PrivateKey: "daf15504c22a352648a71ef2926334fe040ac1d5005019e09f6c979808024dc7",
	},
	{
		Name:       "m/44'/60'/0'/0/12",
		Address:    "0x589A698b7b7dA0Bec545177D3963A2741105C7C9",
		PrivateKey: "eaba42282ad33c8ef2524f07277c03a776d98ae19f581990ce75becb7cfa1c23",
	},
	{
		Name:       "m/44'/60'/0'/0/13",
		Address:    "0x4d1CB4eB7969f8806E2CaAc0cbbB71f88C8ec413",
		PrivateKey: "3fd98b5187bf6526734efaa644ffbb4e3670d66f5d0268ce0323ec09124bff61",
	},
	{
		Name:       "m/44'/60'/0'/0/14",
		Address:    "0xF5504cE2BcC52614F121aff9b93b2001d92715CA",
		PrivateKey: "5288e2f440c7f0cb61a9be8afdeb4295f786383f96f5e35eb0c94ef103996b64",
	},
	{
		Name:       "m/44'/60'/0'/0/15",
		Address:    "0xF61E98E7D47aB884C244E39E031978E33162ff4b",
		PrivateKey: "f296c7802555da2a5a662be70e078cbd38b44f96f8615ae529da41122ce8db05",
	},
	{
		Name:       "m/44'/60'/0'/0/16",
		Address:    "0xf1424826861ffbbD25405F5145B5E50d0F1bFc90",
		PrivateKey: "bf3beef3bd999ba9f2451e06936f0423cd62b815c9233dd3bc90f7e02a1e8673",
	},
	{
		Name:       "m/44'/60'/0'/0/17",
		Address:    "0xfDCe42116f541fc8f7b0776e2B30832bD5621C85",
		PrivateKey: "6ecadc396415970e91293726c3f5775225440ea0844ae5616135fd10d66b5954",
	},
	{
		Name:       "m/44'/60'/0'/0/18",
		Address:    "0xD9211042f35968820A3407ac3d80C725f8F75c14",
		PrivateKey: "a492823c3e193d6c595f37a18e3c06650cf4c74558cc818b16130b293716106f",
	},
	{
		Name:       "m/44'/60'/0'/0/19",
		Address:    "0xD8F3183DEF51A987222D845be228e0Bbb932C222",
		PrivateKey: "c5114526e042343c6d1899cad05e1c00ba588314de9b96929914ee0df18d46b2",
	},
}

// ParseDeployerData downloads and parses the op-deployer state
func ParseDeployerData(enclave string) (*DeployerData, error) {
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
	fState, err := os.Open(stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open state file: %w", err)
	}
	defer fState.Close()

	state, err := parseStateFile(fState)
	if err != nil {
		return nil, err
	}

	walletsFile := filepath.Join(tmpDir, "wallets.json")
	fWallets, err := os.Open(walletsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open wallets file: %w", err)
	}
	defer fWallets.Close()

	wallets, err := parseWalletsFile(fWallets)
	if err != nil {
		return nil, err
	}

	wallets = append(wallets, knownWallets...)

	return &DeployerData{State: state, Wallets: wallets}, nil
}
