// Defines core structures for classrooms, enrollments, and current lessons.

package classroom

import "database/sql"

type Classroom struct {
	ID        int64
	TeacherID int64
	Name      string
	CreatedAt string
}

type Student struct {
	ID              int64
	Username        string
	Role            string
	Status          string
	CreatedAt       string
	CurrentLessonID sql.NullInt64
}

type ClassroomLesson struct {
	ID          int64
	Title       string
	Description string
	SortOrder   int
	CreatedAt   string
	IsCurrent   bool
}

type StudentCurrentLesson struct {
	ID          int64
	Title       string
	Description string
	SortOrder   int
	CreatedAt   string
	Questions   []StudentCurrentLessonQuestion
}

type StudentCurrentLessonQuestion struct {
	LessonQuestionID int64
	QuestionID       int64
	Title            string
	SortOrder        int
}
