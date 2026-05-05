// Defines submission entities and status fields.

package submission

import "encoding/json"

type Submission struct {
	ID               int64
	EnrollmentID     int64
	LessonID         int64
	LessonQuestionID int64
	QuestionID       int64
	QuestionTitle    string
	Status           string
	Verdict          string
	SourceCode       string
	StdoutBuffer     json.RawMessage
	ErrorMessage     string
	JudgeReport      json.RawMessage
	SubmittedAt      string
	FinishedAt       string
}

type submissionTarget struct {
	EnrollmentID     int64
	LessonID         int64
	LessonQuestionID int64
	QuestionID       int64
	QuestionTitle    string
}
