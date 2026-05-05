// Defines request and response structures for the question module.

package question

import "encoding/json"

type createQuestionRequest struct {
	Title         string          `json:"title"`
	Description   json.RawMessage `json:"description"`
	StarterCode   string          `json:"starter_code"`
	ReferenceCode string          `json:"reference_code"`
	TestCases     json.RawMessage `json:"test_cases"`
}

type questionDTO struct {
	ID            int64           `json:"id"`
	Title         string          `json:"title"`
	Description   json.RawMessage `json:"description"`
	StarterCode   string          `json:"starter_code"`
	ReferenceCode string          `json:"reference_code"`
	TestCases     json.RawMessage `json:"test_cases"`
	CreatedAt     string          `json:"created_at"`
}

type studentQuestionDTO struct {
	ID               int64           `json:"id"`
	LessonQuestionID int64           `json:"lesson_question_id"`
	Title            string          `json:"title"`
	Description      json.RawMessage `json:"description"`
	StarterCode      string          `json:"starter_code"`
	SortOrder        int             `json:"sort_order"`
	CreatedAt        string          `json:"created_at"`
}

func newQuestionDTO(question Question) questionDTO {
	return questionDTO{
		ID:            question.ID,
		Title:         question.Title,
		Description:   marshalJSONObjectString(question.Description, "Description"),
		StarterCode:   question.StarterCode,
		ReferenceCode: question.ReferenceCode,
		TestCases:     json.RawMessage(question.TestCases),
		CreatedAt:     question.CreatedAt,
	}
}

func newStudentQuestionDTO(question StudentQuestion) studentQuestionDTO {
	return studentQuestionDTO{
		ID:               question.ID,
		LessonQuestionID: question.LessonQuestionID,
		Title:            question.Title,
		Description:      marshalJSONObjectString(question.Description, "Description"),
		StarterCode:      question.StarterCode,
		SortOrder:        question.SortOrder,
		CreatedAt:        question.CreatedAt,
	}
}
