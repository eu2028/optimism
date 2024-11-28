package e2e_backends

import (
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-chain-ops/interopgen"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
)

func newFullSuperSystemSpec(spec *interfaces.SuperSystemSpec) *interfaces.FullSuperSystemSpec {
	l2ChainIDs := []uint64{}
	for i := 0; i < spec.Config.NumberOfL2s(); i++ {
		l2ChainIDs = append(l2ChainIDs, uint64(900200+i))
	}

	recipe := interopgen.InteropDevRecipe{
		L1ChainID:        900100,
		L2ChainIDs:       l2ChainIDs,
		GenesisTimestamp: uint64(time.Now().Unix() + 3), // start chain 3 seconds from now
	}
	worldResources := interfaces.WorldResourcePaths{
		FoundryArtifacts: "../../packages/contracts-bedrock/forge-artifacts",
		SourceMap:        "../../packages/contracts-bedrock",
	}

	return &interfaces.FullSuperSystemSpec{
		Recipe:          &recipe,
		World:           worldResources,
		SuperSystemSpec: spec,
	}
}

func NewSpecifiedSuperSystem(t testing.TB, spec *interfaces.SuperSystemSpec) (interfaces.SuperSystem, error) {
	fspec := newFullSuperSystemSpec(spec)
	return NewSuperSystem(t, fspec.Recipe, fspec.World, fspec.Config), nil
}
