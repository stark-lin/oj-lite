package seed

import "encoding/json"

type lessonFile struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	SortOrder   int                  `json:"sort_order"`
	Questions   []lessonQuestionFile `json:"questions"`
}

type lessonQuestionFile struct {
	Title         string          `json:"title"`
	Description   json.RawMessage `json:"description"`
	StarterCode   string          `json:"starter_code"`
	ReferenceCode string          `json:"reference_code"`
	TestCases     json.RawMessage `json:"test_cases"`
	SortOrder     int             `json:"sort_order"`
}
