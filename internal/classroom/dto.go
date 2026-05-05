// Defines request and response structures for the classroom module.

package classroom

type createClassroomRequest struct {
	Name string `json:"name"`
}

type createStudentRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type renameStudentRequest struct {
	Username string `json:"username"`
}

type resetStudentPasswordRequest struct {
	Password string `json:"password"`
}

type setCurrentLessonRequest struct {
	LessonID int64 `json:"lesson_id"`
}

type classroomDTO struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type studentDTO struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at,omitempty"`
}

type classroomLessonDTO struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	IsCurrent   bool   `json:"is_current"`
}

type studentCurrentLessonDTO struct {
	ID          int64                             `json:"id"`
	Title       string                            `json:"title"`
	Description string                            `json:"description"`
	SortOrder   int                               `json:"sort_order"`
	CreatedAt   string                            `json:"created_at"`
	Questions   []studentCurrentLessonQuestionDTO `json:"questions"`
}

type studentCurrentLessonQuestionDTO struct {
	LessonQuestionID int64  `json:"lesson_question_id"`
	QuestionID       int64  `json:"question_id"`
	Title            string `json:"title"`
	SortOrder        int    `json:"sort_order"`
}

func newClassroomDTO(classroom Classroom) classroomDTO {
	return classroomDTO{
		ID:        classroom.ID,
		Name:      classroom.Name,
		CreatedAt: classroom.CreatedAt,
	}
}

func newStudentDTO(student Student) studentDTO {
	return studentDTO{
		ID:        student.ID,
		Username:  student.Username,
		Role:      student.Role,
		Status:    student.Status,
		CreatedAt: student.CreatedAt,
	}
}

func newClassroomLessonDTO(lesson ClassroomLesson) classroomLessonDTO {
	return classroomLessonDTO{
		ID:          lesson.ID,
		Title:       lesson.Title,
		Description: lesson.Description,
		SortOrder:   lesson.SortOrder,
		CreatedAt:   lesson.CreatedAt,
		IsCurrent:   lesson.IsCurrent,
	}
}

func newStudentCurrentLessonDTO(lesson StudentCurrentLesson) studentCurrentLessonDTO {
	questions := make([]studentCurrentLessonQuestionDTO, 0, len(lesson.Questions))
	for _, question := range lesson.Questions {
		questions = append(questions, studentCurrentLessonQuestionDTO{
			LessonQuestionID: question.LessonQuestionID,
			QuestionID:       question.QuestionID,
			Title:            question.Title,
			SortOrder:        question.SortOrder,
		})
	}

	return studentCurrentLessonDTO{
		ID:          lesson.ID,
		Title:       lesson.Title,
		Description: lesson.Description,
		SortOrder:   lesson.SortOrder,
		CreatedAt:   lesson.CreatedAt,
		Questions:   questions,
	}
}
