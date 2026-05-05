// Defines request and response DTOs for the lesson module.

package lesson

type createLessonRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type lessonDTO struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
}

type lessonQuestionDTO struct {
	ID         int64  `json:"id"`
	LessonID   int64  `json:"lesson_id"`
	QuestionID int64  `json:"question_id"`
	Title      string `json:"title"`
	SortOrder  int    `json:"sort_order"`
	CreatedAt  string `json:"created_at"`
}

func newLessonDTO(lesson Lesson) lessonDTO {
	return lessonDTO{
		ID:          lesson.ID,
		Title:       lesson.Title,
		Description: lesson.Description,
		SortOrder:   lesson.SortOrder,
		CreatedAt:   lesson.CreatedAt,
	}
}

func newLessonQuestionDTO(question LessonQuestion) lessonQuestionDTO {
	return lessonQuestionDTO{
		ID:         question.ID,
		LessonID:   question.LessonID,
		QuestionID: question.QuestionID,
		Title:      question.Title,
		SortOrder:  question.SortOrder,
		CreatedAt:  question.CreatedAt,
	}
}
