# Route Overview

This document lists only the routes that are currently registered by the repository, and whether they are usable in the current implementation.

Status legend:

- `available`: the route has implemented behavior.
- `local_only`: the route has implemented behavior but only accepts loopback requests.

There are no currently registered public routes that still return `501 not_implemented`.

## Pages, Assets, And System Endpoints

| Method | Path | Status | Description |
| --- | --- | --- | --- |
| `GET` | `/` | `available` | Login page |
| `GET` | `/admin` | `local_only` | Local admin page |
| `GET` | `/teacher` | `available` | Teacher page; requires a valid teacher session |
| `GET` | `/student` | `available` | Student page; requires a valid student session |
| `GET` | `/assets/app.css` | `available` | Embedded page stylesheet |
| `GET` | `/assets/app.js` | `available` | Embedded page script |
| `GET` | `/healthz` | `available` | Health check |

## Auth

| Method | Path | Status | Description |
| --- | --- | --- | --- |
| `POST` | `/api/login` | `available` | Teacher / student login |
| `POST` | `/api/logout` | `available` | Clear the business session |
| `GET` | `/api/me` | `available` | Read the current user |
| `POST` | `/api/me/password` | `available` | Change the current user's password |

## Admin

`/admin/*` is protected by `AdminAuth` and only accepts requests from loopback IP addresses. There is no separate admin session in the current implementation. `POST /admin/login` and `POST /admin/logout` are lightweight probe endpoints used by the local admin page and return `{ "ok": true }`.

| Method | Path | Status | Description |
| --- | --- | --- | --- |
| `POST` | `/admin/login` | `local_only` | Local admin entry probe |
| `POST` | `/admin/logout` | `local_only` | Local admin exit probe |
| `POST` | `/admin/teachers` | `local_only` | Create a teacher |
| `GET` | `/admin/teachers` | `local_only` | List teachers |
| `GET` | `/admin/teachers/:teacherId` | `local_only` | Read one teacher |
| `PATCH` | `/admin/teachers/:teacherId` | `local_only` | Update a teacher username or status |
| `POST` | `/admin/teachers/:teacherId/reset-password` | `local_only` | Reset a teacher password |
| `DELETE` | `/admin/teachers/:teacherId` | `local_only` | Delete or disable a teacher |
| `POST` | `/admin/lessons` | `local_only` | Create a lesson and its questions |
| `GET` | `/admin/lessons` | `local_only` | List lessons |
| `GET` | `/admin/lessons/:lessonId` | `local_only` | Read one lesson |
| `PUT` | `/admin/lessons/:lessonId` | `local_only` | Replace a lesson and its questions |
| `DELETE` | `/admin/lessons/:lessonId` | `local_only` | Delete an unreferenced lesson |

## Teacher

The teacher side owns classroom operations and read-only content access. Lesson and question creation/editing are currently centralized in the local admin lesson JSON API.

| Method | Path | Status | Description |
| --- | --- | --- | --- |
| `POST` | `/api/teacher/classrooms` | `available` | Create a classroom |
| `GET` | `/api/teacher/classrooms` | `available` | List classrooms |
| `GET` | `/api/teacher/classrooms/:classroomId` | `available` | Read one classroom |
| `POST` | `/api/teacher/classrooms/:classroomId/students` | `available` | Create a student and add the student to a classroom |
| `GET` | `/api/teacher/classrooms/:classroomId/students` | `available` | List classroom students |
| `GET` | `/api/teacher/classrooms/:classroomId/students/:studentId` | `available` | Read one student |
| `PATCH` | `/api/teacher/classrooms/:classroomId/students/:studentId/name` | `available` | Rename a student |
| `POST` | `/api/teacher/classrooms/:classroomId/students/:studentId/reset-password` | `available` | Reset a student password |
| `DELETE` | `/api/teacher/classrooms/:classroomId/students/:studentId` | `available` | Remove a student from a classroom |
| `GET` | `/api/teacher/lessons` | `available` | List global lessons |
| `GET` | `/api/teacher/lessons/:lessonId` | `available` | Read one global lesson |
| `GET` | `/api/teacher/questions` | `available` | List global questions |
| `GET` | `/api/teacher/questions/:questionId` | `available` | Read one global question |
| `GET` | `/api/teacher/lessons/:lessonId/questions` | `available` | List questions for one lesson |
| `GET` | `/api/teacher/classrooms/:classroomId/lessons` | `available` | List lessons available to the classroom with current-lesson markers |
| `POST` | `/api/teacher/classrooms/:classroomId/current-lesson` | `available` | Set a classroom's current lesson |
| `GET` | `/api/teacher/classrooms/:classroomId/progress` | `available` | Read classroom progress |
| `GET` | `/api/teacher/classrooms/:classroomId/submissions` | `available` | List classroom submissions |
| `GET` | `/api/teacher/classrooms/:classroomId/submissions/:submissionId` | `available` | Read one classroom submission |
| `DELETE` | `/api/teacher/classrooms/:classroomId/submissions/:submissionId` | `available` | Delete one classroom submission |

Legacy teacher write routes that are not currently registered:

- `POST /api/teacher/lessons`
- `POST /api/teacher/questions`
- `POST /api/teacher/lessons/:lessonId/questions`

Those paths are handled as ordinary unmatched routes.

## Student

| Method | Path | Status | Description |
| --- | --- | --- | --- |
| `GET` | `/api/student/current-lesson` | `available` | Read the current lesson |
| `GET` | `/api/student/questions/:lessonQuestionId` | `available` | Read a question in the current lesson |
| `GET` | `/api/student/questions/:lessonQuestionId/submissions` | `available` | List the current student's submissions for one question |
| `POST` | `/api/student/submissions` | `available` | Create a submission |
| `GET` | `/api/student/submissions` | `available` | List the current student's submissions |
| `GET` | `/api/student/submissions/:submissionId` | `available` | Read one current-student submission |
