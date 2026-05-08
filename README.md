# oj-lite

[English](#english) | [中文](#中文)

## English

`oj-lite` is a lightweight programming practice system for classroom teaching, built with Go, Gin, and SQLite.

It is not intended to be a general-purpose online judge. Instead, it provides a simple classroom workflow that works well on a local machine or LAN:

- A local admin manages teacher accounts and global lesson JSON.
- Teachers create classrooms and students.
- Teachers advance each classroom to the current lesson.
- Students only see the content for the current lesson.
- Students submit Lua code.
- The system judges submissions asynchronously and returns results.
- Teachers review classroom progress and submission history.

If you need a low-friction classroom practice tool instead of a public problem bank, contest platform, or full LMS, this project is closer to that use case.

### Project Scope

`oj-lite` focuses on a narrow teaching model:

- The local admin owns teacher account management and lesson/question authoring.
- Teachers control the classroom pace, and classrooms progress through a unified lesson order.
- Student permissions are intentionally small, limited to the current lesson and the student's own submissions.
- Submission and judging are decoupled, with submissions processed asynchronously in the background.
- Simple deployment and low maintenance cost are prioritized over a broad feature set.

The current judging model uses Lua functions. Student code is expected to expose:

```lua
function solution(...)
    -- student code
end
```

Each question stores `description`, `starter_code`, `reference_code`, and `test_cases`. During judging, the reference solution generates the expected result, then the student's return values are compared against it.

### Current Features

The repository currently provides a runnable classroom practice loop:

- Local admin page and APIs for teacher management.
- Local admin lesson JSON APIs for creating, replacing, listing, and deleting unreferenced lessons with their questions.
- Teacher and student login, logout, `GET /api/me`, and password changes.
- Teacher APIs for classrooms, students, lesson/question reads, lesson-question order, current lessons, progress, and submissions.
- Student APIs for current lesson, question reading, submission creation, submission lists, and submission details.
- Background scheduler and judge pipeline for asynchronous submissions.
- Embedded admin, teacher, student, and login pages.
- SQLite initialization plus an independent demo seed module.

Important boundaries:

- `/admin` and `/admin/*` are local-only loopback endpoints, not a production admin auth system.
- Teacher lesson/question write routes are not registered; content writes are handled by admin lesson JSON.
- Student access is scoped to the current classroom lesson.

External error semantics are intentionally small for business APIs:

- `401` means unauthenticated.
- `404` means the authenticated user cannot see, use, or perform the requested operation.
- `500` means an internal server error.

Public login and local admin validation errors may return `400`.

The current API state is documented in [docs/API_DESCRIBE.md](docs/API_DESCRIBE.md) and [docs/API_REF.md](docs/API_REF.md).

### Tech Stack

- Go 1.26
- Gin
- SQLite
- [gopher-lua](https://github.com/yuin/gopher-lua)

The application is a monolith. HTTP routing, authentication, judging, scheduling, database initialization, and embedded pages all run in the same process.

### Quick Start

Requirements:

- Go 1.26+

Start the service:

```sh
go run ./cmd/server
```

By default, the service listens on `0.0.0.0:8080` and stores data at `data/oj-lite.db`.

Start without demo seed data:

```sh
go run ./cmd/server --skip-seed
```

Available pages and probes:

- `GET /`: login page
- `GET /admin`: local admin page, loopback only
- `GET /teacher`: teacher page
- `GET /student`: student page
- `GET /healthz`: health check

Run tests:

```sh
go test ./...
```

Run the architecture dependency check:

```sh
go install -v github.com/arch-go/arch-go/v2@v2.1.2
arch-go --color no
```

The dependency rules are generated from [docs/DEPENDENCE.md](docs/DEPENDENCE.md)
and are enforced by [arch-go.yml](arch-go.yml).

Demo accounts are initialized on first startup unless `--skip-seed` is used:

- `teacher / teacher`
- `student / student`

The seed data also creates sample classrooms and the embedded 24-lesson course under `internal/seed/lessons/` so the full workflow can be tried immediately. Use `--skip-seed` when you want to start from an empty database schema.

Common environment variables:

- `APP_NAME`: default `oj-lite`
- `APP_ENV`: default `local`
- `HTTP_HOST`: default `0.0.0.0`
- `HTTP_PORT`: default `8080`
- `GIN_MODE`: default `debug`
- `HTTP_READ_TIMEOUT`: default `5s`
- `HTTP_WRITE_TIMEOUT`: default `10s`
- `HTTP_IDLE_TIMEOUT`: default `60s`
- `HTTP_SHUTDOWN_TIMEOUT`: default `10s`
- `DB_PATH`: default `data/oj-lite.db`
- `DB_BUSY_TIMEOUT`: default `5s`
- `SCHEDULER_CONCURRENCY`: default `4`
- `SCHEDULER_FETCH_BATCH_SIZE`: defaults to `SCHEDULER_CONCURRENCY`
- `SCHEDULER_IDLE_SLEEP`: default `1s`

Notes:

- Database initialization currently supports creating the final schema only from an empty database.
- If an existing local `data/oj-lite.db` uses an incompatible old schema, delete it and start the service again.

### Use Cases And Non-goals

Good fit:

- Teachers running unified programming exercises in a class or lab.
- LAN or single-machine deployments that need a usable practice system quickly.
- Small teaching scenarios that need minimal operational complexity.

Not a good fit:

- Public problem bank platforms.
- ACM / OI style contest systems.
- General multi-language online judging.
- Multi-tenant organization collaboration.
- Full production security for public internet exposure.
- A complete LMS or production-grade admin console.

### Repository Layout

```text
cmd/server/           Program entry point
internal/app/         Application wiring, routes, embedded pages, and assets
internal/admin/       Local admin teacher and lesson JSON management
internal/classroom/   Classrooms, students, and current lessons
internal/lesson/      Teacher-facing lesson and lesson-question reads
internal/question/    Teacher question reads and student question reads
internal/submission/  Student submission APIs
internal/progress/    Teacher progress and submission queries
internal/judge/       Lua judging logic
internal/scheduler/   Background scheduler and workers
internal/seed/        Demo data seeding
internal/platform/    Auth, user, config, database, session, middleware, and shared infrastructure
docs/           Product, API, permission, ER, and other supplemental docs
```

### Documentation
- [docs/PRD.md](docs/PRD.md): product goals, scope, and non-goals
- [docs/PERMISSION.md](docs/PERMISSION.md): permission boundaries and key validation rules
- [docs/API_DESCRIBE.md](docs/API_DESCRIBE.md): route overview and current implementation status
- [docs/API_REF.md](docs/API_REF.md): API contract
- [docs/ER.md](docs/ER.md): core data model
- [docs/DEPENDENCE.md](docs/DEPENDENCE.md): package dependency graph and architecture check workflow
- [docs/FILES.md](docs/FILES.md): repository layout summary

### Roadmap

Near-term work:

- Add a real admin authentication flow if the admin surface needs to move beyond loopback-only use.
- Continue improving the teacher and student pages.
- Improve error handling, observability, and test coverage.
- Improve judging isolation and stability while keeping the system lightweight.

### License

`oj-lite` is licensed under the GNU General Public License v3.0 or later.

See [COPYING](COPYING) for the full license text.

SPDX-License-Identifier: GPL-3.0-or-later

## 中文

`oj-lite` 是一个面向课堂教学场景的轻量编程训练系统，使用 Go、Gin 和 SQLite 构建。

它的目标不是做一个通用在线判题平台，而是提供一套适合本地或局域网部署的课堂练习闭环：

- 本机 admin 管理 teacher 账号和全局 lesson JSON。
- 教师创建班级和学生。
- 教师推进班级当前 lesson。
- 学生只看到当前 lesson 的内容。
- 学生提交 Lua 代码。
- 系统异步判题并返回结果。
- 教师查看班级进度和提交情况。

如果你需要的是课堂内低摩擦练习工具，而不是开放题库、竞赛平台或完整 LMS，这个项目更接近实际需求。

### 项目定位

`oj-lite` 聚焦一个非常明确的教学模型：

- 本机 admin 负责 teacher 账号管理和 lesson/question 内容维护。
- 教师主导课堂节奏，班级围绕统一的 lesson 顺序推进。
- 学生权限尽量收敛，只访问当前 lesson 与自己的提交。
- 判题流程与提交流程解耦，提交进入后台异步执行。
- 优先保证部署简单、维护成本低，而不是功能面面俱到。

当前系统采用 Lua 函数式判题模型，学生代码入口约定为：

```lua
function solution(...)
    -- student code
end
```

题目保存 `description`、`starter_code`、`reference_code` 和 `test_cases`，判题时会先运行参考实现生成期望结果，再比较学生代码返回值。

### 当前能力

当前仓库已经具备一套可运行的课堂训练闭环：

- 本地 admin 页面与 teacher 管理接口。
- 本地 admin lesson JSON 接口，可创建、替换、查看、删除未被引用的 lesson 及其 questions。
- Teacher / Student 登录、登出、`/api/me`、修改密码。
- 教师侧班级、学生、lesson/question 读取、lesson-question、current lesson、progress、submissions 接口。
- 学生侧当前 lesson、题目读取、提交列表、提交创建与提交详情接口。
- 后台 scheduler + judge 异步判题链路。
- 内嵌 admin、teacher、student 和 login 页面。
- SQLite 初始化，以及独立的 demo seed 模块。

重要边界：

- `/admin` 和 `/admin/*` 只允许 loopback 本机访问，不是生产级 admin 鉴权系统。
- teacher 侧不注册 lesson/question 写接口；内容写入由 admin lesson JSON 负责。
- student 访问被收敛到当前 classroom lesson。

业务 API 对外错误语义已经收口：

- `401` 只用于未认证。
- `404` 用于已认证后的一切“不可见 / 不可用 / 不给做”。
- `500` 用于内部异常。

公共登录和本机 admin 字段校验可能返回 `400`。

接口现状以 [docs/API_DESCRIBE.md](docs/API_DESCRIBE.md) 和 [docs/API_REF.md](docs/API_REF.md) 为准。

### 技术栈

- Go 1.26
- Gin
- SQLite
- [gopher-lua](https://github.com/yuin/gopher-lua)

整体形态是一个单体服务，HTTP、鉴权、判题调度、数据库初始化和页面都在同一个进程内完成。

### 快速开始

运行要求：

- Go 1.26+

启动服务：

```sh
go run ./cmd/server
```

默认监听 `0.0.0.0:8080`，默认数据库路径为 `data/oj-lite.db`。

启动后可以访问：

- `GET /`：登录页
- `GET /admin`：本机 admin 页面，仅 loopback 可访问
- `GET /teacher`：教师页
- `GET /student`：学生页
- `GET /healthz`：健康检查

运行测试：

```sh
go test ./...
```

首次启动会自动初始化一组演示账号：

- `teacher / teacher`
- `student / student`

同时会写入示例班级，以及 `internal/seed/lessons/` 中内置的 24 节课程，方便直接体验完整流程。

常用环境变量：

- `APP_NAME`：默认 `oj-lite`
- `APP_ENV`：默认 `local`
- `HTTP_HOST`：默认 `0.0.0.0`
- `HTTP_PORT`：默认 `8080`
- `GIN_MODE`：默认 `debug`
- `HTTP_READ_TIMEOUT`：默认 `5s`
- `HTTP_WRITE_TIMEOUT`：默认 `10s`
- `HTTP_IDLE_TIMEOUT`：默认 `60s`
- `HTTP_SHUTDOWN_TIMEOUT`：默认 `10s`
- `DB_PATH`：默认 `data/oj-lite.db`
- `DB_BUSY_TIMEOUT`：默认 `5s`
- `SCHEDULER_CONCURRENCY`：默认 `4`
- `SCHEDULER_FETCH_BATCH_SIZE`：默认与 `SCHEDULER_CONCURRENCY` 相同
- `SCHEDULER_IDLE_SLEEP`：默认 `1s`

说明：

- 当前数据库初始化只支持空库直接创建最终 schema，不支持旧 schema 的增量迁移。
- 如果本地已有不兼容的旧版 `data/oj-lite.db`，需要删除后重新启动。

### 适用场景与非目标

适合：

- 教师在课堂或实验课中统一推进编程练习。
- 在局域网或单机环境下快速部署一个可用练习系统。
- 需要最小化运维复杂度的小规模教学场景。

不适合：

- 开放式刷题平台。
- ACM / OI 风格竞赛系统。
- 多语言通用 OJ。
- 多租户、组织级协作平台。
- 面向公网的完整生产安全方案。
- 完整 LMS 或生产级 admin console。

### 仓库结构

```text
cmd/server/           程序入口
internal/app/         应用装配、路由、内嵌页面与资源
internal/admin/       本地 admin teacher 与 lesson JSON 管理
internal/classroom/   班级、学生与 current lesson
internal/lesson/      teacher 侧 lesson 与 lesson-question 只读接口
internal/question/    teacher 侧 question 读取与 student 读题接口
internal/submission/  学生提交接口
internal/progress/    教师进度与提交查询
internal/judge/       Lua 判题逻辑
internal/scheduler/   后台调度与 worker
internal/seed/        demo 数据初始化
internal/platform/    auth、user、配置、数据库、session、中间件等基础设施
docs/           产品、接口、权限、ER 等补充文档
```

### 文档

- [docs/PRD.md](docs/PRD.md)：产品目标、范围与非目标
- [docs/PERMISSION.md](docs/PERMISSION.md)：权限边界与关键校验规则
- [docs/API_DESCRIBE.md](docs/API_DESCRIBE.md)：路由总览与当前实现状态
- [docs/API_REF.md](docs/API_REF.md)：接口契约
- [docs/ER.md](docs/ER.md)：核心数据模型
- [docs/FILES.md](docs/FILES.md)：仓库结构摘要

### Roadmap

短期内比较明确的后续工作包括：

- 如果 admin 面需要超出本机 loopback 使用，补充真实 admin 鉴权流程。
- 继续完善教师端与学生端页面。
- 增强错误处理、可观测性和测试覆盖。
- 在保持轻量前提下优化判题执行隔离与稳定性。

### License

`oj-lite` 使用 GNU General Public License v3.0 or later 授权。

完整许可证文本见 [COPYING](COPYING)。

SPDX-License-Identifier: GPL-3.0-or-later
