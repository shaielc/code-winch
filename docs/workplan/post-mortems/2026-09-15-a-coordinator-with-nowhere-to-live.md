# 2026-09-15 — A coordinator with nowhere to live

**Found:** auditing P0-008 against pull request #53 at `b2cf802`, before merge.
**Symptom:** three rounds over two days moved the same ~180 lines of run
orchestration through three different packages, added the honest API problem
code for a refusal the brief requires and then reverted it, and landed on a
placement that inverts the dependency rule in `docs/code-structure.md` §2.
Every round is a faithful reading of the brief. The brief cannot be satisfied.

## What the rounds did

| Commit | Date | Where the coordinator went | Also |
|---|---|---|---|
| `f558bdb` | Sep 11 | inline in `cmd/winchd/main.go` (+182) | profile check and a raw state assignment inside `RunService.Start` |
| `d8fba75` | Sep 13 | `internal/supervisor/coordinator.go` (+159), `main.go` −126 | added `ApplyRunTransition` and the `RunRuntime.Validate` port; added `ErrUnsupportedProfile` → 400 `unsupported_run_profile`; aliased `runner/local.Observation` to the application type |
| `b2cf802` | Sep 13, 48 min later | `internal/application/runexec/coordinator.go` | reverted both edits outside the write set; added bounded shutdown |

Round 1 also wrote the queued transition by hand —
`record.Attempts[len(record.Attempts)-1].State = domain.RunStateQueued` — and
put `record.HarnessProfile != "fake" || record.SandboxProfile != "local"`
inside the provider-neutral use case. Round 2 corrected both: the domain state
machine got `ApplyRunTransition`, and profile selection moved behind
`RunRuntime.Validate`. That part of the back and forth is ordinary review
distance. The placement churn is not.

## Root cause

Two declaration defects in one brief. No other brief is involved, and the
brief's Scope and Acceptance criteria are each individually sound.

### A — the Scope requires an orchestrator the Write set gives no home

Scope: "Construct `supervisor.Supervisor` and `runner/local.Runner`, wiring
runner observations to `Supervisor.Observe` and durable `EventStore` append."
That is a component, not a registration — a lease, two fenced runner commands,
an observation pump, four durable transitions, and a shutdown path.

Write set: `cmd/winchd/main.go` **(registry registration only — no wholesale
rewrite)**, `internal/application/` (start use case), `cmd/winch/`,
`test/e2e/start_test.go`, and tests. `internal/supervisor/` is absent.

So each round obeyed one declaration by breaking another:

- Round 1 obeyed "construct … in `cmd/winchd`" and broke "registration only":
  a 182-line composition root.
- Round 2 obeyed "registration only" and wrote 159 lines into
  `internal/supervisor/`, which the write set does not name.
- Round 3 obeyed the write set by filing orchestration under
  `internal/application/` — and that is where the dependency rule broke.
  `internal/application/runexec` imports `internal/supervisor` (`coordinator.go:14`),
  the only such import outside tests, while `docs/code-structure.md` §2 has
  supervisor and the adapters depending *on* the application layer:

  ```sh
  $ grep -rn "code-winch/internal/supervisor" --include=*.go . | grep -v _test.go
  ./internal/application/runexec/coordinator.go:14
  ```

  The same package hardcodes `"fake"`, `"local"`, and a `fakeProfileRedactor`,
  so profile-specific policy now lives in the application tree that AGENTS.md
  asks to keep provider-neutral.

There was no fourth option. The one package where an orchestrator that consumes
both the application layer and the supervisor belongs — beside them, or in the
composition root the brief had just restricted — was unavailable by
declaration.

### B — the Scope requires a refusal the Contract surfaces cannot express

Scope: "refuse anything else with a stable, content-free error". Acceptance
criterion 3 repeats it: "refused with a stable error code, and no harness
process is launched."

Contract surfaces: the two API operations and the driver namespace. No problem
code. And the write set excludes `internal/adapters/transport/httpapi/`, which
owns the code-to-status mapping (`server.go:291-306`), while
`api/openapi/code-winch.yaml:80-87` declares `startRun`'s responses as 202,
400, 401, 404, 409, 412, 428 — none of which means "this profile is not
supported".

Round 2 wrote the honest answer and had to leave the write set to do it:

```go
// d8fba75, internal/adapters/transport/httpapi/server.go
+	ErrUnsupportedProfile  = errors.New("unsupported run profile")
+	case errors.Is(err, ErrUnsupportedProfile):
+		s.problem(w, r, 400, "unsupported_run_profile", "Unsupported run profile", …)
```

Round 3 reverted it and mapped the refusal onto the nearest declared code, so
the shipped behavior is:

```
$ curl -X POST …/runs/$ID/start -H 'If-Match: "1"'   # harnessProfile=claude
409 {"code":"run_state_conflict","detail":"The command is invalid for the run's current state."}
```

The run's state is `created`, which is exactly the state `start` is valid from.
The code is stable, so the criterion passes; the sentence is false.

## Why no gate caught it

- `make check` passes in all three shapes. Write-set and contract-surface
  conformance are prose, checked by review, and nothing in CI reads them.
- `api-validate` and `api-compat` check the document against itself and against
  the v1 baseline. Neither can notice that a refusal the brief mandates has no
  code in the document to carry it.
- The acceptance criteria are behavioral and were met by every round from
  round 2 onward. None of them constrains where the code lives, so the audit
  that judges them cannot see the churn that produced them.
- The brief's Demonstration is the happy path plus a process count. The
  placement question and the refusal's wording are both invisible to it.

## Consequences if left as is

- Orchestration sits in `internal/application/`, importing outward. P0-009
  (input), P0-011 (stop), and P0-014 (memory profile) all extend that package
  and inherit the inversion, each making it more expensive to undo.
- `unsupported_run_profile` remains unexpressible, so an operator debugging a
  refused start reads a sentence about run state that is not true.
- The next seam brief in this phase is written to the same template, with the
  same "registration only" restriction on `cmd/winchd/main.go` and the same
  absence of a package for what that restriction displaces.

## Remediation

P0-008's own defect — a launch failure left the run in a nonterminal state
behind a held lease — is fixed on its branch. The declaration defects are not
P0-008's to absorb: they are why the code looks the way it does.

1. **P0-019** (hardening, `revision` edge to P0-008) owns the coordinator's
   home, the dependency direction, the refusal's problem code, and the
   nonterminal control record. Its brief names `internal/supervisor/` and
   `internal/adapters/transport/httpapi/` in its write set, and the problem
   code and the `startRun` response set among its contract surfaces — the two
   declarations P0-008 lacked.
2. **P0-008's Deferrals table** gains the four rows naming P0-019, so the
   deviations this audit found are owned rather than noted.

## Prevention

- A Scope bullet that says "construct X and wire it" must have a package for X
  in the Write set. "Registration only" on the composition root plus a scope
  bullet that builds a component is a contradiction, and the implementer
  resolves it by guessing. Here, three times.
- Whenever a write set restricts the composition root, ask what the restriction
  displaces and where it goes. That is the same derivation error as the
  idempotency key with nowhere to live: a brief that requires an artifact and
  declares no place to keep it.
- `internal/application/` is too coarse for a write-set entry in a layer with a
  documented dependency rule. Naming the package (`internal/application/runexec`)
  would have put the inversion in the declaration, where a reviewer reads it,
  instead of in the import graph, where nobody looks.
- Requiring a refusal means declaring the surface that carries it. A problem
  code is a contract surface exactly like a path or a port signature. A brief
  that mandates a stable error and declares no code for it has under-declared
  its own scope — and the implementation will either leave the write set or say
  something untrue.
- A revert is evidence. Round 3 reverting round 2's problem code is the
  implementer discovering the missing declaration and retreating rather than
  reporting it; the task skill asks for the second, and a brief that cannot be
  satisfied is a finding to raise, not a puzzle to solve quietly.
