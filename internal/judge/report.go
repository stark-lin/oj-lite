// Builds structured judge reports for submission updates and frontend display.

package judge

type Request struct {
	ReferenceCode string
	SourceCode    string
	TestCases     string
}

type Report struct {
	Cases []CaseReport `json:"cases"`
}

type CaseReport struct {
	Comparison ComparisonReport `json:"comparison"`
	Index      int              `json:"index"`
	Input      []any            `json:"input"`
	Reference  ExecuteReport    `json:"reference"`
	Student    ExecuteReport    `json:"student"`
}

type ComparisonReport struct {
	Matched bool   `json:"matched"`
	Reason  string `json:"reason,omitempty"`
}

type ExecuteReport struct {
	ErrorMessage        string `json:"errorMessage,omitempty"`
	ReturnValues        []any  `json:"returnValues,omitempty"`
	StdoutBuffer        string `json:"stdoutBuffer,omitempty"`
	StdoutLimitExceeded bool   `json:"-"`
}
