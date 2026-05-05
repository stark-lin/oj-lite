package app

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"oj-lite/internal/platform/config"
	platformdb "oj-lite/internal/platform/db"
	"oj-lite/internal/platform/logger"
	platformpassword "oj-lite/internal/platform/password"
	"oj-lite/internal/platform/session"
	"oj-lite/internal/seed"
)

func TestTeacherSimpleCRUDRoutes(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherCookie := loginAs(t, app, "teacher", "teacher")

	meResponse := performRequest(t, app, http.MethodGet, "/api/me", nil, teacherCookie)
	if meResponse.Code != http.StatusOK {
		t.Fatalf("GET /api/me status = %d, want %d body=%s", meResponse.Code, http.StatusOK, meResponse.Body.String())
	}

	var meEnvelope struct {
		Data struct {
			User struct {
				Username string `json:"username"`
				Role     string `json:"role"`
			} `json:"user"`
		} `json:"data"`
	}
	decodeJSON(t, meResponse.Body.Bytes(), &meEnvelope)
	if meEnvelope.Data.User.Username != "teacher" || meEnvelope.Data.User.Role != "teacher" {
		t.Fatalf("GET /api/me user = %#v, want teacher account", meEnvelope.Data.User)
	}

	classroomPayload := []byte(`{"name":"crud_classroom"}`)
	createClassroom := performRequest(t, app, http.MethodPost, "/api/teacher/classrooms", classroomPayload, teacherCookie)
	if createClassroom.Code != http.StatusCreated {
		t.Fatalf("POST /api/teacher/classrooms status = %d, want %d body=%s", createClassroom.Code, http.StatusCreated, createClassroom.Body.String())
	}

	var classroomEnvelope struct {
		Data struct {
			Classroom struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
			} `json:"classroom"`
		} `json:"data"`
	}
	decodeJSON(t, createClassroom.Body.Bytes(), &classroomEnvelope)
	if classroomEnvelope.Data.Classroom.Name != "crud_classroom" {
		t.Fatalf("created classroom name = %q, want %q", classroomEnvelope.Data.Classroom.Name, "crud_classroom")
	}

	listClassrooms := performRequest(t, app, http.MethodGet, "/api/teacher/classrooms", nil, teacherCookie)
	if listClassrooms.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/classrooms status = %d, want %d body=%s", listClassrooms.Code, http.StatusOK, listClassrooms.Body.String())
	}

	var classroomsEnvelope struct {
		Data struct {
			Classrooms []struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
			} `json:"classrooms"`
		} `json:"data"`
	}
	decodeJSON(t, listClassrooms.Body.Bytes(), &classroomsEnvelope)
	if !containsNamedClassroom(classroomsEnvelope.Data.Classrooms, classroomEnvelope.Data.Classroom.ID, "crud_classroom") {
		t.Fatalf("GET /api/teacher/classrooms missing created classroom: %#v", classroomsEnvelope.Data.Classrooms)
	}

	getClassroom := performRequest(
		t,
		app,
		http.MethodGet,
		"/api/teacher/classrooms/"+itoa(classroomEnvelope.Data.Classroom.ID),
		nil,
		teacherCookie,
	)
	if getClassroom.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/classrooms/:id status = %d, want %d body=%s", getClassroom.Code, http.StatusOK, getClassroom.Body.String())
	}

	lessonID := insertLesson(t, app.db, "Lesson A", "Intro lesson", 1)

	listLessons := performRequest(t, app, http.MethodGet, "/api/teacher/lessons", nil, teacherCookie)
	if listLessons.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/lessons status = %d, want %d body=%s", listLessons.Code, http.StatusOK, listLessons.Body.String())
	}

	var lessonsEnvelope struct {
		Data struct {
			Lessons []struct {
				ID    int64  `json:"id"`
				Title string `json:"title"`
			} `json:"lessons"`
		} `json:"data"`
	}
	decodeJSON(t, listLessons.Body.Bytes(), &lessonsEnvelope)
	if !containsNamedLesson(lessonsEnvelope.Data.Lessons, lessonID, "Lesson A") {
		t.Fatalf("GET /api/teacher/lessons missing created lesson: %#v", lessonsEnvelope.Data.Lessons)
	}

	getLesson := performRequest(
		t,
		app,
		http.MethodGet,
		"/api/teacher/lessons/"+itoa(lessonID),
		nil,
		teacherCookie,
	)
	if getLesson.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/lessons/:id status = %d, want %d body=%s", getLesson.Code, http.StatusOK, getLesson.Body.String())
	}

	questionID := insertQuestionWithContent(
		t,
		app.db,
		"Question A",
		`{"statement":"Sum numbers","input":"Two integers a and b.","output":"Return their sum."}`,
		"function solution(a,b)\n  return 0\nend",
		"function solution(a,b)\n  return a+b\nend",
	)

	listQuestions := performRequest(t, app, http.MethodGet, "/api/teacher/questions", nil, teacherCookie)
	if listQuestions.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/questions status = %d, want %d body=%s", listQuestions.Code, http.StatusOK, listQuestions.Body.String())
	}

	var questionsEnvelope struct {
		Data struct {
			Questions []struct {
				ID          int64           `json:"id"`
				Title       string          `json:"title"`
				Description json.RawMessage `json:"description"`
			} `json:"questions"`
		} `json:"data"`
	}
	decodeJSON(t, listQuestions.Body.Bytes(), &questionsEnvelope)
	var listedQuestion *struct {
		ID          int64           `json:"id"`
		Title       string          `json:"title"`
		Description json.RawMessage `json:"description"`
	}
	for index := range questionsEnvelope.Data.Questions {
		item := &questionsEnvelope.Data.Questions[index]
		if item.ID == questionID {
			listedQuestion = item
			break
		}
	}
	if listedQuestion == nil || listedQuestion.Title != "Question A" {
		t.Fatalf("GET /api/teacher/questions missing created question: %#v", questionsEnvelope.Data.Questions)
	}
	if string(listedQuestion.Description) != `{"statement":"Sum numbers","input":"Two integers a and b.","output":"Return their sum."}` {
		t.Fatalf("listed question description = %s, want compact JSON object", listedQuestion.Description)
	}

	getQuestion := performRequest(
		t,
		app,
		http.MethodGet,
		"/api/teacher/questions/"+itoa(questionID),
		nil,
		teacherCookie,
	)
	if getQuestion.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/questions/:id status = %d, want %d body=%s", getQuestion.Code, http.StatusOK, getQuestion.Body.String())
	}

	var getQuestionEnvelope struct {
		Data struct {
			Question struct {
				Description json.RawMessage `json:"description"`
			} `json:"question"`
		} `json:"data"`
	}
	decodeJSON(t, getQuestion.Body.Bytes(), &getQuestionEnvelope)
	if string(getQuestionEnvelope.Data.Question.Description) != `{"statement":"Sum numbers","input":"Two integers a and b.","output":"Return their sum."}` {
		t.Fatalf("get question description = %s, want compact JSON object", getQuestionEnvelope.Data.Question.Description)
	}
}

func TestTeacherSimpleCRUDRoutesKeepClassroomScopedAndContentGlobal(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherCookie := loginAs(t, app, "teacher", "teacher")
	otherTeacherID := insertTeacher(t, app.db, "other_teacher")
	otherClassroomID := insertClassroom(t, app.db, otherTeacherID, "other_classroom")
	otherLessonID := insertLesson(t, app.db, "Shared Lesson", "desc", 1)
	otherQuestionID := insertQuestion(t, app.db, "Shared Question")

	getClassroom := performRequest(
		t,
		app,
		http.MethodGet,
		"/api/teacher/classrooms/"+itoa(otherClassroomID),
		nil,
		teacherCookie,
	)
	if getClassroom.Code != http.StatusNotFound {
		t.Fatalf("GET foreign classroom status = %d, want %d body=%s", getClassroom.Code, http.StatusNotFound, getClassroom.Body.String())
	}

	getLesson := performRequest(
		t,
		app,
		http.MethodGet,
		"/api/teacher/lessons/"+itoa(otherLessonID),
		nil,
		teacherCookie,
	)
	if getLesson.Code != http.StatusOK {
		t.Fatalf("GET shared lesson status = %d, want %d body=%s", getLesson.Code, http.StatusOK, getLesson.Body.String())
	}

	getQuestion := performRequest(
		t,
		app,
		http.MethodGet,
		"/api/teacher/questions/"+itoa(otherQuestionID),
		nil,
		teacherCookie,
	)
	if getQuestion.Code != http.StatusOK {
		t.Fatalf("GET shared question status = %d, want %d body=%s", getQuestion.Code, http.StatusOK, getQuestion.Body.String())
	}
}

func TestTeacherSimpleCRUDRoutesValidateRequests(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherCookie := loginAs(t, app, "teacher", "teacher")

	invalidClassroom := performRequest(t, app, http.MethodPost, "/api/teacher/classrooms", []byte(`{"name":"   "}`), teacherCookie)
	if invalidClassroom.Code != http.StatusNotFound {
		t.Fatalf("invalid classroom status = %d, want %d body=%s", invalidClassroom.Code, http.StatusNotFound, invalidClassroom.Body.String())
	}

	createLesson := performRequest(t, app, http.MethodPost, "/api/teacher/lessons", []byte(`{"title":"Lesson","description":"","sort_order":1}`), teacherCookie)
	if createLesson.Code != http.StatusNotFound {
		t.Fatalf("teacher create lesson route status = %d, want %d body=%s", createLesson.Code, http.StatusNotFound, createLesson.Body.String())
	}

	createQuestion := performRequest(t, app, http.MethodPost, "/api/teacher/questions", []byte(`{"title":"Question"}`), teacherCookie)
	if createQuestion.Code != http.StatusNotFound {
		t.Fatalf("teacher create question route status = %d, want %d body=%s", createQuestion.Code, http.StatusNotFound, createQuestion.Body.String())
	}

	addLessonQuestion := performRequest(t, app, http.MethodPost, "/api/teacher/lessons/1/questions", []byte(`{"question_id":1,"sort_order":1}`), teacherCookie)
	if addLessonQuestion.Code != http.StatusNotFound {
		t.Fatalf("teacher add lesson question route status = %d, want %d body=%s", addLessonQuestion.Code, http.StatusNotFound, addLessonQuestion.Body.String())
	}
}

func TestStudentCurrentLessonRoutes(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherID := insertTeacher(t, app.db, "student_route_teacher")
	classroomID := insertClassroom(t, app.db, teacherID, "student_route_classroom")
	lessonID := insertLesson(t, app.db, "Student Lesson", "Current lesson for student", 1)
	questionOneID := insertQuestionWithContent(
		t,
		app.db,
		"Student Question 1",
		`{"Statement":"Describe question one","Input":"Sample input 1","Output":"Sample output 1"}`,
		"function solution()\n    return 1\nend",
		"function solution()\n    return 2\nend",
	)
	questionTwoID := insertQuestionWithContent(
		t,
		app.db,
		"Student Question 2",
		`{"Statement":"Describe question two","Input":"Sample input 2","Output":"Sample output 2"}`,
		"function solution()\n    return 3\nend",
		"function solution()\n    return 4\nend",
	)
	lessonQuestionOneID := insertLessonQuestion(t, app.db, lessonID, questionOneID, 1)
	lessonQuestionTwoID := insertLessonQuestion(t, app.db, lessonID, questionTwoID, 2)
	studentID := insertStudent(t, app.db, "student_route_user", "studentpw")
	insertEnrollment(t, app.db, classroomID, studentID, &lessonID)

	studentCookie := loginAs(t, app, "student_route_user", "studentpw")

	currentLesson := performRequest(t, app, http.MethodGet, "/api/student/current-lesson", nil, studentCookie)
	if currentLesson.Code != http.StatusOK {
		t.Fatalf("GET /api/student/current-lesson status = %d, want %d body=%s", currentLesson.Code, http.StatusOK, currentLesson.Body.String())
	}

	var lessonEnvelope struct {
		Data struct {
			Lesson struct {
				ID          int64  `json:"id"`
				Title       string `json:"title"`
				Description string `json:"description"`
				SortOrder   int    `json:"sort_order"`
				Questions   []struct {
					LessonQuestionID int64  `json:"lesson_question_id"`
					QuestionID       int64  `json:"question_id"`
					Title            string `json:"title"`
					SortOrder        int    `json:"sort_order"`
				} `json:"questions"`
			} `json:"lesson"`
		} `json:"data"`
	}
	decodeJSON(t, currentLesson.Body.Bytes(), &lessonEnvelope)
	if lessonEnvelope.Data.Lesson.ID != lessonID ||
		lessonEnvelope.Data.Lesson.Title != "Student Lesson" ||
		lessonEnvelope.Data.Lesson.Description != "Current lesson for student" ||
		lessonEnvelope.Data.Lesson.SortOrder != 1 {
		t.Fatalf("current lesson payload = %#v, want lesson metadata preserved", lessonEnvelope.Data.Lesson)
	}
	if len(lessonEnvelope.Data.Lesson.Questions) != 2 {
		t.Fatalf("current lesson questions len = %d, want %d", len(lessonEnvelope.Data.Lesson.Questions), 2)
	}
	if lessonEnvelope.Data.Lesson.Questions[0].LessonQuestionID != lessonQuestionOneID ||
		lessonEnvelope.Data.Lesson.Questions[0].QuestionID != questionOneID ||
		lessonEnvelope.Data.Lesson.Questions[0].SortOrder != 1 {
		t.Fatalf("current lesson first question = %#v, want lesson question one", lessonEnvelope.Data.Lesson.Questions[0])
	}
	if lessonEnvelope.Data.Lesson.Questions[1].LessonQuestionID != lessonQuestionTwoID ||
		lessonEnvelope.Data.Lesson.Questions[1].QuestionID != questionTwoID ||
		lessonEnvelope.Data.Lesson.Questions[1].SortOrder != 2 {
		t.Fatalf("current lesson second question = %#v, want lesson question two", lessonEnvelope.Data.Lesson.Questions[1])
	}

	questionDetail := performRequest(
		t,
		app,
		http.MethodGet,
		"/api/student/questions/"+itoa(lessonQuestionOneID),
		nil,
		studentCookie,
	)
	if questionDetail.Code != http.StatusOK {
		t.Fatalf("GET /api/student/questions/:lessonQuestionId status = %d, want %d body=%s", questionDetail.Code, http.StatusOK, questionDetail.Body.String())
	}

	var questionEnvelope struct {
		Data struct {
			Question map[string]any `json:"question"`
		} `json:"data"`
	}
	decodeJSON(t, questionDetail.Body.Bytes(), &questionEnvelope)
	if questionEnvelope.Data.Question["id"] != float64(questionOneID) {
		t.Fatalf("student question id = %v, want %d", questionEnvelope.Data.Question["id"], questionOneID)
	}
	if questionEnvelope.Data.Question["lesson_question_id"] != float64(lessonQuestionOneID) {
		t.Fatalf("student lesson_question_id = %v, want %d", questionEnvelope.Data.Question["lesson_question_id"], lessonQuestionOneID)
	}
	if questionEnvelope.Data.Question["title"] != "Student Question 1" {
		t.Fatalf("student question title = %v, want %q", questionEnvelope.Data.Question["title"], "Student Question 1")
	}
	if questionEnvelope.Data.Question["starter_code"] != "function solution()\n    return 1\nend" {
		t.Fatalf("student question starter_code = %v, want starter code preserved", questionEnvelope.Data.Question["starter_code"])
	}
	description, ok := questionEnvelope.Data.Question["description"].(map[string]any)
	if !ok {
		t.Fatalf("student question description = %#v, want JSON object", questionEnvelope.Data.Question["description"])
	}
	if description["Statement"] != "Describe question one" || description["Input"] != "Sample input 1" || description["Output"] != "Sample output 1" {
		t.Fatalf("student question description = %#v, want structured content", description)
	}
	if _, exists := questionEnvelope.Data.Question["reference_code"]; exists {
		t.Fatalf("student question unexpectedly exposed reference_code: %#v", questionEnvelope.Data.Question)
	}
	if _, exists := questionEnvelope.Data.Question["test_cases"]; exists {
		t.Fatalf("student question unexpectedly exposed test_cases: %#v", questionEnvelope.Data.Question)
	}
}

func TestStudentCurrentLessonRouteNotFoundWithoutCurrentLesson(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherID := insertTeacher(t, app.db, "student_no_lesson_teacher")
	classroomID := insertClassroom(t, app.db, teacherID, "student_no_lesson_classroom")
	studentID := insertStudent(t, app.db, "student_no_lesson_user", "studentpw")
	insertEnrollment(t, app.db, classroomID, studentID, nil)

	studentCookie := loginAs(t, app, "student_no_lesson_user", "studentpw")
	response := performRequest(t, app, http.MethodGet, "/api/student/current-lesson", nil, studentCookie)
	if response.Code != http.StatusNotFound {
		t.Fatalf("GET /api/student/current-lesson without current lesson status = %d, want %d body=%s", response.Code, http.StatusNotFound, response.Body.String())
	}
}

func TestStudentQuestionRouteRejectsQuestionOutsideCurrentLesson(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherID := insertTeacher(t, app.db, "student_question_scope_teacher")
	classroomID := insertClassroom(t, app.db, teacherID, "student_question_scope_classroom")
	currentLessonID := insertLesson(t, app.db, "Current Student Lesson", "", 1)
	otherLessonID := insertLesson(t, app.db, "Other Student Lesson", "", 2)
	currentQuestionID := insertQuestion(t, app.db, "Current Student Question")
	otherQuestionID := insertQuestion(t, app.db, "Other Student Question")
	insertLessonQuestion(t, app.db, currentLessonID, currentQuestionID, 1)
	otherLessonQuestionID := insertLessonQuestion(t, app.db, otherLessonID, otherQuestionID, 1)
	studentID := insertStudent(t, app.db, "student_question_scope_user", "studentpw")
	insertEnrollment(t, app.db, classroomID, studentID, &currentLessonID)

	studentCookie := loginAs(t, app, "student_question_scope_user", "studentpw")
	response := performRequest(
		t,
		app,
		http.MethodGet,
		"/api/student/questions/"+itoa(otherLessonQuestionID),
		nil,
		studentCookie,
	)
	if response.Code != http.StatusNotFound {
		t.Fatalf("GET /api/student/questions outside current lesson status = %d, want %d body=%s", response.Code, http.StatusNotFound, response.Body.String())
	}
}

func TestStudentCreateSubmissionRoute(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherID := insertTeacher(t, app.db, "student_submission_teacher")
	classroomID := insertClassroom(t, app.db, teacherID, "student_submission_classroom")
	lessonID := insertLesson(t, app.db, "Submission Lesson", "Current lesson", 1)
	questionID := insertQuestionWithContent(
		t,
		app.db,
		"Submission Question",
		`{"Statement":"Return 1"}`,
		"function solution()\n    return 0\nend",
		"function solution()\n    return 1\nend",
	)
	lessonQuestionID := insertLessonQuestion(t, app.db, lessonID, questionID, 1)
	studentID := insertStudent(t, app.db, "student_submission_user", "studentpw")
	enrollmentID := insertEnrollment(t, app.db, classroomID, studentID, &lessonID)

	studentCookie := loginAs(t, app, "student_submission_user", "studentpw")
	response := performRequest(t, app, http.MethodPost, "/api/student/submissions", []byte(`{
		"lesson_question_id": `+itoa(lessonQuestionID)+`,
		"source_code": "function solution()\n    return 1\nend"
	}`), studentCookie)
	if response.Code != http.StatusCreated {
		t.Fatalf("POST /api/student/submissions status = %d, want %d body=%s", response.Code, http.StatusCreated, response.Body.String())
	}

	var envelope struct {
		Data struct {
			Submission struct {
				ID               int64  `json:"id"`
				EnrollmentID     int64  `json:"enrollment_id"`
				LessonID         int64  `json:"lesson_id"`
				LessonQuestionID int64  `json:"lesson_question_id"`
				QuestionID       int64  `json:"question_id"`
				QuestionTitle    string `json:"question_title"`
				Status           string `json:"status"`
				SubmittedAt      string `json:"submitted_at"`
			} `json:"submission"`
		} `json:"data"`
	}
	decodeJSON(t, response.Body.Bytes(), &envelope)
	if envelope.Data.Submission.EnrollmentID != enrollmentID ||
		envelope.Data.Submission.LessonID != lessonID ||
		envelope.Data.Submission.LessonQuestionID != lessonQuestionID ||
		envelope.Data.Submission.QuestionID != questionID ||
		envelope.Data.Submission.QuestionTitle != "Submission Question" ||
		envelope.Data.Submission.Status != "pending" {
		t.Fatalf("created submission payload = %#v, want pending submission metadata", envelope.Data.Submission)
	}
	if envelope.Data.Submission.ID <= 0 || envelope.Data.Submission.SubmittedAt == "" {
		t.Fatalf("created submission id/submitted_at = %#v, want persisted values", envelope.Data.Submission)
	}

	var (
		actualEnrollmentID     int64
		actualLessonID         int64
		actualLessonQuestionID int64
		actualStatus           string
		actualSourceCode       string
		actualVerdict          sql.NullString
		actualFinishedAt       sql.NullString
	)
	if err := app.db.QueryRowContext(context.Background(), `
		SELECT enrollment_id, lesson_id, lesson_question_id, status, source_code, verdict, finished_at
		FROM submission
		WHERE id = ?
	`, envelope.Data.Submission.ID).Scan(
		&actualEnrollmentID,
		&actualLessonID,
		&actualLessonQuestionID,
		&actualStatus,
		&actualSourceCode,
		&actualVerdict,
		&actualFinishedAt,
	); err != nil {
		t.Fatalf("load created submission: %v", err)
	}
	if actualEnrollmentID != enrollmentID ||
		actualLessonID != lessonID ||
		actualLessonQuestionID != lessonQuestionID ||
		actualStatus != "pending" ||
		actualSourceCode != "function solution()\n    return 1\nend" {
		t.Fatalf(
			"created submission row = enrollment=%d lesson=%d lesson_question=%d status=%q source=%q",
			actualEnrollmentID,
			actualLessonID,
			actualLessonQuestionID,
			actualStatus,
			actualSourceCode,
		)
	}
	if actualVerdict.Valid {
		t.Fatalf("created submission verdict = %#v, want NULL before judging", actualVerdict)
	}
	if actualFinishedAt.Valid {
		t.Fatalf("created submission finished_at = %#v, want NULL before judging", actualFinishedAt)
	}
}

func TestStudentCreateSubmissionRouteValidatesRequest(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherID := insertTeacher(t, app.db, "submission_validate_teacher")
	classroomID := insertClassroom(t, app.db, teacherID, "submission_validate_room")
	lessonID := insertLesson(t, app.db, "Submission Validate Lesson", "", 1)
	questionID := insertQuestion(t, app.db, "Submission Validate Question")
	lessonQuestionID := insertLessonQuestion(t, app.db, lessonID, questionID, 1)
	studentID := insertStudent(t, app.db, "student_submission_validate_user", "studentpw")
	insertEnrollment(t, app.db, classroomID, studentID, &lessonID)

	studentCookie := loginAs(t, app, "student_submission_validate_user", "studentpw")
	response := performRequest(t, app, http.MethodPost, "/api/student/submissions", []byte(`{
		"lesson_question_id": `+itoa(lessonQuestionID)+`,
		"source_code": "   "
	}`), studentCookie)
	if response.Code != http.StatusNotFound {
		t.Fatalf("POST /api/student/submissions with blank source status = %d, want %d body=%s", response.Code, http.StatusNotFound, response.Body.String())
	}

	var submissionCount int
	if err := app.db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM submission`).Scan(&submissionCount); err != nil {
		t.Fatalf("count submissions: %v", err)
	}
	if submissionCount != 0 {
		t.Fatalf("submission count = %d, want 0 after validation failure", submissionCount)
	}
}

func TestStudentCreateSubmissionRouteRejectsQuestionOutsideCurrentLesson(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherID := insertTeacher(t, app.db, "student_submission_scope_teacher")
	classroomID := insertClassroom(t, app.db, teacherID, "student_submission_scope_classroom")
	currentLessonID := insertLesson(t, app.db, "Current Submission Lesson", "", 1)
	otherLessonID := insertLesson(t, app.db, "Other Submission Lesson", "", 2)
	currentQuestionID := insertQuestion(t, app.db, "Current Submission Question")
	otherQuestionID := insertQuestion(t, app.db, "Other Submission Question")
	insertLessonQuestion(t, app.db, currentLessonID, currentQuestionID, 1)
	otherLessonQuestionID := insertLessonQuestion(t, app.db, otherLessonID, otherQuestionID, 1)
	studentID := insertStudent(t, app.db, "student_submission_scope_user", "studentpw")
	insertEnrollment(t, app.db, classroomID, studentID, &currentLessonID)

	studentCookie := loginAs(t, app, "student_submission_scope_user", "studentpw")
	response := performRequest(t, app, http.MethodPost, "/api/student/submissions", []byte(`{
		"lesson_question_id": `+itoa(otherLessonQuestionID)+`,
		"source_code": "function solution()\n    return 1\nend"
	}`), studentCookie)
	if response.Code != http.StatusNotFound {
		t.Fatalf("POST /api/student/submissions outside current lesson status = %d, want %d body=%s", response.Code, http.StatusNotFound, response.Body.String())
	}

	var submissionCount int
	if err := app.db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM submission`).Scan(&submissionCount); err != nil {
		t.Fatalf("count submissions: %v", err)
	}
	if submissionCount != 0 {
		t.Fatalf("submission count = %d, want 0 after scope rejection", submissionCount)
	}
}

func TestStudentSubmissionReadRoutes(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherID := insertTeacher(t, app.db, "stu_sub_read_t")
	classroomID := insertClassroom(t, app.db, teacherID, "stu_sub_read_room")
	lessonID := insertLesson(t, app.db, "Student Submission Read Lesson", "", 1)
	questionID := insertQuestion(t, app.db, "Student Submission Read Question")
	lessonQuestionID := insertLessonQuestion(t, app.db, lessonID, questionID, 1)
	studentID := insertStudent(t, app.db, "stu_sub_read_u", "studentpw")
	enrollmentID := insertEnrollment(t, app.db, classroomID, studentID, &lessonID)
	otherStudentID := insertStudent(t, app.db, "stu_sub_read_o", "studentpw")
	otherEnrollmentID := insertEnrollment(t, app.db, classroomID, otherStudentID, &lessonID)

	firstSubmissionID := insertFinishedSubmission(
		t,
		app.db,
		enrollmentID,
		lessonID,
		lessonQuestionID,
		"wrong_answer",
		"function solution() return 0 end",
		"",
		`{"cases":[{"status":"wrong_answer"}]}`,
	)
	secondSubmissionID := insertFinishedSubmission(
		t,
		app.db,
		enrollmentID,
		lessonID,
		lessonQuestionID,
		"accepted",
		"function solution() return 1 end",
		`{"cases":[{"index":1,"stdout":"ok"}]}`,
		`{"cases":[{"status":"accepted"}]}`,
	)
	insertFinishedSubmission(
		t,
		app.db,
		otherEnrollmentID,
		lessonID,
		lessonQuestionID,
		"accepted",
		"function solution() return 2 end",
		"",
		`{"cases":[{"status":"accepted"}]}`,
	)

	studentCookie := loginAs(t, app, "stu_sub_read_u", "studentpw")

	listResponse := performRequest(t, app, http.MethodGet, "/api/student/submissions", nil, studentCookie)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("GET /api/student/submissions status = %d, want %d body=%s", listResponse.Code, http.StatusOK, listResponse.Body.String())
	}

	var listEnvelope struct {
		Data struct {
			Submissions []map[string]any `json:"submissions"`
		} `json:"data"`
	}
	decodeJSON(t, listResponse.Body.Bytes(), &listEnvelope)
	if len(listEnvelope.Data.Submissions) != 2 {
		t.Fatalf("student submission list len = %d, want 2", len(listEnvelope.Data.Submissions))
	}
	if listEnvelope.Data.Submissions[0]["id"] != float64(secondSubmissionID) || listEnvelope.Data.Submissions[1]["id"] != float64(firstSubmissionID) {
		t.Fatalf("student submission list order = %#v, want latest submission first", listEnvelope.Data.Submissions)
	}
	if _, exists := listEnvelope.Data.Submissions[0]["source_code"]; exists {
		t.Fatalf("student submission list unexpectedly exposed source_code: %#v", listEnvelope.Data.Submissions[0])
	}

	getResponse := performRequest(
		t,
		app,
		http.MethodGet,
		"/api/student/submissions/"+itoa(secondSubmissionID),
		nil,
		studentCookie,
	)
	if getResponse.Code != http.StatusOK {
		t.Fatalf("GET /api/student/submissions/:submissionId status = %d, want %d body=%s", getResponse.Code, http.StatusOK, getResponse.Body.String())
	}

	var getEnvelope struct {
		Data struct {
			Submission struct {
				ID            int64           `json:"id"`
				SourceCode    string          `json:"source_code"`
				Verdict       string          `json:"verdict"`
				StdoutBuffer  json.RawMessage `json:"stdout_buffer"`
				QuestionTitle string          `json:"question_title"`
			} `json:"submission"`
		} `json:"data"`
	}
	decodeJSON(t, getResponse.Body.Bytes(), &getEnvelope)
	if getEnvelope.Data.Submission.ID != secondSubmissionID ||
		getEnvelope.Data.Submission.SourceCode != "function solution() return 1 end" ||
		getEnvelope.Data.Submission.Verdict != "accepted" ||
		string(getEnvelope.Data.Submission.StdoutBuffer) != `{"cases":[{"index":1,"stdout":"ok"}]}` ||
		getEnvelope.Data.Submission.QuestionTitle != "Student Submission Read Question" {
		t.Fatalf("student submission detail = %#v, want full own submission detail", getEnvelope.Data.Submission)
	}
}

func TestStudentSubmissionReadRoutesRejectForeignSubmission(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherID := insertTeacher(t, app.db, "stu_sub_forbid_t")
	classroomID := insertClassroom(t, app.db, teacherID, "stu_sub_forbid_room")
	lessonID := insertLesson(t, app.db, "Student Submission Forbidden Lesson", "", 1)
	questionID := insertQuestion(t, app.db, "Student Submission Forbidden Question")
	lessonQuestionID := insertLessonQuestion(t, app.db, lessonID, questionID, 1)
	studentID := insertStudent(t, app.db, "stu_sub_forbid_u", "studentpw")
	insertEnrollment(t, app.db, classroomID, studentID, &lessonID)
	otherStudentID := insertStudent(t, app.db, "stu_sub_forbid_o", "studentpw")
	otherEnrollmentID := insertEnrollment(t, app.db, classroomID, otherStudentID, &lessonID)
	foreignSubmissionID := insertFinishedSubmission(
		t,
		app.db,
		otherEnrollmentID,
		lessonID,
		lessonQuestionID,
		"accepted",
		"function solution() return 2 end",
		"",
		`{"cases":[{"status":"accepted"}]}`,
	)

	studentCookie := loginAs(t, app, "stu_sub_forbid_u", "studentpw")
	response := performRequest(
		t,
		app,
		http.MethodGet,
		"/api/student/submissions/"+itoa(foreignSubmissionID),
		nil,
		studentCookie,
	)
	if response.Code != http.StatusNotFound {
		t.Fatalf("GET foreign /api/student/submissions/:submissionId status = %d, want %d body=%s", response.Code, http.StatusNotFound, response.Body.String())
	}
}

func TestStudentQuestionSubmissionRoute(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherID := insertTeacher(t, app.db, "stu_q_sub_t")
	classroomID := insertClassroom(t, app.db, teacherID, "stu_q_sub_room")
	currentLessonID := insertLesson(t, app.db, "Student Question Submission Lesson", "", 1)
	otherLessonID := insertLesson(t, app.db, "Student Question Submission Other Lesson", "", 2)
	currentQuestionID := insertQuestion(t, app.db, "Student Question Submission Current")
	currentOtherQuestionID := insertQuestion(t, app.db, "Student Question Submission Current Other")
	otherLessonQuestionID := insertLessonQuestion(t, app.db, otherLessonID, insertQuestion(t, app.db, "Student Question Submission Foreign Lesson"), 1)
	lessonQuestionID := insertLessonQuestion(t, app.db, currentLessonID, currentQuestionID, 1)
	currentOtherLessonQuestionID := insertLessonQuestion(t, app.db, currentLessonID, currentOtherQuestionID, 2)
	studentID := insertStudent(t, app.db, "stu_q_sub_u", "studentpw")
	enrollmentID := insertEnrollment(t, app.db, classroomID, studentID, &currentLessonID)
	otherStudentID := insertStudent(t, app.db, "stu_q_sub_o", "studentpw")
	otherEnrollmentID := insertEnrollment(t, app.db, classroomID, otherStudentID, &currentLessonID)

	firstSubmissionID := insertFinishedSubmission(
		t,
		app.db,
		enrollmentID,
		currentLessonID,
		lessonQuestionID,
		"wrong_answer",
		"function solution() return 0 end",
		"",
		`{"cases":[{"status":"wrong_answer"}]}`,
	)
	secondSubmissionID := insertFinishedSubmission(
		t,
		app.db,
		enrollmentID,
		currentLessonID,
		lessonQuestionID,
		"accepted",
		"function solution() return 1 end",
		`{"cases":[{"index":1,"stdout":"ok"}]}`,
		`{"cases":[{"status":"accepted"}]}`,
	)
	insertFinishedSubmission(
		t,
		app.db,
		enrollmentID,
		currentLessonID,
		currentOtherLessonQuestionID,
		"accepted",
		"function solution() return 2 end",
		"",
		`{"cases":[{"status":"accepted"}]}`,
	)
	insertFinishedSubmission(
		t,
		app.db,
		otherEnrollmentID,
		currentLessonID,
		lessonQuestionID,
		"accepted",
		"function solution() return 3 end",
		"",
		`{"cases":[{"status":"accepted"}]}`,
	)

	studentCookie := loginAs(t, app, "stu_q_sub_u", "studentpw")
	response := performRequest(
		t,
		app,
		http.MethodGet,
		"/api/student/questions/"+itoa(lessonQuestionID)+"/submissions",
		nil,
		studentCookie,
	)
	if response.Code != http.StatusOK {
		t.Fatalf("GET /api/student/questions/:lessonQuestionId/submissions status = %d, want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	var envelope struct {
		Data struct {
			Submissions []map[string]any `json:"submissions"`
		} `json:"data"`
	}
	decodeJSON(t, response.Body.Bytes(), &envelope)
	if len(envelope.Data.Submissions) != 2 {
		t.Fatalf("question submission list len = %d, want 2", len(envelope.Data.Submissions))
	}
	if envelope.Data.Submissions[0]["id"] != float64(secondSubmissionID) || envelope.Data.Submissions[1]["id"] != float64(firstSubmissionID) {
		t.Fatalf("question submission list order = %#v, want latest own submission first", envelope.Data.Submissions)
	}
	if envelope.Data.Submissions[0]["lesson_question_id"] != float64(lessonQuestionID) {
		t.Fatalf("question submission list lesson_question_id = %v, want %d", envelope.Data.Submissions[0]["lesson_question_id"], lessonQuestionID)
	}
	if _, exists := envelope.Data.Submissions[0]["source_code"]; exists {
		t.Fatalf("question submission list unexpectedly exposed source_code: %#v", envelope.Data.Submissions[0])
	}

	foreignLessonResponse := performRequest(
		t,
		app,
		http.MethodGet,
		"/api/student/questions/"+itoa(otherLessonQuestionID)+"/submissions",
		nil,
		studentCookie,
	)
	if foreignLessonResponse.Code != http.StatusNotFound {
		t.Fatalf("GET foreign lesson question submissions status = %d, want %d body=%s", foreignLessonResponse.Code, http.StatusNotFound, foreignLessonResponse.Body.String())
	}
}

func TestTeacherStudentManagementRoutes(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherCookie := loginAs(t, app, "teacher", "teacher")
	classroomID := insertClassroom(t, app.db, 1, "teacher_students")

	createStudent := performRequest(t, app, http.MethodPost, "/api/teacher/classrooms/"+itoa(classroomID)+"/students", []byte(`{
		"username":"student_manage_01",
		"password":"studentpw1"
	}`), teacherCookie)
	if createStudent.Code != http.StatusCreated {
		t.Fatalf("POST /api/teacher/classrooms/:id/students status = %d, want %d body=%s", createStudent.Code, http.StatusCreated, createStudent.Body.String())
	}

	var createStudentEnvelope struct {
		Data struct {
			Student struct {
				ID       int64  `json:"id"`
				Username string `json:"username"`
				Role     string `json:"role"`
				Status   string `json:"status"`
			} `json:"student"`
		} `json:"data"`
	}
	decodeJSON(t, createStudent.Body.Bytes(), &createStudentEnvelope)
	if createStudentEnvelope.Data.Student.Username != "student_manage_01" || createStudentEnvelope.Data.Student.Role != "student" || createStudentEnvelope.Data.Student.Status != "active" {
		t.Fatalf("created student payload = %#v, want active student account", createStudentEnvelope.Data.Student)
	}

	studentCookie := loginAs(t, app, "student_manage_01", "studentpw1")
	if studentCookie == nil {
		t.Fatal("new student should be able to login")
	}

	listStudents := performRequest(t, app, http.MethodGet, "/api/teacher/classrooms/"+itoa(classroomID)+"/students", nil, teacherCookie)
	if listStudents.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/classrooms/:id/students status = %d, want %d body=%s", listStudents.Code, http.StatusOK, listStudents.Body.String())
	}

	var listStudentsEnvelope struct {
		Data struct {
			Students []struct {
				ID       int64  `json:"id"`
				Username string `json:"username"`
			} `json:"students"`
		} `json:"data"`
	}
	decodeJSON(t, listStudents.Body.Bytes(), &listStudentsEnvelope)
	if len(listStudentsEnvelope.Data.Students) != 1 || listStudentsEnvelope.Data.Students[0].ID != createStudentEnvelope.Data.Student.ID {
		t.Fatalf("student list = %#v, want exactly created student", listStudentsEnvelope.Data.Students)
	}

	getStudent := performRequest(t, app, http.MethodGet, "/api/teacher/classrooms/"+itoa(classroomID)+"/students/"+itoa(createStudentEnvelope.Data.Student.ID), nil, teacherCookie)
	if getStudent.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/classrooms/:id/students/:studentId status = %d, want %d body=%s", getStudent.Code, http.StatusOK, getStudent.Body.String())
	}

	renameStudent := performRequest(t, app, http.MethodPatch, "/api/teacher/classrooms/"+itoa(classroomID)+"/students/"+itoa(createStudentEnvelope.Data.Student.ID)+"/name", []byte(`{
		"username":"student_manage_02"
	}`), teacherCookie)
	if renameStudent.Code != http.StatusOK {
		t.Fatalf("PATCH /api/teacher/classrooms/:id/students/:studentId/name status = %d, want %d body=%s", renameStudent.Code, http.StatusOK, renameStudent.Body.String())
	}

	loginAfterRename := loginAs(t, app, "student_manage_02", "studentpw1")
	if loginAfterRename == nil {
		t.Fatal("renamed student should still login with old password")
	}

	resetPassword := performRequest(t, app, http.MethodPost, "/api/teacher/classrooms/"+itoa(classroomID)+"/students/"+itoa(createStudentEnvelope.Data.Student.ID)+"/reset-password", []byte(`{
		"password":"studentpw2"
	}`), teacherCookie)
	if resetPassword.Code != http.StatusOK {
		t.Fatalf("POST /api/teacher/classrooms/:id/students/:studentId/reset-password status = %d, want %d body=%s", resetPassword.Code, http.StatusOK, resetPassword.Body.String())
	}

	oldPasswordLogin := performRequest(t, app, http.MethodPost, "/api/login", []byte(`{"username":"student_manage_02","password":"studentpw1"}`), nil)
	if oldPasswordLogin.Code != http.StatusUnauthorized {
		t.Fatalf("old password login status = %d, want %d body=%s", oldPasswordLogin.Code, http.StatusUnauthorized, oldPasswordLogin.Body.String())
	}

	newPasswordLogin := loginAs(t, app, "student_manage_02", "studentpw2")
	if newPasswordLogin == nil {
		t.Fatal("reset password should take effect")
	}

	removeStudent := performRequest(t, app, http.MethodDelete, "/api/teacher/classrooms/"+itoa(classroomID)+"/students/"+itoa(createStudentEnvelope.Data.Student.ID), nil, teacherCookie)
	if removeStudent.Code != http.StatusOK {
		t.Fatalf("DELETE /api/teacher/classrooms/:id/students/:studentId status = %d, want %d body=%s", removeStudent.Code, http.StatusOK, removeStudent.Body.String())
	}

	postRemoveGet := performRequest(t, app, http.MethodGet, "/api/teacher/classrooms/"+itoa(classroomID)+"/students/"+itoa(createStudentEnvelope.Data.Student.ID), nil, teacherCookie)
	if postRemoveGet.Code != http.StatusNotFound {
		t.Fatalf("GET removed student status = %d, want %d body=%s", postRemoveGet.Code, http.StatusNotFound, postRemoveGet.Body.String())
	}

	postRemoveLogin := performRequest(t, app, http.MethodPost, "/api/login", []byte(`{"username":"student_manage_02","password":"studentpw2"}`), nil)
	if postRemoveLogin.Code != http.StatusUnauthorized {
		t.Fatalf("removed student login status = %d, want %d body=%s", postRemoveLogin.Code, http.StatusUnauthorized, postRemoveLogin.Body.String())
	}
}

func TestTeacherLessonQuestionAndCurrentLessonRoutes(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherCookie := loginAs(t, app, "teacher", "teacher")
	teacherID := int64(1)
	classroomID := insertClassroom(t, app.db, teacherID, "teacher_lessons")
	studentID := insertStudent(t, app.db, "lesson_student_01", "studentpw")
	insertEnrollment(t, app.db, classroomID, studentID, nil)
	lessonOneID := insertLesson(t, app.db, "Teacher Lesson 1", "Lesson one", 1)
	lessonTwoID := insertLesson(t, app.db, "Teacher Lesson 2", "Lesson two", 2)
	questionID := insertQuestion(t, app.db, "Lesson Question 1")
	insertLessonQuestion(t, app.db, lessonOneID, questionID, 1)

	listLessonQuestions := performRequest(t, app, http.MethodGet, "/api/teacher/lessons/"+itoa(lessonOneID)+"/questions", nil, teacherCookie)
	if listLessonQuestions.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/lessons/:lessonId/questions status = %d, want %d body=%s", listLessonQuestions.Code, http.StatusOK, listLessonQuestions.Body.String())
	}

	var listLessonQuestionsEnvelope struct {
		Data struct {
			Questions []struct {
				ID         int64 `json:"id"`
				QuestionID int64 `json:"question_id"`
			} `json:"questions"`
		} `json:"data"`
	}
	decodeJSON(t, listLessonQuestions.Body.Bytes(), &listLessonQuestionsEnvelope)
	if len(listLessonQuestionsEnvelope.Data.Questions) != 1 || listLessonQuestionsEnvelope.Data.Questions[0].QuestionID != questionID {
		t.Fatalf("lesson question list = %#v, want inserted question", listLessonQuestionsEnvelope.Data.Questions)
	}

	setCurrentLesson := performRequest(t, app, http.MethodPost, "/api/teacher/classrooms/"+itoa(classroomID)+"/current-lesson", []byte(`{
		"lesson_id": `+itoa(lessonTwoID)+`
	}`), teacherCookie)
	if setCurrentLesson.Code != http.StatusOK {
		t.Fatalf("POST /api/teacher/classrooms/:id/current-lesson status = %d, want %d body=%s", setCurrentLesson.Code, http.StatusOK, setCurrentLesson.Body.String())
	}

	listClassroomLessons := performRequest(t, app, http.MethodGet, "/api/teacher/classrooms/"+itoa(classroomID)+"/lessons", nil, teacherCookie)
	if listClassroomLessons.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/classrooms/:id/lessons status = %d, want %d body=%s", listClassroomLessons.Code, http.StatusOK, listClassroomLessons.Body.String())
	}

	var listClassroomLessonsEnvelope struct {
		Data struct {
			Lessons []struct {
				ID        int64 `json:"id"`
				IsCurrent bool  `json:"is_current"`
			} `json:"lessons"`
		} `json:"data"`
	}
	decodeJSON(t, listClassroomLessons.Body.Bytes(), &listClassroomLessonsEnvelope)
	var currentFound bool
	for _, lesson := range listClassroomLessonsEnvelope.Data.Lessons {
		if lesson.ID == lessonTwoID && lesson.IsCurrent {
			currentFound = true
		}
		if lesson.ID == lessonOneID && lesson.IsCurrent {
			t.Fatalf("lesson one unexpectedly marked current: %#v", listClassroomLessonsEnvelope.Data.Lessons)
		}
	}
	if !currentFound {
		t.Fatalf("classroom lessons = %#v, want lesson two marked current", listClassroomLessonsEnvelope.Data.Lessons)
	}
}

func TestTeacherSetCurrentLessonPersistsForEmptyClassroom(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherCookie := loginAs(t, app, "teacher", "teacher")
	teacherID := int64(1)
	classroomID := insertClassroom(t, app.db, teacherID, "empty_current_lesson_classroom")
	lessonOneID := insertLesson(t, app.db, "Empty Class Lesson 1", "Lesson one", 1)
	lessonTwoID := insertLesson(t, app.db, "Empty Class Lesson 2", "Lesson two", 2)

	setCurrentLesson := performRequest(t, app, http.MethodPost, "/api/teacher/classrooms/"+itoa(classroomID)+"/current-lesson", []byte(`{
		"lesson_id": `+itoa(lessonTwoID)+`
	}`), teacherCookie)
	if setCurrentLesson.Code != http.StatusOK {
		t.Fatalf("POST /api/teacher/classrooms/:id/current-lesson for empty classroom status = %d, want %d body=%s", setCurrentLesson.Code, http.StatusOK, setCurrentLesson.Body.String())
	}

	listClassroomLessons := performRequest(t, app, http.MethodGet, "/api/teacher/classrooms/"+itoa(classroomID)+"/lessons", nil, teacherCookie)
	if listClassroomLessons.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/classrooms/:id/lessons for empty classroom status = %d, want %d body=%s", listClassroomLessons.Code, http.StatusOK, listClassroomLessons.Body.String())
	}

	var listClassroomLessonsEnvelope struct {
		Data struct {
			Lessons []struct {
				ID        int64 `json:"id"`
				IsCurrent bool  `json:"is_current"`
			} `json:"lessons"`
		} `json:"data"`
	}
	decodeJSON(t, listClassroomLessons.Body.Bytes(), &listClassroomLessonsEnvelope)

	var currentFound bool
	for _, lesson := range listClassroomLessonsEnvelope.Data.Lessons {
		if lesson.ID == lessonTwoID && lesson.IsCurrent {
			currentFound = true
		}
		if lesson.ID == lessonOneID && lesson.IsCurrent {
			t.Fatalf("lesson one unexpectedly marked current in empty classroom: %#v", listClassroomLessonsEnvelope.Data.Lessons)
		}
	}
	if !currentFound {
		t.Fatalf("empty classroom lessons = %#v, want lesson two marked current", listClassroomLessonsEnvelope.Data.Lessons)
	}

	studentID := insertStudent(t, app.db, "empty_class_student_01", "studentpw")
	insertEnrollment(t, app.db, classroomID, studentID, nil)

	var enrolledCurrentLessonID int64
	if err := app.db.QueryRowContext(context.Background(), `
		SELECT current_lesson_id
		FROM enrollment
		WHERE classroom_id = ? AND student_id = ?
	`, classroomID, studentID).Scan(&enrolledCurrentLessonID); err != nil {
		t.Fatalf("load enrollment current lesson: %v", err)
	}
	if enrolledCurrentLessonID != lessonTwoID {
		t.Fatalf("enrollment current lesson = %d, want %d", enrolledCurrentLessonID, lessonTwoID)
	}
}

func TestTeacherProgressAndSubmissionRoutes(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherCookie := loginAs(t, app, "teacher", "teacher")
	teacherID := int64(1)
	classroomID := insertClassroom(t, app.db, teacherID, "teacher_progress")
	lessonID := insertLesson(t, app.db, "Progress Lesson", "Progress lesson", 1)
	questionID := insertQuestion(t, app.db, "Progress Question")
	lessonQuestionID := insertLessonQuestion(t, app.db, lessonID, questionID, 1)
	studentID := insertStudent(t, app.db, "progress_student_01", "studentpw")
	enrollmentID := insertEnrollment(t, app.db, classroomID, studentID, &lessonID)
	submissionID := insertFinishedSubmission(t, app.db, enrollmentID, lessonID, lessonQuestionID, "accepted", "function solution() return 1 end", "", `{"cases":[]}`)

	progressResponse := performRequest(t, app, http.MethodGet, "/api/teacher/classrooms/"+itoa(classroomID)+"/progress", nil, teacherCookie)
	if progressResponse.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/classrooms/:id/progress status = %d, want %d body=%s", progressResponse.Code, http.StatusOK, progressResponse.Body.String())
	}

	var progressEnvelope struct {
		Data struct {
			Progress struct {
				Students []struct {
					ID             int64  `json:"id"`
					Username       string `json:"username"`
					LessonProgress struct {
						Accepted int `json:"accepted"`
						Total    int `json:"total"`
					} `json:"lesson_progress"`
					Latest *struct {
						ID            int64  `json:"id"`
						Verdict       string `json:"verdict"`
						QuestionTitle string `json:"question_title"`
					} `json:"latest_submission"`
				} `json:"students"`
			} `json:"progress"`
		} `json:"data"`
	}
	decodeJSON(t, progressResponse.Body.Bytes(), &progressEnvelope)
	if len(progressEnvelope.Data.Progress.Students) != 1 {
		t.Fatalf("progress students len = %d, want 1", len(progressEnvelope.Data.Progress.Students))
	}
	progressStudent := progressEnvelope.Data.Progress.Students[0]
	if progressStudent.Username != "progress_student_01" || progressStudent.LessonProgress.Accepted != 1 || progressStudent.LessonProgress.Total != 1 {
		t.Fatalf("progress student payload = %#v, want accepted progress", progressStudent)
	}
	if progressStudent.Latest == nil || progressStudent.Latest.ID != submissionID || progressStudent.Latest.Verdict != "accepted" {
		t.Fatalf("progress latest submission = %#v, want inserted submission", progressStudent.Latest)
	}

	otherLessonID := insertLesson(t, app.db, "Progress Other Lesson", "Progress other lesson", 2)
	otherQuestionID := insertQuestion(t, app.db, "Progress Other Question")
	otherLessonQuestionID := insertLessonQuestion(t, app.db, otherLessonID, otherQuestionID, 1)
	if _, err := app.db.ExecContext(context.Background(), `
		UPDATE classroom
		SET current_lesson_id = ?
		WHERE id = ?
	`, otherLessonID, classroomID); err != nil {
		t.Fatalf("temporarily update classroom current lesson: %v", err)
	}
	otherLessonSubmissionID := insertFinishedSubmission(
		t,
		app.db,
		enrollmentID,
		otherLessonID,
		otherLessonQuestionID,
		"accepted",
		"function solution() return 2 end",
		"",
		`{"cases":[]}`,
	)
	if _, err := app.db.ExecContext(context.Background(), `
		UPDATE classroom
		SET current_lesson_id = ?
		WHERE id = ?
	`, lessonID, classroomID); err != nil {
		t.Fatalf("restore classroom current lesson: %v", err)
	}
	latestCurrentLessonSubmissionID := insertFinishedSubmission(
		t,
		app.db,
		enrollmentID,
		lessonID,
		lessonQuestionID,
		"wrong_answer",
		"function solution() return 0 end",
		"",
		`{"cases":[]}`,
	)

	listSubmissions := performRequest(t, app, http.MethodGet, "/api/teacher/classrooms/"+itoa(classroomID)+"/submissions", nil, teacherCookie)
	if listSubmissions.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/classrooms/:id/submissions status = %d, want %d body=%s", listSubmissions.Code, http.StatusOK, listSubmissions.Body.String())
	}

	var submissionsEnvelope struct {
		Data struct {
			Submissions []struct {
				ID            int64  `json:"id"`
				StudentID     int64  `json:"student_id"`
				QuestionTitle string `json:"question_title"`
			} `json:"submissions"`
		} `json:"data"`
	}
	decodeJSON(t, listSubmissions.Body.Bytes(), &submissionsEnvelope)
	if len(submissionsEnvelope.Data.Submissions) != 2 {
		t.Fatalf("submission list len = %d, want current lesson submissions only: %#v", len(submissionsEnvelope.Data.Submissions), submissionsEnvelope.Data.Submissions)
	}
	if submissionsEnvelope.Data.Submissions[0].ID != latestCurrentLessonSubmissionID || submissionsEnvelope.Data.Submissions[1].ID != submissionID {
		t.Fatalf("submission list order = %#v, want current lesson submissions ordered latest first", submissionsEnvelope.Data.Submissions)
	}
	for _, submission := range submissionsEnvelope.Data.Submissions {
		if submission.ID == otherLessonSubmissionID {
			t.Fatalf("submission list included other lesson submission: %#v", submissionsEnvelope.Data.Submissions)
		}
		if submission.StudentID != studentID {
			t.Fatalf("submission list student_id = %d, want %d", submission.StudentID, studentID)
		}
	}

	getSubmission := performRequest(t, app, http.MethodGet, "/api/teacher/classrooms/"+itoa(classroomID)+"/submissions/"+itoa(submissionID), nil, teacherCookie)
	if getSubmission.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/classrooms/:id/submissions/:submissionId status = %d, want %d body=%s", getSubmission.Code, http.StatusOK, getSubmission.Body.String())
	}

	var getSubmissionEnvelope struct {
		Data struct {
			Submission struct {
				ID         int64  `json:"id"`
				SourceCode string `json:"source_code"`
				Verdict    string `json:"verdict"`
			} `json:"submission"`
		} `json:"data"`
	}
	decodeJSON(t, getSubmission.Body.Bytes(), &getSubmissionEnvelope)
	if getSubmissionEnvelope.Data.Submission.ID != submissionID || getSubmissionEnvelope.Data.Submission.SourceCode != "function solution() return 1 end" || getSubmissionEnvelope.Data.Submission.Verdict != "accepted" {
		t.Fatalf("submission detail payload = %#v, want full submission detail", getSubmissionEnvelope.Data.Submission)
	}

	deleteSubmission := performRequest(t, app, http.MethodDelete, "/api/teacher/classrooms/"+itoa(classroomID)+"/submissions/"+itoa(submissionID), nil, teacherCookie)
	if deleteSubmission.Code != http.StatusOK {
		t.Fatalf("DELETE /api/teacher/classrooms/:id/submissions/:submissionId status = %d, want %d body=%s", deleteSubmission.Code, http.StatusOK, deleteSubmission.Body.String())
	}

	getDeletedSubmission := performRequest(t, app, http.MethodGet, "/api/teacher/classrooms/"+itoa(classroomID)+"/submissions/"+itoa(submissionID), nil, teacherCookie)
	if getDeletedSubmission.Code != http.StatusNotFound {
		t.Fatalf("GET deleted submission status = %d, want %d body=%s", getDeletedSubmission.Code, http.StatusNotFound, getDeletedSubmission.Body.String())
	}
}

func newTestApp(t *testing.T) *App {
	t.Helper()

	cfg := config.Config{
		App: config.AppConfig{
			Name: "oj-lite-test",
			Env:  "test",
		},
		HTTP: config.HTTPConfig{
			Host:            "127.0.0.1",
			Port:            0,
			GinMode:         "test",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     5 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
		DB: config.DBConfig{
			Path:        filepath.Join(t.TempDir(), "oj-lite.db"),
			BusyTimeout: 5 * time.Second,
		},
	}

	database, err := platformdb.Open(context.Background(), cfg.DB)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	if err := platformdb.Migrate(context.Background(), database); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	if err := seed.SeedDemoAccounts(context.Background(), database); err != nil {
		t.Fatalf("seed test database: %v", err)
	}

	sessions, err := session.NewManager()
	if err != nil {
		t.Fatalf("create session manager: %v", err)
	}

	return NewApp(cfg, logger.NewLogger("app-test"), database, sessions)
}

func shutdownTestApp(t *testing.T, app *App) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown test app: %v", err)
	}
}

func loginAs(t *testing.T, app *App, username, password string) *http.Cookie {
	t.Helper()

	body := []byte(`{"username":"` + username + `","password":"` + password + `"}`)
	response := performRequest(t, app, http.MethodPost, "/api/login", body, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == session.DefaultCookieName {
			return cookie
		}
	}

	t.Fatal("login response missing api session cookie")
	return nil
}

func performRequest(t *testing.T, app *App, method, path string, body []byte, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader(body)
	}

	request := httptest.NewRequest(method, path, reader)
	request.RemoteAddr = "127.0.0.1:12345"
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		request.AddCookie(cookie)
	}

	response := httptest.NewRecorder()
	app.Router().ServeHTTP(response, request)
	return response
}

func decodeJSON(t *testing.T, body []byte, target any) {
	t.Helper()

	if err := json.Unmarshal(body, target); err != nil {
		t.Fatalf("decode json: %v body=%s", err, string(body))
	}
}

func containsNamedClassroom(items []struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}, id int64, name string) bool {
	for _, item := range items {
		if item.ID == id && item.Name == name {
			return true
		}
	}

	return false
}

func containsNamedLesson(items []struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}, id int64, title string) bool {
	for _, item := range items {
		if item.ID == id && item.Title == title {
			return true
		}
	}

	return false
}

func containsNamedQuestion(items []struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}, id int64, title string) bool {
	for _, item := range items {
		if item.ID == id && item.Title == title {
			return true
		}
	}

	return false
}

func insertTeacher(t *testing.T, database *sql.DB, username string) int64 {
	t.Helper()

	result, err := database.ExecContext(context.Background(), `
		INSERT INTO user_account (username, password_hash, role, status)
		VALUES (?, ?, 'teacher', 'active')
	`, username, "hash")
	if err != nil {
		t.Fatalf("insert teacher: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("teacher last insert id: %v", err)
	}

	return id
}

func insertStudent(t *testing.T, database *sql.DB, username, rawPassword string) int64 {
	t.Helper()

	passwordHash, err := platformpassword.Hash(rawPassword)
	if err != nil {
		t.Fatalf("hash student password: %v", err)
	}

	result, err := database.ExecContext(context.Background(), `
		INSERT INTO user_account (username, password_hash, role, status)
		VALUES (?, ?, 'student', 'active')
	`, username, passwordHash)
	if err != nil {
		t.Fatalf("insert student: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("student last insert id: %v", err)
	}

	return id
}

func insertClassroom(t *testing.T, database *sql.DB, teacherID int64, name string) int64 {
	t.Helper()

	result, err := database.ExecContext(context.Background(), `
		INSERT INTO classroom (teacher_id, name)
		VALUES (?, ?)
	`, teacherID, name)
	if err != nil {
		t.Fatalf("insert classroom: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("classroom last insert id: %v", err)
	}

	return id
}

func insertLesson(t *testing.T, database *sql.DB, title, description string, sortOrder int) int64 {
	t.Helper()

	result, err := database.ExecContext(context.Background(), `
		INSERT INTO lesson (title, description, sort_order)
		VALUES (?, ?, ?)
	`, title, description, sortOrder)
	if err != nil {
		t.Fatalf("insert lesson: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("lesson last insert id: %v", err)
	}

	return id
}

func insertQuestion(t *testing.T, database *sql.DB, title string) int64 {
	t.Helper()

	return insertQuestionWithContent(
		t,
		database,
		title,
		`{}`,
		"function solution()\n    return 0\nend",
		"function solution()\n    return 1\nend",
	)
}

func insertQuestionWithContent(
	t *testing.T,
	database *sql.DB,
	title, description, starterCode, referenceCode string,
) int64 {
	t.Helper()

	result, err := database.ExecContext(context.Background(), `
		INSERT INTO question (title, description, starter_code, reference_code, test_cases)
		VALUES (?, ?, ?, ?, ?)
	`, title, description, starterCode, referenceCode, `[{"input":[1]}]`)
	if err != nil {
		t.Fatalf("insert question: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("question last insert id: %v", err)
	}

	return id
}

func insertLessonQuestion(t *testing.T, database *sql.DB, lessonID, questionID int64, sortOrder int) int64 {
	t.Helper()

	result, err := database.ExecContext(context.Background(), `
		INSERT INTO lesson_question (lesson_id, question_id, sort_order)
		VALUES (?, ?, ?)
	`, lessonID, questionID, sortOrder)
	if err != nil {
		t.Fatalf("insert lesson_question: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("lesson_question last insert id: %v", err)
	}

	return id
}

func insertEnrollment(t *testing.T, database *sql.DB, classroomID, studentID int64, currentLessonID *int64) int64 {
	t.Helper()

	if currentLessonID != nil {
		if _, err := database.ExecContext(context.Background(), `
			UPDATE classroom
			SET current_lesson_id = ?
			WHERE id = ?
		`, *currentLessonID, classroomID); err != nil {
			t.Fatalf("update classroom current lesson: %v", err)
		}
	} else {
		var classroomCurrentLessonID sql.NullInt64
		if err := database.QueryRowContext(context.Background(), `
			SELECT current_lesson_id
			FROM classroom
			WHERE id = ?
		`, classroomID).Scan(&classroomCurrentLessonID); err != nil {
			t.Fatalf("load classroom current lesson: %v", err)
		}
		if classroomCurrentLessonID.Valid {
			currentLessonID = &classroomCurrentLessonID.Int64
		}
	}

	var (
		result sql.Result
		err    error
	)
	if currentLessonID == nil {
		result, err = database.ExecContext(context.Background(), `
			INSERT INTO enrollment (classroom_id, student_id)
			VALUES (?, ?)
		`, classroomID, studentID)
	} else {
		result, err = database.ExecContext(context.Background(), `
			INSERT INTO enrollment (classroom_id, student_id, current_lesson_id)
			VALUES (?, ?, ?)
		`, classroomID, studentID, *currentLessonID)
	}
	if err != nil {
		t.Fatalf("insert enrollment: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("enrollment last insert id: %v", err)
	}

	return id
}

func insertFinishedSubmission(
	t *testing.T,
	database *sql.DB,
	enrollmentID, lessonID, lessonQuestionID int64,
	verdict, sourceCode, stdoutBuffer, judgeReport string,
) int64 {
	t.Helper()

	result, err := database.ExecContext(context.Background(), `
		INSERT INTO submission (
			enrollment_id,
			lesson_id,
			lesson_question_id,
			status,
			verdict,
			source_code,
			stdout_buffer,
			judge_report,
			submitted_at,
			finished_at
		)
		VALUES (
			?, ?, ?, 'finished', ?, ?, ?, ?,
			strftime('%Y-%m-%dT%H:%M:%fZ', 'now'),
			strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		)
	`, enrollmentID, lessonID, lessonQuestionID, verdict, sourceCode, stdoutBuffer, judgeReport)
	if err != nil {
		t.Fatalf("insert submission: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("submission last insert id: %v", err)
	}

	return id
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}
