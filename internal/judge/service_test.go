package judge

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"oj-lite/internal/platform/logger"
)

func TestServiceRunMatchesStudentAndReferenceReturns(t *testing.T) {
	service := newService(logger.NewLogger("judge-test"))

	report, err := service.run(context.Background(), Request{
		ReferenceCode: "function solution(a, b)\n  return a + b\nend",
		SourceCode:    "function solution(a, b)\n  print(a, b)\n  return a + b\nend",
		TestCases:     `[{"input":[1,2]},{"input":[3,4]}]`,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(report.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(report.Cases))
	}

	first := report.Cases[0]
	if !first.Comparison.Matched {
		t.Fatalf("expected first case to match, reason=%s", first.Comparison.Reason)
	}
	if first.Student.StdoutBuffer != "1\t2" {
		t.Fatalf("unexpected stdout: %q", first.Student.StdoutBuffer)
	}
	if len(first.Student.ReturnValues) != 1 || first.Student.ReturnValues[0] != 3.0 {
		t.Fatalf("unexpected return values: %#v", first.Student.ReturnValues)
	}
}

func TestServiceRunTreatsStdoutLimitAsRuntimeError(t *testing.T) {
	service := newService(logger.NewLogger("judge-test"))
	longLine := strings.Repeat("x", 8193)

	report, err := service.run(context.Background(), Request{
		ReferenceCode: "function solution()\n  return true\nend",
		SourceCode:    fmt.Sprintf("function solution()\n  print(%q)\n  return true\nend", longLine),
		TestCases:     `[[]]`,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(report.Cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(report.Cases))
	}

	result := report.Cases[0].Student
	if result.ErrorMessage != StdoutLimitExceededMessage {
		t.Fatalf("expected stdout limit error, got %q", result.ErrorMessage)
	}
	if result.StdoutBuffer != StdoutLimitExceededMessage {
		t.Fatalf("expected stdout limit buffer, got %q", result.StdoutBuffer)
	}
	if report.Cases[0].Comparison.Matched {
		t.Fatal("expected comparison mismatch after stdout limit")
	}
}

func TestServiceRunCapturesRuntimeErrorIntoMixedStdout(t *testing.T) {
	service := newService(logger.NewLogger("judge-test"))

	report, err := service.run(context.Background(), Request{
		ReferenceCode: "function solution(a)\n  return a + 1\nend",
		SourceCode: "function solution(a)\n" +
			"  print('before crash')\n" +
			"  local x = nil\n" +
			"  return x.value\n" +
			"end",
		TestCases: `[[1]]`,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	if len(report.Cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(report.Cases))
	}

	result := report.Cases[0].Student
	if result.ErrorMessage == "" {
		t.Fatal("expected runtime error message")
	}
	if !strings.Contains(result.StdoutBuffer, "before crash") {
		t.Fatalf("expected stdout to include regular output, got %q", result.StdoutBuffer)
	}
	if !strings.Contains(result.StdoutBuffer, "[runtime_error]") {
		t.Fatalf("expected stdout to include runtime error marker, got %q", result.StdoutBuffer)
	}
	if report.Cases[0].Comparison.Matched {
		t.Fatal("expected comparison mismatch after student runtime error")
	}
}

func TestRunnerParseTestCasesSupportsArrayAndObjectForms(t *testing.T) {
	runner := newRunner(logger.NewLogger("judge-test"))

	cases, err := runner.parseTestCases(`[{"input":[1,2]}, [[3,4]], true]`)
	if err != nil {
		t.Fatalf("parse test cases: %v", err)
	}

	if len(cases) != 3 {
		t.Fatalf("expected 3 cases, got %d", len(cases))
	}
	if len(cases[0].Input) != 2 || cases[0].Input[0] != 1.0 || cases[0].Input[1] != 2.0 {
		t.Fatalf("unexpected first case: %#v", cases[0].Input)
	}
	if len(cases[1].Input) != 1 {
		t.Fatalf("unexpected second case: %#v", cases[1].Input)
	}
	nested, ok := cases[1].Input[0].([]any)
	if !ok || len(nested) != 2 || nested[0] != 3.0 || nested[1] != 4.0 {
		t.Fatalf("unexpected nested array argument: %#v", cases[1].Input)
	}
	if len(cases[2].Input) != 1 || cases[2].Input[0] != true {
		t.Fatalf("unexpected third case: %#v", cases[2].Input)
	}
}
