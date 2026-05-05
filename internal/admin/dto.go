// Defines request and response DTOs for the admin module.

package admin

import "encoding/json"

type createTeacherRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type updateTeacherRequest struct {
	Username *string `json:"username"`
	Status   *string `json:"status"`
}

type resetTeacherPasswordRequest struct {
	Password string `json:"password"`
}

type teacherDTO struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type lessonRequest struct {
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	SortOrder   int                     `json:"sort_order"`
	Questions   []lessonQuestionRequest `json:"questions"`
}

type lessonQuestionRequest struct {
	ID            int64           `json:"id"`
	Title         string          `json:"title"`
	Description   json.RawMessage `json:"description"`
	StarterCode   string          `json:"starter_code"`
	ReferenceCode string          `json:"reference_code"`
	TestCases     json.RawMessage `json:"test_cases"`
	SortOrder     int             `json:"sort_order"`
}

type lessonDTO struct {
	ID          int64               `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	SortOrder   int                 `json:"sort_order"`
	CreatedAt   string              `json:"created_at"`
	Questions   []lessonQuestionDTO `json:"questions"`
}

type lessonQuestionDTO struct {
	LessonQuestionID int64           `json:"lesson_question_id"`
	ID               int64           `json:"id"`
	Title            string          `json:"title"`
	Description      json.RawMessage `json:"description"`
	StarterCode      string          `json:"starter_code"`
	ReferenceCode    string          `json:"reference_code"`
	TestCases        json.RawMessage `json:"test_cases"`
	SortOrder        int             `json:"sort_order"`
	CreatedAt        string          `json:"created_at"`
}

func newTeacherDTO(teacher Teacher) teacherDTO {
	return teacherDTO{
		ID:        teacher.ID,
		Username:  teacher.Username,
		Role:      teacher.Role,
		Status:    teacher.Status,
		CreatedAt: teacher.CreatedAt,
	}
}

func newLessonDTO(lesson Lesson) lessonDTO {
	questions := make([]lessonQuestionDTO, 0, len(lesson.Questions))
	for _, question := range lesson.Questions {
		questions = append(questions, newLessonQuestionDTO(question))
	}

	return lessonDTO{
		ID:          lesson.ID,
		Title:       lesson.Title,
		Description: lesson.Description,
		SortOrder:   lesson.SortOrder,
		CreatedAt:   lesson.CreatedAt,
		Questions:   questions,
	}
}

func newLessonQuestionDTO(question LessonQuestion) lessonQuestionDTO {
	return lessonQuestionDTO{
		LessonQuestionID: question.LessonQuestionID,
		ID:               question.ID,
		Title:            question.Title,
		Description:      marshalJSONObjectString(question.Description, "Description"),
		StarterCode:      question.StarterCode,
		ReferenceCode:    question.ReferenceCode,
		TestCases:        json.RawMessage(question.TestCases),
		SortOrder:        question.SortOrder,
		CreatedAt:        question.CreatedAt,
	}
}
