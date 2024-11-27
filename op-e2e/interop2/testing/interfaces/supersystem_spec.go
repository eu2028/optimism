package interfaces

import (
	"github.com/ethereum-optimism/optimism/op-chain-ops/interopgen"
)

type WorldResourcePaths struct {
	FoundryArtifacts string
	SourceMap        string
}

type SuperSystemConfig struct {
	numberOfL2s      int
	mempoolFiltering bool
}

type SuperSystemConfigOption func(*SuperSystemConfig)

func NewSuperSystemConfig(opts ...SuperSystemConfigOption) SuperSystemConfig {
	// "reasonable" defaults ?
	cfg := SuperSystemConfig{
		numberOfL2s:      2,
		mempoolFiltering: true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

func (cfg *SuperSystemConfig) NumberOfL2s() int {
	return cfg.numberOfL2s
}

func (cfg *SuperSystemConfig) MempoolFiltering() bool {
	return cfg.mempoolFiltering
}

func WithNumberOfL2s(n int) SuperSystemConfigOption {
	return func(cfg *SuperSystemConfig) {
		cfg.numberOfL2s = n
	}
}

func WithMempoolFiltering(b bool) SuperSystemConfigOption {
	return func(cfg *SuperSystemConfig) {
		cfg.mempoolFiltering = b
	}
}

type SuperSystemSpec struct {
	Config SuperSystemConfig
}

type FullSuperSystemSpec struct {
	Recipe *interopgen.InteropDevRecipe
	World  WorldResourcePaths
	*SuperSystemSpec
}

func (s SuperSystemSpec) Conform(ss SuperSystem) bool {
	// TODO: do something !
	return true
}
