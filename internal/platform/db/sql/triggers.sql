CREATE TRIGGER IF NOT EXISTS trg_enrollment_current_lesson_insert
BEFORE INSERT ON enrollment
FOR EACH ROW
WHEN COALESCE(NEW.current_lesson_id, 0) != COALESCE((
    SELECT current_lesson_id
    FROM classroom
    WHERE id = NEW.classroom_id
), 0)
BEGIN
    SELECT RAISE(ABORT, 'enrollment.current_lesson_id must match classroom.current_lesson_id');
END;

CREATE TRIGGER IF NOT EXISTS trg_enrollment_current_lesson_update
BEFORE UPDATE OF classroom_id, current_lesson_id ON enrollment
FOR EACH ROW
WHEN COALESCE(NEW.current_lesson_id, 0) != COALESCE((
    SELECT current_lesson_id
    FROM classroom
    WHERE id = NEW.classroom_id
), 0)
BEGIN
    SELECT RAISE(ABORT, 'enrollment.current_lesson_id must match classroom.current_lesson_id');
END;

CREATE TRIGGER IF NOT EXISTS trg_submission_consistency_insert
BEFORE INSERT ON submission
FOR EACH ROW
BEGIN
    SELECT RAISE(ABORT, 'submission.lesson_question_id must belong to submission.lesson_id')
    WHERE NOT EXISTS (
        SELECT 1
        FROM lesson_question lq
        WHERE lq.id = NEW.lesson_question_id
          AND NEW.lesson_id = lq.lesson_id
    );

    SELECT RAISE(ABORT, 'submission.lesson_id must match classroom.current_lesson_id')
    WHERE NOT EXISTS (
        SELECT 1
        FROM enrollment e
        JOIN classroom c
          ON c.id = e.classroom_id
        WHERE e.id = NEW.enrollment_id
          AND c.current_lesson_id = NEW.lesson_id
    );
END;

CREATE TRIGGER IF NOT EXISTS trg_question_freeze_after_submission
BEFORE UPDATE OF description, starter_code, reference_code, test_cases ON question
FOR EACH ROW
WHEN EXISTS (
    SELECT 1
    FROM lesson_question lq
    JOIN submission s
      ON s.lesson_question_id = lq.id
    WHERE lq.question_id = OLD.id
    LIMIT 1
)
BEGIN
    SELECT RAISE(ABORT, 'question core fields are frozen after the first submission')
    WHERE NEW.description IS NOT OLD.description
       OR NEW.starter_code IS NOT OLD.starter_code
       OR NEW.reference_code IS NOT OLD.reference_code
       OR NEW.test_cases IS NOT OLD.test_cases;
END;

CREATE TRIGGER IF NOT EXISTS trg_submission_core_fields_immutable
BEFORE UPDATE OF enrollment_id, lesson_id, lesson_question_id, source_code ON submission
FOR EACH ROW
WHEN NEW.enrollment_id IS NOT OLD.enrollment_id
   OR NEW.lesson_id IS NOT OLD.lesson_id
   OR NEW.lesson_question_id IS NOT OLD.lesson_question_id
   OR NEW.source_code IS NOT OLD.source_code
BEGIN
    SELECT RAISE(ABORT, 'submission core fields are immutable after insert');
END;
