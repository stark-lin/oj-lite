package app

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminRoutesRejectNonLoopback(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	request := httptest.NewRequest(http.MethodGet, "/admin/teachers", bytes.NewReader(nil))
	request.RemoteAddr = "203.0.113.5:23456"

	response := httptest.NewRecorder()
	app.Router().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("GET /admin/teachers from non-loopback status = %d, want %d body=%s", response.Code, http.StatusNotFound, response.Body.String())
	}
}

func TestAdminPageRoute(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	response := performRequest(t, app, http.MethodGet, "/admin", nil, nil)
	if response.Code != http.StatusOK {
		t.Fatalf("GET /admin status = %d, want %d body=%s", response.Code, http.StatusOK, response.Body.String())
	}

	body := response.Body.String()
	if !bytes.Contains([]byte(body), []byte("Admin")) || !bytes.Contains([]byte(body), []byte("Teacher Management")) {
		t.Fatalf("GET /admin body missing expected markers: %s", body)
	}
}

func TestAdminTeacherRoutes(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	createTeacher := performRequest(t, app, http.MethodPost, "/admin/teachers", []byte(`{
		"username":"admin_teacher_01",
		"password":"teacherpw1"
	}`), nil)
	if createTeacher.Code != http.StatusCreated {
		t.Fatalf("POST /admin/teachers status = %d, want %d body=%s", createTeacher.Code, http.StatusCreated, createTeacher.Body.String())
	}

	var createTeacherEnvelope struct {
		Data struct {
			Teacher struct {
				ID       int64  `json:"id"`
				Username string `json:"username"`
				Status   string `json:"status"`
			} `json:"teacher"`
		} `json:"data"`
	}
	decodeJSON(t, createTeacher.Body.Bytes(), &createTeacherEnvelope)
	if createTeacherEnvelope.Data.Teacher.Username != "admin_teacher_01" || createTeacherEnvelope.Data.Teacher.Status != "active" {
		t.Fatalf("created teacher payload = %#v, want active teacher", createTeacherEnvelope.Data.Teacher)
	}

	loginAs(t, app, "admin_teacher_01", "teacherpw1")

	listTeachers := performRequest(t, app, http.MethodGet, "/admin/teachers", nil, nil)
	if listTeachers.Code != http.StatusOK {
		t.Fatalf("GET /admin/teachers status = %d, want %d body=%s", listTeachers.Code, http.StatusOK, listTeachers.Body.String())
	}

	var listTeachersEnvelope struct {
		Data struct {
			Teachers []struct {
				ID       int64  `json:"id"`
				Username string `json:"username"`
			} `json:"teachers"`
		} `json:"data"`
	}
	decodeJSON(t, listTeachers.Body.Bytes(), &listTeachersEnvelope)
	if !containsNamedTeacher(listTeachersEnvelope.Data.Teachers, createTeacherEnvelope.Data.Teacher.ID, "admin_teacher_01") {
		t.Fatalf("teacher list missing created teacher: %#v", listTeachersEnvelope.Data.Teachers)
	}

	getTeacher := performRequest(t, app, http.MethodGet, "/admin/teachers/"+itoa(createTeacherEnvelope.Data.Teacher.ID), nil, nil)
	if getTeacher.Code != http.StatusOK {
		t.Fatalf("GET /admin/teachers/:teacherId status = %d, want %d body=%s", getTeacher.Code, http.StatusOK, getTeacher.Body.String())
	}

	renameTeacher := performRequest(t, app, http.MethodPatch, "/admin/teachers/"+itoa(createTeacherEnvelope.Data.Teacher.ID), []byte(`{
		"username":"admin_teacher_02"
	}`), nil)
	if renameTeacher.Code != http.StatusOK {
		t.Fatalf("PATCH /admin/teachers/:teacherId status = %d, want %d body=%s", renameTeacher.Code, http.StatusOK, renameTeacher.Body.String())
	}

	resetPassword := performRequest(t, app, http.MethodPost, "/admin/teachers/"+itoa(createTeacherEnvelope.Data.Teacher.ID)+"/reset-password", []byte(`{
		"password":"teacherpw2"
	}`), nil)
	if resetPassword.Code != http.StatusOK {
		t.Fatalf("POST /admin/teachers/:teacherId/reset-password status = %d, want %d body=%s", resetPassword.Code, http.StatusOK, resetPassword.Body.String())
	}

	oldPasswordLogin := performRequest(t, app, http.MethodPost, "/api/login", []byte(`{"username":"admin_teacher_02","password":"teacherpw1"}`), nil)
	if oldPasswordLogin.Code != http.StatusUnauthorized {
		t.Fatalf("old password login status = %d, want %d body=%s", oldPasswordLogin.Code, http.StatusUnauthorized, oldPasswordLogin.Body.String())
	}

	loginAs(t, app, "admin_teacher_02", "teacherpw2")

	insertClassroom(t, app.db, createTeacherEnvelope.Data.Teacher.ID, "admin_teacher_room")

	deleteTeacher := performRequest(t, app, http.MethodDelete, "/admin/teachers/"+itoa(createTeacherEnvelope.Data.Teacher.ID), nil, nil)
	if deleteTeacher.Code != http.StatusOK {
		t.Fatalf("DELETE /admin/teachers/:teacherId status = %d, want %d body=%s", deleteTeacher.Code, http.StatusOK, deleteTeacher.Body.String())
	}

	getDeletedTeacher := performRequest(t, app, http.MethodGet, "/admin/teachers/"+itoa(createTeacherEnvelope.Data.Teacher.ID), nil, nil)
	if getDeletedTeacher.Code != http.StatusOK {
		t.Fatalf("GET disabled teacher status = %d, want %d body=%s", getDeletedTeacher.Code, http.StatusOK, getDeletedTeacher.Body.String())
	}

	var getDeletedTeacherEnvelope struct {
		Data struct {
			Teacher struct {
				Status string `json:"status"`
			} `json:"teacher"`
		} `json:"data"`
	}
	decodeJSON(t, getDeletedTeacher.Body.Bytes(), &getDeletedTeacherEnvelope)
	if getDeletedTeacherEnvelope.Data.Teacher.Status != "disabled" {
		t.Fatalf("deleted teacher status = %q, want %q", getDeletedTeacherEnvelope.Data.Teacher.Status, "disabled")
	}

	disabledLogin := performRequest(t, app, http.MethodPost, "/api/login", []byte(`{"username":"admin_teacher_02","password":"teacherpw2"}`), nil)
	if disabledLogin.Code != http.StatusUnauthorized {
		t.Fatalf("disabled teacher login status = %d, want %d body=%s", disabledLogin.Code, http.StatusUnauthorized, disabledLogin.Body.String())
	}
}

func TestAdminLessonRoutes(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	createLesson := performRequest(t, app, http.MethodPost, "/admin/lessons", []byte(`{
		"title":"Admin Lesson 1",
		"description":"Lesson one",
		"sort_order":1,
		"questions":[
			{
				"title":"Question A",
				"description":{"statement":"A"},
				"starter_code":"function solution() return 0 end",
				"reference_code":"function solution() return 1 end",
				"test_cases":[{"input":[1]}],
				"sort_order":1
			},
			{
				"title":"Question B",
				"description":{"statement":"B"},
				"starter_code":"function solution() return 2 end",
				"reference_code":"function solution() return 3 end",
				"test_cases":[{"input":[2]}],
				"sort_order":2
			}
		]
	}`), nil)
	if createLesson.Code != http.StatusCreated {
		t.Fatalf("POST /admin/lessons status = %d, want %d body=%s", createLesson.Code, http.StatusCreated, createLesson.Body.String())
	}

	var createLessonEnvelope struct {
		Data struct {
			Lesson struct {
				ID        int64 `json:"id"`
				Questions []struct {
					LessonQuestionID int64  `json:"lesson_question_id"`
					ID               int64  `json:"id"`
					Title            string `json:"title"`
					SortOrder        int    `json:"sort_order"`
				} `json:"questions"`
			} `json:"lesson"`
		} `json:"data"`
	}
	decodeJSON(t, createLesson.Body.Bytes(), &createLessonEnvelope)
	if len(createLessonEnvelope.Data.Lesson.Questions) != 2 {
		t.Fatalf("created lesson questions len = %d, want 2", len(createLessonEnvelope.Data.Lesson.Questions))
	}
	firstQuestionID := createLessonEnvelope.Data.Lesson.Questions[0].ID
	secondQuestionID := createLessonEnvelope.Data.Lesson.Questions[1].ID

	listLessons := performRequest(t, app, http.MethodGet, "/admin/lessons", nil, nil)
	if listLessons.Code != http.StatusOK {
		t.Fatalf("GET /admin/lessons status = %d, want %d body=%s", listLessons.Code, http.StatusOK, listLessons.Body.String())
	}

	var listLessonsEnvelope struct {
		Data struct {
			Lessons []struct {
				ID        int64 `json:"id"`
				Questions []struct {
					ID int64 `json:"id"`
				} `json:"questions"`
			} `json:"lessons"`
		} `json:"data"`
	}
	decodeJSON(t, listLessons.Body.Bytes(), &listLessonsEnvelope)

	var listedLesson *struct {
		ID        int64 `json:"id"`
		Questions []struct {
			ID int64 `json:"id"`
		} `json:"questions"`
	}
	for index := range listLessonsEnvelope.Data.Lessons {
		item := &listLessonsEnvelope.Data.Lessons[index]
		if item.ID == createLessonEnvelope.Data.Lesson.ID {
			listedLesson = item
			break
		}
	}
	if listedLesson == nil || len(listedLesson.Questions) != 2 {
		t.Fatalf("admin lesson list missing created lesson: %#v", listLessonsEnvelope.Data.Lessons)
	}

	teacherCookie := loginAs(t, app, "teacher", "teacher")
	teacherLessons := performRequest(t, app, http.MethodGet, "/api/teacher/lessons", nil, teacherCookie)
	if teacherLessons.Code != http.StatusOK {
		t.Fatalf("GET /api/teacher/lessons status = %d, want %d body=%s", teacherLessons.Code, http.StatusOK, teacherLessons.Body.String())
	}

	replaceLesson := performRequest(t, app, http.MethodPut, "/admin/lessons/"+itoa(createLessonEnvelope.Data.Lesson.ID), []byte(`{
		"title":"Admin Lesson 1 Updated",
		"description":"Lesson one updated",
		"sort_order":2,
		"questions":[
			{
				"id":`+itoa(firstQuestionID)+`,
				"title":"Question A Updated",
				"description":{"statement":"A+"},
				"starter_code":"function solution() return 4 end",
				"reference_code":"function solution() return 5 end",
				"test_cases":[{"input":[4]}],
				"sort_order":2
			},
			{
				"title":"Question C",
				"description":{"statement":"C"},
				"starter_code":"function solution() return 6 end",
				"reference_code":"function solution() return 7 end",
				"test_cases":[{"input":[6]}],
				"sort_order":1
			}
		]
	}`), nil)
	if replaceLesson.Code != http.StatusOK {
		t.Fatalf("PUT /admin/lessons/:lessonId status = %d, want %d body=%s", replaceLesson.Code, http.StatusOK, replaceLesson.Body.String())
	}

	var replaceLessonEnvelope struct {
		Data struct {
			Lesson struct {
				Title     string `json:"title"`
				SortOrder int    `json:"sort_order"`
				Questions []struct {
					ID        int64  `json:"id"`
					Title     string `json:"title"`
					SortOrder int    `json:"sort_order"`
				} `json:"questions"`
			} `json:"lesson"`
		} `json:"data"`
	}
	decodeJSON(t, replaceLesson.Body.Bytes(), &replaceLessonEnvelope)
	if replaceLessonEnvelope.Data.Lesson.Title != "Admin Lesson 1 Updated" || replaceLessonEnvelope.Data.Lesson.SortOrder != 2 {
		t.Fatalf("replaced lesson payload = %#v, want updated title/sort_order", replaceLessonEnvelope.Data.Lesson)
	}
	if len(replaceLessonEnvelope.Data.Lesson.Questions) != 2 {
		t.Fatalf("replaced lesson questions len = %d, want 2", len(replaceLessonEnvelope.Data.Lesson.Questions))
	}
	if replaceLessonEnvelope.Data.Lesson.Questions[0].Title != "Question C" || replaceLessonEnvelope.Data.Lesson.Questions[1].ID != firstQuestionID {
		t.Fatalf("replaced lesson question order = %#v, want new question first and updated old question second", replaceLessonEnvelope.Data.Lesson.Questions)
	}

	var removedQuestionCount int
	if err := app.db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM question WHERE id = ?`, secondQuestionID).Scan(&removedQuestionCount); err != nil {
		t.Fatalf("count removed question: %v", err)
	}
	if removedQuestionCount != 0 {
		t.Fatalf("removed question count = %d, want 0", removedQuestionCount)
	}

	deleteLesson := performRequest(t, app, http.MethodDelete, "/admin/lessons/"+itoa(createLessonEnvelope.Data.Lesson.ID), nil, nil)
	if deleteLesson.Code != http.StatusOK {
		t.Fatalf("DELETE /admin/lessons/:lessonId status = %d, want %d body=%s", deleteLesson.Code, http.StatusOK, deleteLesson.Body.String())
	}

	getDeletedLesson := performRequest(t, app, http.MethodGet, "/admin/lessons/"+itoa(createLessonEnvelope.Data.Lesson.ID), nil, nil)
	if getDeletedLesson.Code != http.StatusNotFound {
		t.Fatalf("GET deleted /admin/lessons/:lessonId status = %d, want %d body=%s", getDeletedLesson.Code, http.StatusNotFound, getDeletedLesson.Body.String())
	}
}

func TestAdminDeleteLessonRejectsReferencedLesson(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	teacherID := insertTeacher(t, app.db, "admin_delete_lesson_teacher")
	classroomID := insertClassroom(t, app.db, teacherID, "admin_delete_lesson_room")
	lessonID := insertLesson(t, app.db, "Referenced Lesson", "", 1)
	questionID := insertQuestion(t, app.db, "Referenced Lesson Question")
	insertLessonQuestion(t, app.db, lessonID, questionID, 1)

	if _, err := app.db.ExecContext(context.Background(), `
		UPDATE classroom
		SET current_lesson_id = ?
		WHERE id = ?
	`, lessonID, classroomID); err != nil {
		t.Fatalf("set classroom current lesson: %v", err)
	}

	deleteLesson := performRequest(t, app, http.MethodDelete, "/admin/lessons/"+itoa(lessonID), nil, nil)
	if deleteLesson.Code != http.StatusBadRequest {
		t.Fatalf("DELETE referenced /admin/lessons/:lessonId status = %d, want %d body=%s", deleteLesson.Code, http.StatusBadRequest, deleteLesson.Body.String())
	}
}

func containsNamedTeacher(items []struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}, id int64, username string) bool {
	for _, item := range items {
		if item.ID == id && item.Username == username {
			return true
		}
	}

	return false
}

func ensureQuestionAbsent(t *testing.T, database *sql.DB, questionID int64) {
	t.Helper()

	var count int
	if err := database.QueryRowContext(context.Background(), `
		SELECT COUNT(*)
		FROM question
		WHERE id = ?
	`, questionID).Scan(&count); err != nil {
		t.Fatalf("count question rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("question %d still exists", questionID)
	}
}
