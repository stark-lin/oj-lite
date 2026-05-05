// Defines response structures for progress statistics.

package progress

import "encoding/json"

type studentProgressDTO struct {
	ID             int64                 `json:"id"`
	Username       string                `json:"username"`
	Role           string                `json:"role"`
	Status         string                `json:"status"`
	CreatedAt      string                `json:"created_at"`
	CurrentLesson  progressLessonDTO     `json:"current_lesson"`
	LessonProgress lessonProgressDTO     `json:"lesson_progress"`
	Latest         *submissionSummaryDTO `json:"latest_submission,omitempty"`
}

type progressLessonDTO struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

type lessonProgressDTO struct {
	Accepted int `json:"accepted"`
	Total    int `json:"total"`
}

type submissionSummaryDTO struct {
	ID               int64           `json:"id"`
	EnrollmentID     int64           `json:"enrollment_id"`
	StudentID        int64           `json:"student_id"`
	StudentUsername  string          `json:"student_username"`
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

func newStudentProgressDTO(item StudentProgress) studentProgressDTO {
	dto := studentProgressDTO{
		ID:        item.StudentID,
		Username:  item.Username,
		Role:      item.Role,
		Status:    item.Status,
		CreatedAt: item.CreatedAt,
		CurrentLesson: progressLessonDTO{
			ID:    item.CurrentLessonID,
			Title: item.CurrentLessonTitle,
		},
		LessonProgress: lessonProgressDTO{
			Accepted: item.AcceptedCount,
			Total:    item.TotalCount,
		},
	}
	if item.LatestSubmission != nil {
		summary := newSubmissionSummaryDTO(*item.LatestSubmission, false)
		dto.Latest = &summary
	}
	return dto
}

func newSubmissionSummaryDTO(item Submission, includeSource bool) submissionSummaryDTO {
	dto := submissionSummaryDTO{
		ID:               item.ID,
		EnrollmentID:     item.EnrollmentID,
		StudentID:        item.StudentID,
		StudentUsername:  item.StudentUsername,
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
