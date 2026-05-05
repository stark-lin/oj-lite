package judge

import (
	"testing"

	"oj-lite/internal/platform/logger"

	lua "github.com/yuin/gopher-lua"
)

func TestSandboxNewStateStartsWithoutBuiltinLibraries(t *testing.T) {
	sandbox := newSandbox(logger.NewLogger("judge-test"))
	state := sandbox.newState()
	defer state.Close()

	for _, globalName := range []string{
		"print",
		"require",
		"os",
		"io",
		"debug",
		"package",
		"math",
		"string",
		"table",
		"coroutine",
	} {
		if got := state.GetGlobal(globalName); got != lua.LNil {
			t.Fatalf("expected %q to be unavailable, got %s", globalName, got.Type().String())
		}
	}
}

func TestRunnerNewStateUsesSandboxWithoutBuiltinLibraries(t *testing.T) {
	runner := newRunner(logger.NewLogger("judge-test"))
	state := runner.newState()
	defer state.Close()

	if got := state.GetGlobal("print"); got != lua.LNil {
		t.Fatalf("expected runner state to inherit sandbox restrictions, got %s", got.Type().String())
	}
}
