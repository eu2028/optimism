package chainconfig

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
)

var OPSepoliaChainConfig, OPMainnetChainConfig *params.ChainConfig

func init() {
	mustLoadConfig := func(chainID uint64) *params.ChainConfig {
		cfg, err := params.LoadOPStackChainConfig(chainID)
		if err != nil {
			panic(err)
		}
		return cfg
	}
	OPSepoliaChainConfig = mustLoadConfig(11155420)
	OPMainnetChainConfig = mustLoadConfig(10)
}

var L2ChainConfigsByChainID = map[uint64]*params.ChainConfig{
	11155420: OPSepoliaChainConfig,
	10:       OPMainnetChainConfig,
}

//go:embed configs/*json
var customChainConfigFS embed.FS

func RollupConfigByChainID(chainID uint64) (*rollup.Config, error) {
	config, err := rollup.LoadOPStackRollupConfig(chainID)
	if err == nil {
		return config, err
	}
	return rollupConfigByChainID(chainID, customChainConfigFS)
}

func rollupConfigByChainID(chainID uint64, customChainFS embed.FS) (*rollup.Config, error) {
	// Load custom rollup configs from embed FS
	file, err := customChainFS.Open(fmt.Sprintf("configs/%d-rollup.json", chainID))
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no rollup config available for chain ID: %d", chainID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get rollup config for chain ID %d: %w", chainID, err)
	}
	defer file.Close()

	var customRollupConfig rollup.Config
	return &customRollupConfig, customRollupConfig.ParseRollupConfig(file)
}

func ChainConfigByChainID(chainID uint64) (*params.ChainConfig, error) {
	config, err := params.LoadOPStackChainConfig(chainID)
	if err == nil {
		return config, err
	}
	return chainConfigByChainID(chainID, customChainConfigFS)
}

func chainConfigByChainID(chainID uint64, customChainFS embed.FS) (*params.ChainConfig, error) {
	// Load from custom chain configs from embed FS
	data, err := customChainFS.ReadFile(fmt.Sprintf("configs/%d-genesis-l2.json", chainID))
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no chain config available for chain ID: %d", chainID)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get chain config for chain ID %d: %w", chainID, err)
	}
	var genesis core.Genesis
	err = json.Unmarshal(data, &genesis)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain config for chain ID %d: %w", chainID, err)
	}
	return genesis.Config, nil
}
