CREATE TABLE IF NOT EXISTS user_account (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    CHECK (length(username) BETWEEN 3 AND 32),
    CHECK (username NOT GLOB '*[^A-Za-z0-9_]*'),
    CHECK (length(password_hash) > 0),
    CHECK (role IN ('teacher', 'student')),
    CHECK (status IN ('active', 'disabled'))
);

CREATE TABLE IF NOT EXISTS classroom (
    id INTEGER PRIMARY KEY,
    teacher_id INTEGER NOT NULL REFERENCES user_account(id) ON DELETE RESTRICT,
    current_lesson_id INTEGER REFERENCES lesson(id) ON DELETE RESTRICT,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    CHECK (length(trim(name)) > 0)
);

CREATE TABLE IF NOT EXISTS lesson (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    sort_order INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    CHECK (length(trim(title)) > 0),
    CHECK (sort_order > 0)
);

CREATE TABLE IF NOT EXISTS question (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    starter_code TEXT NOT NULL,
    reference_code TEXT NOT NULL,
    test_cases TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    CHECK (length(trim(title)) > 0),
    CHECK (json_valid(test_cases))
);

CREATE TABLE IF NOT EXISTS enrollment (
    id INTEGER PRIMARY KEY,
    classroom_id INTEGER NOT NULL REFERENCES classroom(id) ON DELETE RESTRICT,
    student_id INTEGER NOT NULL REFERENCES user_account(id) ON DELETE RESTRICT,
    current_lesson_id INTEGER REFERENCES lesson(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE TABLE IF NOT EXISTS lesson_question (
    id INTEGER PRIMARY KEY,
    lesson_id INTEGER NOT NULL REFERENCES lesson(id) ON DELETE RESTRICT,
    question_id INTEGER NOT NULL REFERENCES question(id) ON DELETE RESTRICT,
    sort_order INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    CHECK (sort_order > 0)
);

CREATE TABLE IF NOT EXISTS submission (
    id INTEGER PRIMARY KEY,
    enrollment_id INTEGER NOT NULL REFERENCES enrollment(id) ON DELETE RESTRICT,
    lesson_id INTEGER NOT NULL REFERENCES lesson(id) ON DELETE RESTRICT,
    lesson_question_id INTEGER NOT NULL REFERENCES lesson_question(id) ON DELETE RESTRICT,
    status TEXT NOT NULL,
    verdict TEXT,
    source_code TEXT NOT NULL,
    stdout_buffer TEXT,
    error_message TEXT,
    judge_report TEXT,
    submitted_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    finished_at TEXT,
    CHECK (status IN ('pending', 'judging', 'finished')),
    CHECK (verdict IS NULL OR verdict IN ('accepted', 'wrong_answer', 'runtime_error', 'system_error')),
    CHECK (length(CAST(source_code AS BLOB)) BETWEEN 1 AND 65535),
    CHECK (stdout_buffer IS NULL OR length(CAST(stdout_buffer AS BLOB)) <= 8192),
    CHECK (judge_report IS NULL OR json_valid(judge_report)),
    CHECK ((status = 'finished' AND finished_at IS NOT NULL) OR (status IN ('pending', 'judging') AND finished_at IS NULL)),
    CHECK (status = 'finished' OR verdict IS NULL),
    CHECK (finished_at IS NULL OR finished_at >= submitted_at)
);
