# Permission Model

This document keeps the permission boundaries and validation rules that the implementation must preserve.

## Trust Boundaries

Trusted side:

- Local maintainer.
- Deployer.
- Loopback-only local admin entry point.
- Teacher.

Untrusted side:

- Student.

The design focus is preventing student privilege escalation. It does not attempt to defend against malicious teachers, local maintainers, or deployers.

## Role Boundaries

### Admin

- Performs system management only.
- Does not participate in classroom activity.
- Is not stored in the business `user_account` table.
- Is currently restricted by loopback IP and has no separate admin session.

The current admin can:

- Create teachers.
- Update teacher username or status.
- Reset teacher passwords.
- Delete teachers without classrooms, or disable teachers that already own classrooms.
- Create, replace, and delete unreferenced global lessons.
- Maintain lesson questions through lesson JSON.

### Teacher

- Can only manage their own classrooms.
- Can create students and add them to their own classrooms.
- Can read global lessons, questions, and lesson-question records.
- Can set the current lesson for their own classrooms.
- Can view progress and submissions for their own classrooms.
- Can delete submissions in their own classrooms.

Teachers currently do not write lessons or questions. Content writes are centralized in admin lesson JSON.

### Student

- Can only see their own current lesson.
- Can only read questions in the current lesson.
- Can only submit their own code.
- Can only view their own submissions.
- Cannot access teacher or admin routes.

## Core Rules

### 1. Student Access Must Be Scoped To The Current Lesson

Student question, submission, and result access must be derived from the session classroom and the current lesson stored in the database. The backend must not trust ownership information supplied by the frontend.

### 2. Teacher Access Must Check Object Ownership

Every classroom-related teacher operation must satisfy:

```text
classroom.teacher_id == current_user.id
```

### 3. Submission Ownership Must Be Derived Server-side

When a student creates a submission, the frontend only sends:

- `lesson_question_id`
- `source_code`

The following fields are not trusted inputs:

- `student_id`
- `classroom_id`
- `lesson_id`
- `enrollment_id`

### 4. Content Writes Are Centralized In Local Admin

Lesson and question creation, replacement, and deletion only happen through `/admin/lessons`. The teacher side currently reads global lessons and questions but does not write them.

## Object-level Validation

### Teacher

- Classroom access: `classroom.teacher_id == current_user.id`.
- Enrollment / student access: the target enrollment must belong to the current teacher's classroom.
- Rename student: the student must belong to the current teacher's classroom.
- Reset student password: the student must belong to the current teacher's classroom.
- Remove student: the student must belong to the current teacher's classroom.
- View progress: the classroom must belong to the current teacher.
- View / delete submission: the submission must belong to the current teacher's classroom.
- Access lesson / question: lessons and questions are global resources and are not isolated by teacher.

### Student

- Enrollment access: `enrollment.student_id == current_user.id`.
- The current implementation uses `uk_enrollment_student_id` so each student can belong to only one classroom.
- Current lesson reads are derived from the session classroom and the current student's enrollment.
- Question reads require `lesson_question` to belong to the current classroom's `current_lesson_id`.
- Submission creation requires `lesson_question` to belong to the current classroom's `current_lesson_id`.
- Submission reads must resolve through enrollment and belong to the current user.

### Admin

- Every `/admin/*` request must come from a loopback IP.
- Admin teacher operations only affect `user_account` rows where `role = 'teacher'`.
- When replacing a lesson, any question with an `id` must already belong to the target lesson.
- Deleting a lesson that is referenced by a classroom, enrollment, or submission is rejected by database foreign keys.

## Database-level Constraints Implemented Today

- `user_account.username` is unique.
- `user_account.username` length is `3..32` and only allows `A-Za-z0-9_`.
- `user_account.role` only allows `teacher` / `student`.
- `user_account.status` only allows `active` / `disabled`.
- `enrollment(classroom_id, student_id)` is unique.
- `enrollment(student_id)` is unique.
- `lesson(sort_order)` is globally unique.
- `lesson_question(lesson_id, question_id)` is unique.
- `lesson_question(lesson_id, sort_order)` is unique.
- `question.test_cases` must be valid JSON.
- `submission.status` only allows `pending` / `judging` / `finished`.
- `submission.verdict` only allows `accepted` / `wrong_answer` / `runtime_error` / `system_error`.
- `submission.source_code` length must be `1..65535` bytes.
- `submission.stdout_buffer` is limited to `8192` bytes.
- `submission.judge_report` must be valid JSON.
- `enrollment.current_lesson_id` must match `classroom.current_lesson_id`.
- `submission.lesson_question_id` must belong to `submission.lesson_id`.
- `submission.lesson_id` must match the classroom current lesson.
- After a question has any submission, `description`, `starter_code`, `reference_code`, and `test_cases` cannot be changed.
- After a submission is inserted, `enrollment_id`, `lesson_id`, `lesson_question_id`, and `source_code` cannot be changed.

## Content Freeze Rules

- Once a question has produced a submission, its core judging fields should not be modified directly.
- Submission ownership fields should be treated as immutable.
- Admin lesson replacement may try to update question core fields; if the target question already has submissions, the database trigger rejects the update.

## Current Implementation Notes

The codebase currently implements these permission controls:

- Business Cookie Session middleware.
- Role isolation for `/api/teacher/*` and `/api/student/*`.
- `/teacher` page access only for teachers.
- `/student` page access only for students.
- `/admin` page and `/admin/*` APIs only for loopback requests.
- Teacher object-ownership checks for classroom, student, progress, and submission operations.
- Student current-lesson, question, and submission checks against current lesson and enrollment ownership.
- SQLite foreign keys, unique indexes, and triggers for lower-level consistency.
