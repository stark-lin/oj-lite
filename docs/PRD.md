# PRD: oj-lite MVP

This document keeps the product scope, boundaries, and current MVP behavior. It intentionally avoids duplicating the full API field contract and repository layout details.

## Product Positioning

`oj-lite` is a classroom-assisted programming practice system. It is not an open problem bank, a contest platform, or an LMS.

The core workflow is intentionally narrow:

1. A local admin manages teacher accounts and global lesson JSON.
2. Teachers create classrooms and students.
3. Teachers advance each classroom to its current lesson.
4. Students only access the current lesson.
5. Students submit Lua code.
6. The system judges submissions asynchronously and returns results.
7. Teachers review classroom progress and submissions.

## Roles

### Admin

The admin is a local system-management entry point and does not participate in classroom activity.

Responsibilities:

- Create teachers.
- Update teacher username or status.
- Reset teacher passwords.
- Delete teachers without classrooms, or disable teachers that already own classrooms.
- Create, replace, and delete unreferenced global lessons.
- Maintain lesson questions through lesson JSON.

The current admin entry point is loopback-only. It does not have a separate admin session or public-network admin authentication flow.

### Teacher

Teachers own classroom activity.

Responsibilities:

- Create classrooms.
- Create students and add them to their own classrooms.
- Read global lessons, questions, and lesson-question lists.
- Set the current lesson for their own classrooms.
- View progress and submissions for their own classrooms.
- Delete submissions in their own classrooms.

Teachers currently do not create or edit lessons and questions. Content writes are centralized in the local admin lesson JSON API.

### Student

Students have minimal learning permissions.

Responsibilities:

- View the current lesson.
- View questions in the current lesson.
- Submit Lua code.
- View their own submission results.

## Current MVP Capabilities

- Go monolith.
- SQLite storage.
- Cookie-session authentication for teacher and student users.
- Teacher and student business login, logout, `GET /api/me`, and password changes.
- Local admin page and teacher / lesson JSON management APIs.
- Demo data initialization.
- Classroom model: `classroom`, `lesson`, `question`, `lesson_question`, `enrollment`, and `submission`.
- Lua function-based judging.
- Asynchronous submission scheduling.
- Protected teacher and student pages.
- Local admin page.

## Non-goals

The current scope does not include:

- Open public problem-bank behavior.
- Student browsing of every lesson.
- Student self-selection of lessons.
- Multi-language judging.
- Traditional stdin/stdout online judge mode.
- Multi-tenant organization management.
- Complex teacher collaboration.
- Public-internet security hardening.
- Distributed workers or complex orchestration.
- A full admin login system.

## Product Principles

- The teacher controls pacing; students do not freely explore the content tree.
- Classrooms use the global lesson list and advance through a shared order.
- The student view is always scoped to the current lesson.
- A question is a complete judgeable unit.
- Submission creation and judging are decoupled through an asynchronous flow.
- Lesson and question writes are centralized in local admin; the teacher side is read-only for content.
- Low deployment and maintenance complexity takes priority over breadth.
- The permission model focuses on preventing student privilege escalation.

## Core Objects

- `UserAccount`
- `Classroom`
- `Enrollment`
- `Lesson`
- `Question`
- `LessonQuestion`
- `Submission`

## Judging Model

The system uses Lua function-based judging. Student code is expected to define:

```lua
function solution(...)
    -- student code
end
```

Each question stores:

- `description`
- `starter_code`
- `reference_code`
- `test_cases`

`description` is stored and returned as a JSON object. `test_cases` is stored as JSON text. The judge runner supports array cases, `{"input":[...]}`, `{"args":[...]}`, and scalar single-value cases.

Judging flow:

1. Run `reference_code` against the same `test_cases` to generate expected results.
2. Run the student's `source_code` against the same `test_cases`.
3. Compare the reference and student return values.
4. Produce a verdict, stdout buffer, error message, and judge report.
5. Write the final result back to the submission.

## Acceptance Criteria

The current MVP should satisfy:

- A local admin can create teachers and maintain global lesson JSON with questions.
- A teacher can create a classroom, create students, advance the shared lesson plan, and view progress.
- A student can only access the current lesson.
- A student can submit code and receive an asynchronous judge result.
- A teacher can view classroom-level submissions and completion state.
- The service can run independently on a local machine or LAN.

## Current Implementation Notes

The repository currently includes service startup, database initialization, session handling, login/logout, `GET /api/me`, password changes, protected teacher/student pages, the local admin page, admin teacher management, admin lesson JSON management, teacher classroom/student/lesson/question/lesson-question/current-lesson/progress/submission APIs, and student current-lesson/question/submission read and create APIs.

The background scheduler and judge runner are wired in. The current student HTTP submission flow can create a submission, schedule it, judge it asynchronously, and write the result back.

The current API state is documented in [API_DESCRIBE.md](API_DESCRIBE.md) and [API_REF.md](API_REF.md).
