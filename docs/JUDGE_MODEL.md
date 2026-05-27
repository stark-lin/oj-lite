# Judge Model

`oj-lite` uses an in-process, Lua-only judging model for lightweight classroom exercises. It is designed for the local or LAN deployment boundary described in [PERMISSION.md](PERMISSION.md), not for hostile public online-judge workloads.

## Submission Lifecycle

Submissions persist their lifecycle in SQLite:

```text
pending -> judging -> finished
```

The scheduler runs a single polling loop. Before claiming new work, it marks stale `judging` submissions as `finished` with a `system_error`. It then claims up to `scheduler.fetch_batch_size` pending submissions in a transaction by updating them to `judging` and reading the associated source, reference code, and test cases in the same transaction.

Claimed submissions are dispatched through a semaphore-backed worker limit. At most `scheduler.concurrency` goroutines actively process judge work. If a claimed batch is larger than available slots, dispatch waits for slots; the scheduler does not build an unbounded in-memory channel or goroutine queue.

When processing completes, the worker updates a result only when its persisted status is still `judging`. SQLite is therefore the persisted coordination boundary for claim and finish transitions. The application also opens SQLite with one pooled connection, `foreign_keys=ON`, WAL journaling, and a configured busy timeout.

## Lua Execution

For each test case, student source and trusted `reference_code` are executed independently and their return values are compared. Every individual source execution creates a fresh `gopher-lua` `LState` and closes it afterwards, so student executions, reference executions, and separate test cases do not share Lua state.

The judge creates each state with `SkipOpenLibs: true` and registers only its controlled `print` function before loading source code. It does not open the default Lua standard libraries, so globals such as `os`, `io`, `package`, `debug`, `require`, `dofile`, `loadfile`, `math`, `string`, and `table` are not made available by the judge.

## Current Limits

The implementation currently applies these limits:

| Limit | Current value | Scope |
| --- | ---: | --- |
| Execution timeout | 2 seconds | Each individual student or reference Lua execution, not a total submission deadline |
| Call stack size | 128 | Each fresh `LState` |
| Registry capacity | 256 | Each fresh `LState`; automatic registry growth is disabled with `RegistryMaxSize: 0` |
| Execution stdout buffer | 8192 bytes | Each student or reference Lua execution |
| Persisted `submission.stdout_buffer` | 8192 bytes | Aggregated student stdout, enforced in worker logic and the database schema |
| Stale `judging` recovery window | 1 minute | Checked against `submitted_at` when the scheduler next claims work |

`RegistryGrowStep` is currently set to `32`, but it does not take effect while registry growth remains disabled.

The Lua execution limits are hard-coded in `internal/judge`; the stale recovery interval is hard-coded in `internal/scheduler`. Scheduler concurrency and claim batch size are configurable; their default values are both `4`.

## Security Boundary

This is a lightweight in-process restricted runtime:

- It reduces the student-visible Lua surface by omitting standard libraries and exposing only controlled output.
- It applies per-execution time, call stack, registry, and stdout constraints.
- It gives each execution an independent Lua state.

It is not an OS-level sandbox, a container boundary, a cgroup or seccomp policy, or a separate untrusted process. It does not claim process-level memory or kernel-level isolation from submitted code.

This trade-off is intended for a classroom deployment where students are untrusted, while teachers, local admin access, lesson content, and deployers are trusted. A public internet or adversarial contest deployment would require additional process-level isolation and operational controls.

## Known Limitations

- The two-second timeout is applied to each Lua execution. A submission containing many test cases can take longer overall because both student and reference source are run per case.
- Stale recovery currently measures age from `submitted_at`, not from a separate `judging_started_at` timestamp.
- Limits for the Lua runtime and stale recovery are not configurable through `config.json`.

## Implementation References

- [`internal/judge/sandbox.go`](../internal/judge/sandbox.go): Lua state options and hard-coded execution limits.
- [`internal/judge/runner.go`](../internal/judge/runner.go): per-execution state creation, timeout, and controlled `print`.
- [`internal/judge/service.go`](../internal/judge/service.go): student/reference execution per test case.
- [`internal/scheduler/runner.go`](../internal/scheduler/runner.go): polling loop and bounded worker dispatch.
- [`internal/scheduler/lease.go`](../internal/scheduler/lease.go): SQLite-backed claim and stale recovery.
- [`internal/scheduler/worker.go`](../internal/scheduler/worker.go): result writeback and aggregate stdout enforcement.
