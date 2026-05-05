// Defines question entities and judging-related fields.

package question

type Question struct {
	ID            int64
	Title         string
	Description   string
	StarterCode   string
	ReferenceCode string
	TestCases     string
	CreatedAt     string
}

type StudentQuestion struct {
	ID               int64
	LessonQuestionID int64
	Title            string
	Description      string
	StarterCode      string
	SortOrder        int
	CreatedAt        string
}
