# Repository Structure

This document keeps a high-level directory summary and does not explain every file individually.

## Top-level Structure

```text
oj-lite/
├─ README.md
├─ arch-go.yml
├─ go.mod
├─ go.sum
├─ binary_res/
├─ cmd/
└─ internal/
```

After startup, the service creates `data/oj-lite.db` and SQLite WAL/SHM files by default. That directory is runtime data, not source structure.

## Directories

### `cmd/server/`

Program entry point. It starts the HTTP service and background scheduler, registers signal handling, and shuts down gracefully.

### `internal/app/`

Application wiring layer:

- Loads configuration.
- Initializes the database.
- Registers page, admin, auth, teacher, and student routes.
- Provides the embedded login page, local admin page, teacher page, and student page.
- Provides embedded `/assets/app.css` and `/assets/app.js`.

### `internal/platform/`

Infrastructure layer:

- `auth/`: business login, logout, `GET /api/me`, password changes, and current-user context helpers.
- `config/`: environment-variable configuration.
- `db/`: SQLite connection, migration, and embedded SQL.
- `httpx/`: shared response and error output helpers.
- `middleware/`: request id, logging, auth, and recover middleware.
- `password/`: password hashing and validation.
- `session/`: business session cookie handling.
- `user/`: user account data access and service wrappers.
- `logger/`, `clock/`, `ids/`: shared helper components.

### Business Modules

- `internal/admin/`: local admin teacher management and lesson JSON management.
- `internal/classroom/`: classroom, student, and current-lesson APIs.
- `internal/lesson/`: teacher-side read-only lesson and lesson-question APIs.
- `internal/question/`: teacher-side read-only question APIs and student question reads.
- `internal/submission/`: student submission creation, list, and detail APIs.
- `internal/progress/`: teacher progress and submission queries.
- `internal/judge/`: Lua execution engine, test-case runner, result comparison, and structured reports.
- `internal/scheduler/`: background pending-submission claiming, concurrency-limited execution, and result writeback.
- `internal/seed/`: idempotent demo account, classroom, lesson, and question seeding.

### `binary_res/`

Design and reference documentation:

- `PRD.md`: product scope.
- `PERMISSION.md`: permission rules.
- `API_DESCRIBE.md`: route overview.
- `API_REF.md`: API contract.
- `ER.md`: data model.
- `DEPENDENCE.md`: package dependency graph and architecture check workflow.
- `FILES.md`: repository structure.

## Suggested Reading Order

For a quick implementation overview, read in this order:

1. `cmd/server/main.go`
2. `internal/app/`
3. `internal/platform/`
4. `internal/admin/`
5. `internal/platform/auth/`
6. `internal/classroom/`
7. `internal/submission/`
8. `internal/seed/`
9. `internal/scheduler/`
10. `internal/judge/`
11. The remaining teacher / student business modules.
