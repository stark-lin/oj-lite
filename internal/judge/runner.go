// Runs Lua code and captures return values, stdout, and runtime errors.

package judge

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"oj-lite/internal/platform/logger"

	lua "github.com/yuin/gopher-lua"
)

const solutionFunctionName = "solution"

type runnerCase struct {
	Input []any
}

type runner struct {
	log     *logger.Logger
	sandbox *sandbox
}

func newRunner(log *logger.Logger) *runner {
	return &runner{
		log:     log,
		sandbox: newSandbox(log),
	}
}

func (runner *runner) newState() *lua.LState {
	return runner.sandbox.newState()
}

func (runner *runner) parseTestCases(raw string) ([]runnerCase, error) {
	var decoded []any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return nil, fmt.Errorf("parse test_cases: %w", err)
	}

	cases := make([]runnerCase, 0, len(decoded))
	for index, item := range decoded {
		parsed, err := normalizeRunnerCase(item)
		if err != nil {
			return nil, fmt.Errorf("parse test_cases[%d]: %w", index, err)
		}
		cases = append(cases, parsed)
	}

	return cases, nil
}

func (runner *runner) execute(ctx context.Context, code string, testCase runnerCase) ExecuteReport {
	state := runner.newState()
	defer state.Close()

	output := newOutputBuffer(runner.sandbox.config.stdoutMaxBytes)
	registerPrint(state, output)

	runCtx := ctx
	cancel := func() {}
	if runCtx == nil {
		runCtx = context.Background()
	}
	if runner.sandbox.config.executionTimeout > 0 {
		runCtx, cancel = context.WithTimeout(runCtx, runner.sandbox.config.executionTimeout)
	}
	defer cancel()
	state.SetContext(runCtx)

	if err := state.DoString(code); err != nil {
		message := err.Error()
		output.appendError("runtime_error", message)
		return ExecuteReport{
			ErrorMessage:        message,
			StdoutBuffer:        output.String(),
			StdoutLimitExceeded: output.limitExceeded,
		}
	}

	fn := state.GetGlobal(solutionFunctionName)
	if fn.Type() != lua.LTFunction {
		message := "missing global function solution"
		output.appendError("runtime_error", message)
		return ExecuteReport{
			ErrorMessage:        message,
			StdoutBuffer:        output.String(),
			StdoutLimitExceeded: output.limitExceeded,
		}
	}

	args := make([]lua.LValue, 0, len(testCase.Input))
	for index, item := range testCase.Input {
		value, err := toLuaValue(state, item)
		if err != nil {
			message := fmt.Sprintf("convert input[%d]: %v", index, err)
			output.appendError("system_error", message)
			return ExecuteReport{
				ErrorMessage:        message,
				StdoutBuffer:        output.String(),
				StdoutLimitExceeded: output.limitExceeded,
			}
		}
		args = append(args, value)
	}

	top := state.GetTop()
	if err := state.CallByParam(lua.P{
		Fn:      fn,
		NRet:    lua.MultRet,
		Protect: true,
	}, args...); err != nil {
		message := err.Error()
		output.appendError("runtime_error", message)
		return ExecuteReport{
			ErrorMessage:        message,
			StdoutBuffer:        output.String(),
			StdoutLimitExceeded: output.limitExceeded,
		}
	}

	returnCount := state.GetTop() - top
	returnValues := make([]any, 0, returnCount)
	for i := 0; i < returnCount; i++ {
		value, err := normalizeLuaValue(state.Get(top + 1 + i))
		if err != nil {
			message := fmt.Sprintf("normalize return[%d]: %v", i, err)
			output.appendError("system_error", message)
			state.Pop(returnCount)
			return ExecuteReport{
				ErrorMessage:        message,
				StdoutBuffer:        output.String(),
				StdoutLimitExceeded: output.limitExceeded,
			}
		}
		returnValues = append(returnValues, value)
	}
	state.Pop(returnCount)

	return ExecuteReport{
		ReturnValues:        returnValues,
		StdoutBuffer:        output.String(),
		StdoutLimitExceeded: output.limitExceeded,
	}
}

type outputBuffer struct {
	bytes         int
	lines         []string
	maxBytes      int
	limitExceeded bool
}

func newOutputBuffer(maxBytes int) *outputBuffer {
	return &outputBuffer{maxBytes: maxBytes}
}

func (buffer *outputBuffer) append(line string) {
	if buffer.limitExceeded {
		return
	}

	added := len(line)
	if len(buffer.lines) > 0 {
		added++
	}
	if buffer.maxBytes > 0 && buffer.bytes+added > buffer.maxBytes {
		buffer.limitExceeded = true
		buffer.lines = []string{StdoutLimitExceededMessage}
		buffer.bytes = len(StdoutLimitExceededMessage)
		return
	}

	buffer.lines = append(buffer.lines, line)
	buffer.bytes += added
}

func (buffer *outputBuffer) appendError(kind, message string) {
	buffer.append(fmt.Sprintf("[%s] %s", kind, message))
}

func (buffer *outputBuffer) String() string {
	return strings.Join(buffer.lines, "\n")
}

func registerPrint(state *lua.LState, output *outputBuffer) {
	state.SetGlobal("print", state.NewFunction(func(L *lua.LState) int {
		parts := make([]string, 0, L.GetTop())
		for index := 1; index <= L.GetTop(); index++ {
			parts = append(parts, L.Get(index).String())
		}
		output.append(strings.Join(parts, "\t"))
		if output.limitExceeded {
			L.RaiseError(StdoutLimitExceededMessage)
		}
		return 0
	}))
}

func normalizeRunnerCase(raw any) (runnerCase, error) {
	switch value := raw.(type) {
	case []any:
		return runnerCase{Input: value}, nil
	case map[string]any:
		if input, ok := value["input"]; ok {
			return inputToRunnerCase(input)
		}
		if input, ok := value["args"]; ok {
			return inputToRunnerCase(input)
		}
		return runnerCase{Input: []any{value}}, nil
	default:
		return runnerCase{Input: []any{value}}, nil
	}
}

func inputToRunnerCase(raw any) (runnerCase, error) {
	switch value := raw.(type) {
	case []any:
		return runnerCase{Input: value}, nil
	default:
		return runnerCase{Input: []any{value}}, nil
	}
}

func toLuaValue(state *lua.LState, value any) (lua.LValue, error) {
	switch typed := value.(type) {
	case nil:
		return lua.LNil, nil
	case bool:
		if typed {
			return lua.LTrue, nil
		}
		return lua.LFalse, nil
	case float64:
		return lua.LNumber(typed), nil
	case string:
		return lua.LString(typed), nil
	case []any:
		table := state.NewTable()
		for index, item := range typed {
			converted, err := toLuaValue(state, item)
			if err != nil {
				return nil, err
			}
			table.RawSetInt(index+1, converted)
		}
		return table, nil
	case map[string]any:
		table := state.NewTable()
		for key, item := range typed {
			converted, err := toLuaValue(state, item)
			if err != nil {
				return nil, err
			}
			table.RawSetString(key, converted)
		}
		return table, nil
	default:
		return nil, fmt.Errorf("unsupported value type %T", value)
	}
}

func normalizeLuaValue(value lua.LValue) (any, error) {
	switch typed := value.(type) {
	case *lua.LNilType:
		return nil, nil
	case lua.LBool:
		return bool(typed), nil
	case lua.LNumber:
		return float64(typed), nil
	case lua.LString:
		return string(typed), nil
	case *lua.LTable:
		return normalizeLuaTable(typed)
	default:
		return nil, fmt.Errorf("unsupported lua value type %s", value.Type().String())
	}
}

func normalizeLuaTable(table *lua.LTable) (any, error) {
	arrayValues := map[int]any{}
	objectValues := map[string]any{}
	count := 0
	maxArrayIndex := 0
	arrayOnly := true

	var firstErr error
	table.ForEach(func(key lua.LValue, value lua.LValue) {
		if firstErr != nil {
			return
		}
		count++

		normalized, err := normalizeLuaValue(value)
		if err != nil {
			firstErr = err
			return
		}

		if index, ok := luaArrayIndex(key); ok {
			arrayValues[index] = normalized
			if index > maxArrayIndex {
				maxArrayIndex = index
			}
			return
		}

		arrayOnly = false
		objectKey, err := normalizeLuaObjectKey(key)
		if err != nil {
			firstErr = err
			return
		}
		objectValues[objectKey] = normalized
	})
	if firstErr != nil {
		return nil, firstErr
	}

	if arrayOnly && len(arrayValues) == count && maxArrayIndex == count {
		values := make([]any, count)
		for index := 1; index <= count; index++ {
			values[index-1] = arrayValues[index]
		}
		return values, nil
	}

	for index, value := range arrayValues {
		objectValues[fmt.Sprintf("n:%d", index)] = value
	}
	return objectValues, nil
}

func luaArrayIndex(value lua.LValue) (int, bool) {
	number, ok := value.(lua.LNumber)
	if !ok {
		return 0, false
	}
	raw := float64(number)
	if raw < 1 || math.Trunc(raw) != raw {
		return 0, false
	}
	return int(raw), true
}

func normalizeLuaObjectKey(value lua.LValue) (string, error) {
	switch typed := value.(type) {
	case lua.LString:
		return "s:" + string(typed), nil
	case lua.LBool:
		return fmt.Sprintf("b:%t", bool(typed)), nil
	case lua.LNumber:
		return fmt.Sprintf("n:%v", float64(typed)), nil
	default:
		return "", fmt.Errorf("unsupported lua table key type %s", value.Type().String())
	}
}
