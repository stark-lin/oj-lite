// Defines aggregate entities used by admin management APIs.

package admin

type Teacher struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string
	Status       string
	CreatedAt    string
}

type Lesson struct {
	ID          int64
	Title       string
	Description string
	SortOrder   int
	CreatedAt   string
	Questions   []LessonQuestion
}

type LessonQuestion struct {
	LessonQuestionID int64
	ID               int64
	Title            string
	Description      string
	StarterCode      string
	ReferenceCode    string
	TestCases        string
	SortOrder        int
	CreatedAt        string
}

type lessonWrite struct {
	Title       string
	Description string
	SortOrder   int
	Questions   []lessonQuestionWrite
}

type lessonQuestionWrite struct {
	ID            int64
	Title         string
	Description   string
	StarterCode   string
	ReferenceCode string
	TestCases     string
	SortOrder     int
}
