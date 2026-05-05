package progress

import "encoding/json"

type StudentProgress struct {
	StudentID          int64
	Username           string
	Role               string
	Status             string
	CreatedAt          string
	CurrentLessonID    int64
	CurrentLessonTitle string
	AcceptedCount      int
	TotalCount         int
	LatestSubmission   *Submission
}

type Submission struct {
	ID               int64
	EnrollmentID     int64
	StudentID        int64
	StudentUsername  string
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
