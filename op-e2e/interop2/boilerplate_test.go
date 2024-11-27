package interop2

import (
	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/runners"
)

type SuperSystem = interfaces.SuperSystem
type Test = interfaces.Test
type TestSpec = interfaces.TestSpec[SuperSystem]
type SystemTest = runners.SystemTest[SuperSystem]
type TestLogic = interfaces.TestLogic[SuperSystem]
type TestLogicFunc = interfaces.TestLogicFunc[SuperSystem]
