// Defines lesson and lesson_question relationship structures.

package lesson

type Lesson struct {
	ID          int64
	Title       string
	Description string
	SortOrder   int
	CreatedAt   string
}

type LessonQuestion struct {
	ID         int64
	LessonID   int64
	QuestionID int64
	Title      string
	SortOrder  int
	CreatedAt  string
}
