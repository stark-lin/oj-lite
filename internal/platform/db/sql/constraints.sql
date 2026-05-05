CREATE UNIQUE INDEX IF NOT EXISTS uk_user_account_username
    ON user_account (username);

CREATE INDEX IF NOT EXISTS idx_classroom_teacher_id
    ON classroom (teacher_id);

CREATE INDEX IF NOT EXISTS idx_classroom_current_lesson_id
    ON classroom (current_lesson_id);

CREATE UNIQUE INDEX IF NOT EXISTS uk_lesson_sort_order
    ON lesson (sort_order);

CREATE UNIQUE INDEX IF NOT EXISTS uk_enrollment_classroom_student
    ON enrollment (classroom_id, student_id);

CREATE UNIQUE INDEX IF NOT EXISTS uk_enrollment_student_id
    ON enrollment (student_id);

CREATE INDEX IF NOT EXISTS idx_enrollment_student_id
    ON enrollment (student_id);

CREATE INDEX IF NOT EXISTS idx_enrollment_current_lesson_id
    ON enrollment (current_lesson_id);

CREATE UNIQUE INDEX IF NOT EXISTS uk_lesson_question_lesson_question
    ON lesson_question (lesson_id, question_id);

CREATE UNIQUE INDEX IF NOT EXISTS uk_lesson_question_lesson_sort_order
    ON lesson_question (lesson_id, sort_order);

CREATE INDEX IF NOT EXISTS idx_lesson_question_question_id
    ON lesson_question (question_id);

CREATE INDEX IF NOT EXISTS idx_submission_status_submitted_at
    ON submission (status, submitted_at DESC);

CREATE INDEX IF NOT EXISTS idx_submission_enrollment_submitted_at
    ON submission (enrollment_id, submitted_at DESC);

CREATE INDEX IF NOT EXISTS idx_submission_lesson_submitted_at
    ON submission (lesson_id, submitted_at DESC);

CREATE INDEX IF NOT EXISTS idx_submission_lesson_question_submitted_at
    ON submission (lesson_question_id, submitted_at DESC);
