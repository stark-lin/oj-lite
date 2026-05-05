// Defines request and response DTOs for the submission module.

package submission

import "encoding/json"

type createSubmissionRequest struct {
	LessonQuestionID int64  `json:"lesson_question_id"`
	SourceCode       string `json:"source_code"`
}

type submissionDTO struct {
	ID               int64           `json:"id"`
	EnrollmentID     int64           `json:"enrollment_id"`
	LessonID         int64           `json:"lesson_id"`
	LessonQuestionID int64           `json:"lesson_question_id"`
	QuestionID       int64           `json:"question_id"`
	QuestionTitle    string          `json:"question_title"`
	Status           string          `json:"status"`
	Verdict          string          `json:"verdict,omitempty"`
	SourceCode       string          `json:"source_code,omitempty"`
	StdoutBuffer     json.RawMessage `json:"stdout_buffer,omitempty"`
	ErrorMessage     string          `json:"error_message,omitempty"`
	JudgeReport      json.RawMessage `json:"judge_report,omitempty"`
	SubmittedAt      string          `json:"submitted_at"`
	FinishedAt       string          `json:"finished_at,omitempty"`
}

func newSubmissionDTO(item Submission, includeSource bool) submissionDTO {
	dto := submissionDTO{
		ID:               item.ID,
		EnrollmentID:     item.EnrollmentID,
		LessonID:         item.LessonID,
		LessonQuestionID: item.LessonQuestionID,
		QuestionID:       item.QuestionID,
		QuestionTitle:    item.QuestionTitle,
		Status:           item.Status,
		Verdict:          item.Verdict,
		StdoutBuffer:     item.StdoutBuffer,
		ErrorMessage:     item.ErrorMessage,
		JudgeReport:      item.JudgeReport,
		SubmittedAt:      item.SubmittedAt,
		FinishedAt:       item.FinishedAt,
	}
	if includeSource {
		dto.SourceCode = item.SourceCode
	}
	return dto
}
