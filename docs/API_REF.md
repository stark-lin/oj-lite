# API Contract

This document describes the behavior of the currently registered and usable interfaces in the codebase. For the full route list, see [API_DESCRIBE.md](API_DESCRIBE.md).

## General Conventions

### Protocol And Response Wrapper

- HTTP + JSON.
- Encoding: `UTF-8`.
- Teacher and student business APIs use Cookie Session authentication.
- The admin side is currently a loopback-only local management entry point and does not use a separate admin session.

Successful JSON responses use this wrapper:

```json
{
  "data": {}
}
```

Error responses use this wrapper:

```json
{
  "error": {
    "code": "validation_error",
    "message": "username already exists",
    "details": {
      "field": "username"
    }
  },
  "request_id": "..."
}
```

Matched JSON API responses include `request_id` when the global request-id middleware has assigned one. The examples below omit `request_id` unless it matters. Requests that do not match any Gin route are returned by the framework as a default `404` and are not guaranteed to use this JSON wrapper.

### Current Error Codes

Currently registered endpoints may return:

- `bad_request`
- `validation_error`
- `unauthorized`
- `not_found`
- `internal_error`

The `not_implemented` helper still exists in the codebase, but no currently registered public scaffold route returns `501`.

### Error Semantics

Business APIs under `/api/*`:

- `401 unauthorized`: not logged in, invalid or expired session, or current user no longer exists.
- `404 not_found`: authenticated user cannot see, use, or perform the requested operation; path or body parameters are invalid; or a business precondition is not met.
- `500 internal_error`: database, judge, scheduler, or code failure.

Public login and local admin APIs:

- Invalid JSON, field validation failures, and path-parameter parse failures usually return `400`.
- Missing or unavailable objects return `404`.
- Non-loopback access to `/admin/*` returns `404`.

### Business Session

- Cookie name: `ojlite_session`.
- `HttpOnly`.
- `SameSite=Lax`.
- TTL: 8 hours.
- Automatically refreshed when the remaining lifetime is 2 hours or less.
- Existing sessions become invalid after service restart because the signing/encryption key is generated on process startup.

## System And Pages

### `GET /healthz`

```json
{
  "data": {
    "service": "oj-lite",
    "env": "local",
    "status": "ok"
  }
}
```

### Pages And Assets

- `GET /`: login page.
- `GET /admin`: local admin page; loopback only.
- `GET /teacher`: teacher page; requires a valid teacher session, otherwise redirects to `/`.
- `GET /student`: student page; requires a valid student session, otherwise redirects to `/`.
- `GET /assets/app.css`: embedded page stylesheet.
- `GET /assets/app.js`: embedded page script.

## Auth API

### `POST /api/login`

Description: login for teacher and student business users.

Request body:

```json
{
  "username": "student",
  "password": "student"
}
```

Validation:

- `username` is trimmed and must not be empty.
- `password` length must be `7..128`.

Successful response:

```json
{
  "data": {
    "user": {
      "id": 2,
      "username": "student",
      "role": "student",
      "status": "active",
      "created_at": "2026-04-18T00:00:00.000Z"
    }
  }
}
```

Errors:

- `400 validation_error`: request fields are invalid.
- `401 unauthorized`: username or password is wrong, or the account is disabled.

### `POST /api/logout`

Description: clear the current business session cookie. This endpoint does not require an authenticated session.

```json
{
  "data": {
    "ok": true
  }
}
```

### `GET /api/me`

Description: read the current logged-in user.

Errors:

- `401 unauthorized`: missing, invalid, or expired session, or current user no longer exists.

### `POST /api/me/password`

Description: change the current logged-in user's password.

Request body:

```json
{
  "old_password": "student",
  "new_password": "student123"
}
```

Validation:

- `old_password` length must be `7..128`.
- `new_password` length must be `7..128`.
- The new password must differ from the old password.

Successful response:

```json
{
  "data": {
    "ok": true
  }
}
```

Errors:

- `401 unauthorized`: missing, invalid, or expired session, or current user no longer exists.
- `404 not_found`: old password is incorrect, current account is unavailable, request body is not valid JSON, or parameters are invalid.

## Admin API

Every `/admin/*` API only accepts loopback requests. There is no separate admin user table or admin session. `AdminAuth` treats local loopback requests as `admin`.

### `POST /admin/login`

Local admin entry probe. The current implementation does not validate credentials and does not create an admin session.

```json
{
  "data": {
    "ok": true
  }
}
```

### `POST /admin/logout`

Local admin exit probe.

```json
{
  "data": {
    "ok": true
  }
}
```

### Teacher Management

#### `POST /admin/teachers`

Request body:

```json
{
  "username": "teacher_01",
  "password": "teacherpw1"
}
```

Validation:

- `username` is trimmed and length must be `3..32`.
- `username` may only contain `A-Za-z0-9_`.
- `username` must be unique in `user_account`.
- `password` length must be `7..128`.

Successful response: `201 Created`

```json
{
  "data": {
    "teacher": {
      "id": 3,
      "username": "teacher_01",
      "role": "teacher",
      "status": "active",
      "created_at": "2026-04-19T00:00:00.000Z"
    }
  }
}
```

#### `GET /admin/teachers`

Description: list all teachers ordered by `id ASC`.

```json
{
  "data": {
    "teachers": [
      {
        "id": 1,
        "username": "teacher",
        "role": "teacher",
        "status": "active",
        "created_at": "2026-04-19T00:00:00.000Z"
      }
    ]
  }
}
```

#### `GET /admin/teachers/:teacherId`

Description: read one teacher. A non-integer `:teacherId` returns `400 bad_request`; a missing teacher returns `404 not_found`.

#### `PATCH /admin/teachers/:teacherId`

Request body:

```json
{
  "username": "teacher_02",
  "status": "disabled"
}
```

Notes:

- `username` and `status` are both optional.
- `username` uses the same validation rules as teacher creation.
- `status` must be `active` or `disabled`.

Successful response returns the updated `teacher`.

#### `POST /admin/teachers/:teacherId/reset-password`

Request body:

```json
{
  "password": "teacherpw2"
}
```

Successful response:

```json
{
  "data": {
    "ok": true
  }
}
```

#### `DELETE /admin/teachers/:teacherId`

Description:

- If the teacher has no classroom, the `user_account` row is deleted.
- If the teacher already has a classroom, the teacher status is set to `disabled`.

Successful response:

```json
{
  "data": {
    "ok": true
  }
}
```

### Lesson JSON Management

Admin is the only current write entry point for lessons and questions. A lesson write contains lesson metadata and its questions; the system updates `lesson`, `question`, and `lesson_question` together.

#### Lesson Write Request Body

`POST /admin/lessons` and `PUT /admin/lessons/:lessonId` use the same structure:

```json
{
  "title": "Lesson 1",
  "description": "Intro lesson",
  "sort_order": 1,
  "questions": [
    {
      "id": 10,
      "title": "Sum",
      "description": {
        "statement": "Return the sum of two numbers.",
        "input": "Two numbers a and b.",
        "output": "a + b"
      },
      "starter_code": "function solution(a, b)\n    return 0\nend",
      "reference_code": "function solution(a, b)\n    return a + b\nend",
      "test_cases": [
        [1, 2],
        { "input": [3, 4] },
        { "args": [10, -3] }
      ],
      "sort_order": 1
    }
  ]
}
```

Notes:

- Omit `id` or send `0` when creating a new question.
- When replacing a lesson, any question with an `id` must already belong to that lesson.
- When replacing a lesson, old questions missing from the request are removed from that lesson. If a removed question is no longer used by any lesson, it is deleted from `question`.
- Question `description` must be a JSON object.
- `test_cases` must be valid JSON. The judge supports array cases, `{"input":[...]}`, `{"args":[...]}`, and scalar single-value cases.

Validation:

- Lesson `title` is trimmed and must not be empty.
- Lesson `sort_order` must be greater than `0` and globally unique.
- Question `title` is trimmed and must not be empty.
- Question `sort_order` must be greater than `0` and unique within the lesson.
- Question `id` values cannot repeat within the same request.
- After a question has any submission, its core judging fields are frozen by a database trigger.

#### Lesson Response Structure

```json
{
  "data": {
    "lesson": {
      "id": 1,
      "title": "Lesson 1",
      "description": "Intro lesson",
      "sort_order": 1,
      "created_at": "2026-04-19T00:00:00.000Z",
      "questions": [
        {
          "lesson_question_id": 100,
          "id": 10,
          "title": "Sum",
          "description": {
            "statement": "Return the sum of two numbers."
          },
          "starter_code": "function solution(a, b)\n    return 0\nend",
          "reference_code": "function solution(a, b)\n    return a + b\nend",
          "test_cases": [
            [1, 2]
          ],
          "sort_order": 1,
          "created_at": "2026-04-19T00:00:00.000Z"
        }
      ]
    }
  }
}
```

#### `POST /admin/lessons`

Description: create a lesson and its questions. Successful response: `201 Created`.

#### `GET /admin/lessons`

Description: list all lessons ordered by `sort_order ASC, id ASC`, including each lesson's questions.

#### `GET /admin/lessons/:lessonId`

Description: read one lesson, including questions.

#### `PUT /admin/lessons/:lessonId`

Description: replace the whole lesson and its questions. Successful response returns the full replaced lesson.

#### `DELETE /admin/lessons/:lessonId`

Description: delete a lesson that is not referenced by classroom, enrollment, submission, or other protected objects. Referenced lessons return `400 validation_error`.

Successful response:

```json
{
  "data": {
    "ok": true
  }
}
```

## Teacher API

Every `/api/teacher/*` endpoint requires a valid business session with role `teacher`.

When this is not satisfied:

- `401 unauthorized`: missing, invalid, or expired session.
- `404 not_found`: current session role is not `teacher`.

### Classroom

#### `POST /api/teacher/classrooms`

Request body:

```json
{
  "name": "classroom_a"
}
```

Validation:

- `name` is trimmed and must not be empty.

Successful response: `201 Created`

```json
{
  "data": {
    "classroom": {
      "id": 1,
      "name": "classroom_a",
      "created_at": "2026-04-19T00:00:00.000Z"
    }
  }
}
```

#### `GET /api/teacher/classrooms`

Description: list the current teacher's classrooms ordered by `classroom.id ASC`.

#### `GET /api/teacher/classrooms/:classroomId`

Description: read one classroom owned by the current teacher. Ownership is checked by `id + teacher_id`.

### Student Management

#### `POST /api/teacher/classrooms/:classroomId/students`

Request body:

```json
{
  "username": "student_02",
  "password": "student123"
}
```

Successful response: `201 Created`

```json
{
  "data": {
    "student": {
      "id": 2,
      "username": "student_02",
      "role": "student",
      "status": "active",
      "created_at": "2026-04-19T00:00:00.000Z"
    }
  }
}
```

Validation:

- `username` is trimmed and length must be `3..32`.
- `username` may only contain `A-Za-z0-9_`.
- `username` must be unique in `user_account`.
- `password` length must be `7..128`.
- The classroom must belong to the current teacher.

If the target classroom has no current lesson, the service tries to set it to the first global lesson by `sort_order ASC, id ASC` before creating the enrollment.

#### `GET /api/teacher/classrooms/:classroomId/students`

Description: list students in the classroom ordered by `user_account.id ASC`.

#### `GET /api/teacher/classrooms/:classroomId/students/:studentId`

Description: read one student in the classroom.

#### `PATCH /api/teacher/classrooms/:classroomId/students/:studentId/name`

Request body:

```json
{
  "username": "student_03"
}
```

Description: rename a student in the teacher's own classroom. The same username rules used for student creation apply.

#### `POST /api/teacher/classrooms/:classroomId/students/:studentId/reset-password`

Request body:

```json
{
  "password": "studentpw2"
}
```

Validation:

- The student must belong to the teacher's classroom.
- `password` length must be `7..128`.

#### `DELETE /api/teacher/classrooms/:classroomId/students/:studentId`

Description: remove the student's enrollment from the classroom. If submissions already reference that enrollment, database foreign-key constraints may currently surface as `500 internal_error`.

### Read-only Content APIs

The teacher side currently reads global lessons and questions. Writes are handled by admin lesson JSON APIs.

#### `GET /api/teacher/lessons`

Description: list all global lessons ordered by `sort_order ASC, id ASC`.

#### `GET /api/teacher/lessons/:lessonId`

Description: read one global lesson. Lessons are global resources and are not isolated by teacher.

#### `GET /api/teacher/questions`

Description: list all global questions ordered by `question.id ASC`.

#### `GET /api/teacher/questions/:questionId`

Description: read one global question. Questions are global resources and are not isolated by teacher.

#### `GET /api/teacher/lessons/:lessonId/questions`

Description: list questions for one lesson.

```json
{
  "data": {
    "questions": [
      {
        "id": 100,
        "lesson_id": 1,
        "question_id": 10,
        "title": "Sum",
        "sort_order": 1,
        "created_at": "2026-04-19T00:00:00.000Z"
      }
    ]
  }
}
```

### Classroom Lesson And Progress

#### `GET /api/teacher/classrooms/:classroomId/lessons`

Description: list global lessons available to the classroom with an `is_current` marker.

```json
{
  "data": {
    "lessons": [
      {
        "id": 1,
        "title": "Lesson 1",
        "description": "Intro lesson",
        "sort_order": 1,
        "created_at": "2026-04-19T00:00:00.000Z",
        "is_current": true
      }
    ]
  }
}
```

#### `POST /api/teacher/classrooms/:classroomId/current-lesson`

Request body:

```json
{
  "lesson_id": 1
}
```

Description: set the classroom current lesson and synchronize `current_lesson_id` for enrollments in that classroom.

Successful response:

```json
{
  "data": {
    "current_lesson": {
      "id": 1,
      "title": "Lesson 1",
      "description": "Intro lesson",
      "sort_order": 1,
      "created_at": "2026-04-19T00:00:00.000Z",
      "is_current": true
    }
  }
}
```

#### `GET /api/teacher/classrooms/:classroomId/progress`

Description: read each student's current-lesson completion state and latest submission summary for a classroom.

```json
{
  "data": {
    "progress": {
      "students": [
        {
          "id": 2,
          "username": "student",
          "role": "student",
          "status": "active",
          "created_at": "2026-04-19T00:00:00.000Z",
          "current_lesson": {
            "id": 1,
            "title": "Lesson 1"
          },
          "lesson_progress": {
            "accepted": 1,
            "total": 2
          },
          "latest_submission": {
            "id": 10,
            "enrollment_id": 5,
            "student_id": 2,
            "student_username": "student",
            "lesson_id": 1,
            "lesson_question_id": 100,
            "question_id": 10,
            "question_title": "Sum",
            "status": "finished",
            "verdict": "accepted",
            "submitted_at": "2026-04-19T00:00:00.000Z",
            "finished_at": "2026-04-19T00:00:01.000Z"
          }
        }
      ]
    }
  }
}
```

### Teacher Submission Queries

#### `GET /api/teacher/classrooms/:classroomId/submissions`

Description: list classroom submissions ordered by `submitted_at DESC, id DESC`. The list response does not include `source_code`.

#### `GET /api/teacher/classrooms/:classroomId/submissions/:submissionId`

Description: read one classroom submission. The detail response includes `source_code`.

#### `DELETE /api/teacher/classrooms/:classroomId/submissions/:submissionId`

Description: delete one classroom submission.

Successful response:

```json
{
  "data": {
    "ok": true
  }
}
```

## Student API

Every `/api/student/*` endpoint requires a valid business session with role `student`.

When this is not satisfied:

- `401 unauthorized`: missing, invalid, or expired session.
- `404 not_found`: current session role is not `student`.

### `GET /api/student/current-lesson`

Description: read the student's current lesson. Students cannot freely browse all lessons.

```json
{
  "data": {
    "lesson": {
      "id": 1,
      "title": "Lesson 1",
      "description": "Intro lesson",
      "sort_order": 1,
      "created_at": "2026-04-19T00:00:00.000Z",
      "questions": [
        {
          "lesson_question_id": 100,
          "question_id": 10,
          "title": "Sum",
          "sort_order": 1
        }
      ]
    }
  }
}
```

### `GET /api/student/questions/:lessonQuestionId`

Description: read one question in the current lesson. Only student-visible fields are returned; `reference_code` and `test_cases` are not returned.

```json
{
  "data": {
    "question": {
      "id": 10,
      "lesson_question_id": 100,
      "title": "Sum",
      "description": {
        "statement": "Return the sum of two numbers."
      },
      "starter_code": "function solution(a, b)\n    return 0\nend",
      "sort_order": 1,
      "created_at": "2026-04-19T00:00:00.000Z"
    }
  }
}
```

### `GET /api/student/questions/:lessonQuestionId/submissions`

Description: list the current student's submissions for one question in the current lesson, ordered by `submitted_at DESC, id DESC`. The list response does not include `source_code`.

### `POST /api/student/submissions`

Request body:

```json
{
  "lesson_question_id": 100,
  "source_code": "function solution(a, b)\n    return a + b\nend"
}
```

Description:

- `lesson_question_id` must belong to the current lesson.
- `source_code` is trimmed for validation and must not be empty.
- `source_code` maximum size is `65535` bytes.
- A newly created submission starts with `status` set to `pending`.

Successful response: `201 Created`

```json
{
  "data": {
    "submission": {
      "id": 10,
      "enrollment_id": 5,
      "lesson_id": 1,
      "lesson_question_id": 100,
      "question_id": 10,
      "question_title": "Sum",
      "status": "pending",
      "submitted_at": "2026-04-19T00:00:00.000Z"
    }
  }
}
```

### `GET /api/student/submissions`

Description: list the current student's submissions ordered by `submitted_at DESC, id DESC`. The list response does not include `source_code`.

### `GET /api/student/submissions/:submissionId`

Description: read one current-student submission. The detail response includes `source_code`.

## Submission And Judge Fields

Common submission response fields:

```json
{
  "id": 10,
  "enrollment_id": 5,
  "lesson_id": 1,
  "lesson_question_id": 100,
  "question_id": 10,
  "question_title": "Sum",
  "status": "finished",
  "verdict": "accepted",
  "source_code": "function solution(a, b) return a + b end",
  "stdout_buffer": {
    "cases": [
      {
        "index": 1,
        "stdout": "debug output"
      }
    ]
  },
  "error_message": "",
  "judge_report": {
    "cases": [
      {
        "index": 1,
        "input": [1, 2],
        "comparison": {
          "matched": true
        },
        "reference": {
          "returnValues": [3]
        },
        "student": {
          "returnValues": [3]
        }
      }
    ]
  },
  "submitted_at": "2026-04-19T00:00:00.000Z",
  "finished_at": "2026-04-19T00:00:01.000Z"
}
```

Notes:

- `status` currently allows `pending`, `judging`, and `finished`.
- `verdict` currently allows `accepted`, `wrong_answer`, `runtime_error`, and `system_error`.
- List endpoints usually omit `source_code`.
- `stdout_buffer` only appears when student code writes stdout.
- `error_message` only appears for runtime or system errors.
- `judge_report` only appears after judging has finished and the report was successfully written.
