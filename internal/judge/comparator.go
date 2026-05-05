// Compares reference output and student output for equivalence.

package judge

import (
	"fmt"
	"reflect"

	"oj-lite/internal/platform/logger"
)

type comparator struct {
	log *logger.Logger
}

func newComparator(log *logger.Logger) *comparator {
	return &comparator{log: log}
}

func (comparator *comparator) compare(student, reference ExecuteReport) ComparisonReport {
	if reference.ErrorMessage != "" {
		return ComparisonReport{
			Matched: false,
			Reason:  fmt.Sprintf("reference execution failed: %s", reference.ErrorMessage),
		}
	}

	if student.ErrorMessage != "" {
		return ComparisonReport{
			Matched: false,
			Reason:  fmt.Sprintf("student execution failed: %s", student.ErrorMessage),
		}
	}

	if !reflect.DeepEqual(student.ReturnValues, reference.ReturnValues) {
		return ComparisonReport{
			Matched: false,
			Reason:  fmt.Sprintf("return values differ: student=%v reference=%v", student.ReturnValues, reference.ReturnValues),
		}
	}

	return ComparisonReport{Matched: true}
}
