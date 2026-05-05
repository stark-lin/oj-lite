// Defines judge sandbox limits, including timeouts, output size, allowed standard libraries, and disabled file I/O, external commands, and dynamic loading.

package judge

import (
	"time"

	"oj-lite/internal/platform/logger"

	lua "github.com/yuin/gopher-lua"
)

type sandboxConfig struct {
	callStackSize    int
	executionTimeout time.Duration
	registrySize     int
	registryMaxSize  int
	registryGrowStep int
	stdoutMaxBytes   int
}

type sandbox struct {
	log    *logger.Logger
	config sandboxConfig
}

func newSandbox(log *logger.Logger) *sandbox {
	return &sandbox{
		log: log,
		config: sandboxConfig{
			callStackSize:    128,
			executionTimeout: 2 * time.Second,
			registrySize:     256,
			registryMaxSize:  0,
			registryGrowStep: 32,
			stdoutMaxBytes:   8192,
		},
	}
}

func (sandbox *sandbox) newState() *lua.LState {
	return lua.NewState(lua.Options{
		CallStackSize:    sandbox.config.callStackSize,
		RegistrySize:     sandbox.config.registrySize,
		RegistryMaxSize:  sandbox.config.registryMaxSize,
		RegistryGrowStep: sandbox.config.registryGrowStep,
		SkipOpenLibs:     true,
	})
}
