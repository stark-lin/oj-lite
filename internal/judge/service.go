// Provides the main judging entry point; evaluates a question definition and source code and returns a result.

package judge

import (
	"context"

	"oj-lite/internal/platform/logger"
)

type service struct {
	comparator *comparator
	log        *logger.Logger
	runner     *runner
}

func newService(log *logger.Logger) *service {
	return &service{
		comparator: newComparator(log),
		log:        log,
		runner:     newRunner(log),
	}
}

type Engine struct {
	service *service
}

func New(log *logger.Logger) *Engine {
	return &Engine{service: newService(log)}
}

func (engine *Engine) Run(ctx context.Context, request Request) (Report, error) {
	return engine.service.run(ctx, request)
}

func (service *service) run(ctx context.Context, request Request) (Report, error) {
	testCases, err := service.runner.parseTestCases(request.TestCases)
	if err != nil {
		return Report{}, err
	}

	report := Report{
		Cases: make([]CaseReport, 0, len(testCases)),
	}
	for index, testCase := range testCases {
		student := service.runner.execute(ctx, request.SourceCode, testCase)
		reference := service.runner.execute(ctx, request.ReferenceCode, testCase)
		if student.StdoutLimitExceeded {
			student.ErrorMessage = StdoutLimitExceededMessage
			student.ReturnValues = nil
		}
		report.Cases = append(report.Cases, CaseReport{
			Comparison: service.comparator.compare(student, reference),
			Index:      index + 1,
			Input:      testCase.Input,
			Reference:  reference,
			Student:    student,
		})
	}

	return report, nil
}
