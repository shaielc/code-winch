# P1-061: Harden integration-suite clocks and process-state checks

**Phase:** 1 — Local single-user vertical slice
**Shape:** hardening
**Dependencies:** P1-012 (semantic: the outbox claim this constrains does not exist until the outbox does), P1-049 (semantic: it reworks the local sandbox driver and its tests, so hardening them first would be rewritten)

## Objective

An unreaped child process fails the suite under its own name rather than as
"survived cleanup", and no integration test can pass on the day it is written
and fail on a later date.

## Scope

- A process-state helper that distinguishes **live**, **zombie**, and **absent**
  by reading `/proc/<pid>/stat`, instead of `kill(pid, 0)` — which succeeds
  against a zombie and so reports a dead process as alive.
- `TestCleanupKillsChildAndToleratesRetries` keeps failing on a zombie. A
  zombie is a real defect: something in the process tree is not reaping, and
  the run leaks a process-table entry per occurrence. The change is the
  diagnosis, not the verdict — the failure names the state so the next reader
  is not sent to look for a bug in `Cleanup`, which is correct.
- Poll to the existing deadline before classifying, so the ordinary window
  between kill and reap is not reported as a leak.
- State and apply the clock rule for integration tests: **a claim clock
  compared against a timestamp the database supplied must derive from wall
  time.** A fixed date is safe only where the test itself writes the timestamp
  being compared.
- Audit the remaining fixed dates in the integration tests against that rule and
  record the result where the next author will see it.

## Non-goals

- Changing `Cleanup`'s kill behavior. It terminates the process group correctly;
  the failure this task addresses was in how the test observes the result.
- Making the suite pass without an init process. Reaping is the container's job;
  this task makes its absence legible, not tolerable.
- A general-purpose clock abstraction for tests. The rule is one sentence and
  two call sites, not a framework.

## Runtime reachability

`make test-env && make test-integration`, per `deployments/README.md` — the
compose `test` profile is the documented and only supported way to run this
suite, since the host carries no Go toolchain.

## Owned surfaces

`test/procstate/`, `internal/adapters/sandbox/local/local_test.go`,
`internal/adapters/postgres/repository_integration_test.go`,
`internal/adapters/postgres/workflow_repository_integration_test.go`.

`deployments/compose.yml` and `deployments/README.md` are P1-048's surfaces; the
container's reaping requirement is documented there by that task, not this one.
This is quality work on code other tasks are still moving, so it is deliberately
sequenced behind P1-049 rather than written ahead of it.

## Demonstration

    $ make test-integration
    → expect: every package ok, including internal/adapters/sandbox/local

    # Remove `init: true` from the runner service, then:
    $ docker compose -f deployments/compose.yml up -d --force-recreate runner
    $ docker compose -f deployments/compose.yml exec -T runner \
        go test -tags integration -count=1 ./internal/adapters/sandbox/local/
    → expect: FAIL naming the child as an unreaped zombie and pointing at the
      missing init — not "survived cleanup", which reads as a driver bug

    $ docker compose -f deployments/compose.yml exec -T runner ps -eo pid,ppid,stat,comm
    → expect: with init restored, no `Z` entries accumulate across repeated runs

## Verification

- Full integration suite green in the runner: `make test-integration`.
- Focused tests for the classifier: a running process reports live; a killed but
  unreaped child reports zombie; an unused PID reports absent.
- `gofmt -l` and `go vet -tags integration` clean on the touched packages, run
  in the runner container.

## Acceptance criteria

- [ ] A zombie child fails the cleanup test, and the message names the state and
      the likely cause rather than asserting the process is still running.
- [ ] The classifier distinguishes all three states and is exported for reuse.
- [ ] No integration test compares a database-supplied timestamp against a
      hardcoded date.
- [ ] The clock rule is written down at the call site it governs.
- [ ] Repeated runs leave no growing set of zombies in the runner container.

## Deferrals

| Deferred | Owning task |
|---|---|
| Whether the daemon container leaks descendants when a harness is killed, against the real deployment | P1-053 |

## Traces to

`docs/workplan/README.md` "The standing scenario suite"; P1-053 scenario 4
("a harness that ignores `SIGTERM` is escalated and leaves no descendant
process"), which asserts the same property end to end and can reuse the
classifier this task exports.
