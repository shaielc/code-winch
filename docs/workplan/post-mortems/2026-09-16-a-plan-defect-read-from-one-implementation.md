# 2026-09-16 — A plan defect read from one implementation

**Found:** comparing two independent implementations of P0-008, each with its
own audit, before either merged. Pull request #53 was implemented in three
commits and audited in pull request #54 on 2026-09-15. Pull request #55 was
implemented in one commit, `4f61c0e`, written without knowledge of #53, and
audited on 2026-09-16.
**Symptom:** #54 recorded P0-008's brief as unsatisfiable — "Every round is a
faithful reading of the brief. The brief cannot be satisfied." — and opened
P0-019 to own the consequences. #55 satisfied the brief exactly where #54 said
it could not, and failed its own audit for unrelated reasons, all on failure
paths. What the two runs share is what the plan got wrong, and it is smaller
than #54 claimed.

#54's record (`13e9aee:docs/workplan/post-mortems/2026-09-15-a-coordinator-with-nowhere-to-live.md`)
and its P0-019 never merged. This record replaces the first; the second's ID is
not reused, because #54 still cites it.

## The two runs

#53 went through a reviewer agent between commits. The reviewer's requests are
recorded nowhere: the pull request has no reviews or comments, and all three
commits have empty message bodies. What drove each round is therefore unknown,
and this record does not guess. The table shows what each commit changed.

| Concern | #53 `f558bdb` | #53 `d8fba75` | #53 `b2cf802`, + `6229967` on #54 | #55 `4f61c0e` |
|---|---|---|---|---|
| Unsupported-profile refusal | 422 `validation_failed` (`cmd/winchd/main.go:252`) | 400 `unsupported_run_profile` (`httpapi/server.go:305`) | 409 `run_state_conflict` (`cmd/winchd/main.go:284-285`) | 422 `unsupported_profile`, declared (`api/openapi/code-winch.yaml:245`, `httpapi/server.go:306-307`, `docs/contracts.md` §1) |
| Attempt state writes | field assignment (`cmd/winchd/main.go:439`, `internal/application/runs.go:87`) | domain state machine (`runs.go:102`) | unchanged | domain state machine via `domain.RestoreRun` (`internal/domain/run.go:88`) |
| Redaction before persistence | none: `persistRedactor` returns every event (`cmd/winchd/main.go:364-365`) | refuses `secret` and invalid JSON (`internal/supervisor/coordinator.go:45-46`) | unchanged | the envelope rule of `docs/contracts.md` §2 (`internal/application/redaction.go:21-23`) |
| Daemon fake-harness controls | `EarlyExit: true`, nothing else (`cmd/winchd/main.go:91`) | `os.Getenv` (`cmd/winchd/main.go:121`) | unchanged | configuration loader, `WINCH_FAKE_HARNESS_*` (`internal/platform/config/config.go:68-79`) |
| Runs active at shutdown | not handled | not handled | bounded `Close` (`internal/application/runexec/coordinator.go:65`) | `Shutdown` (`internal/application/start.go:268`) |
| A launch that fails | run left `queued` or `preparing` behind a held lease | unchanged | unchanged; `abandon` added in `6229967` (`coordinator.go:114`) | `failPreparation` and `abandon` (`start.go:164`, `:177`) |
| Where orchestration lives | `cmd/winchd/main.go` | `internal/supervisor/coordinator.go` | `internal/application/runexec/`, importing `internal/supervisor` (`coordinator.go:14`) | `internal/application/start.go`, behind its own `RunExecutor` port (`start.go:36`), adapted in the composition root (`cmd/winchd/main.go:266`) |
| Audit verdict | — | — | complete, plus P0-019 (#54) | not complete, three findings |

Line numbers are at the commit named in the column header.

## What #54 claimed, and what #55 shows

- **"The brief cannot be satisfied" and "There was no fourth option"**
  (`:8`, `:63` of #54's record). The claim was that orchestration had no
  package the write set allowed. #55 wrote orchestration as the start use case
  in `internal/application/`, which the write set names, against a port that
  package defines, and adapted the supervisor to that port in the composition
  root. Nothing under `internal/application/` imports `internal/supervisor`,
  and #55's audit raised no finding about placement. **Withdrawn.**
- **"The Scope requires a refusal the Contract surfaces cannot express."** The
  brief declares `POST /api/v1/runs/{runId}/start` as a contract surface, and
  this directory's 2026-09-09 record already holds that owning an operation
  includes its documented error responses. What the brief omitted were the
  files that carry the response. #55 wrote them, disclosed the writes in its
  pull request, and updated `docs/contracts.md` in the same change. Its audit
  accepted the deviation as disclosed. **Withdrawn.** The write-set omission
  itself is real, and is item 3 below.
- **"Every round is a faithful reading of the brief"** (`:8`) and **"the
  implementer discovering the missing declaration and retreating rather than
  reporting it"** (`:159-161`). These are statements about intent, inferred
  from diffs of a review loop whose requests were never recorded.
  **Unsupported.**
- **The launch-failure wedge.** #54 reproduced it against `b2cf802` with a
  nonexistent harness binary and fixed it in `6229967`. **Holds**, as an
  implementation bug.

## What was implementation

The table's first column is most of #53's story. Its first commit missed
requirements the brief and the design set already stated:
- redaction before persistence (`docs/architecture.md` §4,
  `docs/contracts.md` §2);
- lifecycle changes through the domain state machine;
- a controllable fake profile (I3).

The next two commits repaired those. Orchestration moved in the same commits,
whose subjects are "Harden run execution lifecycle" and "Bound daemon shutdown
for active runs". The rounds never settled the refusal: it went through three
codes and ended on one that calls a valid state invalid. And no round handled a
launch that fails.

#55 met the same requirements in one commit. What it missed is on failure
paths, not the main path.

None of that is a plan defect. #54's error was to read it as one.

## What the plan got wrong

Each item below is visible in the brief's own text, or showed up in both runs.
All five could have been found before either implementation started.

1. **A deferral that never reached its owner's scope.** P0-003 defers
   "Daemon-backed runs using the fake profile" to P0-008
   (`phase-0/P0-003-controllable-fake-harness.md:105`). P0-008's scope
   constructs the fake harness and says nothing about controlling it. #53's
   first commit hardcoded early exit and took a round to add controls. #55
   found the deferral in P0-003 and cited it. This is a seam between two briefs.
2. **An objective that promises a guarantee, and criteria that never test it.**
   - **The promise and the criteria.** The objective says events "are durably
     stored". The acceptance criteria cover the happy path, the refusal, and
     outbox rows. None says what happens when an append fails, a start is
     cancelled, a launch fails, or ownership is lost.
   - **Both runs shipped a failure-path defect**, on different paths. #53's
     launch failure survived three review rounds. #55 handled launch failure
     and lost output on a failed append.
   - **The two audits probed different failures.** #54 used a missing harness
     binary. #55's audit used a failing append, a lease acquisition delayed
     under a cancelled request, and a lease takeover. Neither ran the other's
     probes.
   - **They also graded against different things.** #54 graded the criteria
     ("all six acceptance criteria hold at `b2cf802`") and returned complete.
     #55's audit graded the objective ("violates the task's durable-output
     objective") and returned not complete.

   With the brief silent on failures, each audit decided for itself which
   failures counted, and the verdict followed the auditor.
3. **A write set that omits what the scope requires.** A refusal with a stable
   code needs the OpenAPI document and the transport adapter's problem
   mapping. Daemon fake controls need the configuration loader, and new
   commands need operator documentation. None of these is listed. Both runs
   wrote `internal/adapters/transport/httpapi/server.go` and
   `deployments/README.md` (`d8fba75`, `4f61c0e`).
4. **A write-set note that contradicts the scope.**
   - **The note.** `cmd/winchd/main.go` is limited to "registry registration
     only", yet the scope implements `StartRun` and `ListRunEvents` on the
     delegating backend. That backend lives in the same file
     (`origin/main:cmd/winchd/main.go:200`, `:274`).
   - **The registry does not exist.** The brief's own non-goals rule out any
     registry.
   - **What both runs did.** Each added over 170 lines to that file:
     `f558bdb` +170 −12, `4f61c0e` +174 −12.
5. **A demonstration check that cannot pass as written.**
   `ps -eo pid,cmd | grep -c 'fake[-]harness'` counted the operator's own shell
   in both runs; #54's and #55's pull request bodies each report `1` for that
   reason. The deployment image also has no `ps`, per #55's audit. P0-011
   carries the same line (`phase-0/P0-011-stop-run.md:60`).

## Process failures

1. **A root cause from one sample.** #54 attributed #53's churn to the brief
   using only #53's commit history. It could not see the review loop and had
   no second implementation to compare. Its two strongest claims, "cannot be
   satisfied" and "no fourth option", assert that something does not exist,
   and it recorded no search for a counterexample.
2. **An audit that changed what it audited.** In the same commit (`13e9aee`),
   #54 created a task and added four rows naming it to P0-008's Deferrals table,
   and it returned the verdict complete. The task skill reserves revision tasks
   for "the code of an already-completed task" (`skills/task/SKILL.md:159-161`).
   P0-008 was pending and its implementation unmerged, so the findings were
   reasons #53 was not complete.
3. **The task skill contradicts itself on task IDs.**
   - **The contradiction.** Its Finish section requires the pull request body
     to contain "no other task ID", and the next bullet asks for "every deferral
     with its owning task ID" (`skills/task/SKILL.md:114-117`).
   - **What the gate enforces.** Only the first:
     `scripts/task_scheduler.py:137-143` resolves a pull request only when
     exactly one known ID appears.
   - **The effect.** #55 reported its deferrals by ID, and its `validate` check
     failed.
4. **The task skill weights the write set beyond what it models.**
   - **What the model says.** A write set "models merge contention and nothing
     else" (`skills/shared/workplan-model.md:320-321`).
   - **What the task skill says.** It raises the write set at five points:
     - orientation calls it "this change's boundary" (`skills/task/SKILL.md:36-37`);
     - implementation calls a write outside it "usually scope creep" (`:71-72`);
     - the pull request must report each outside write (`:115-117`);
     - audit item 5 checks that the change "stayed inside the declared write
       set" (`:141-143`);
     - the final checklist repeats it (`:171`).
   - **What followed.** #54 treated the write set as a wall and built its root
     cause on it. #55 spent a section of its pull request justifying each write
     outside the set, and its audit spent a paragraph judging them. None of that
     effort serves merge contention, the one thing the write set models. Scope
     is already judged against Scope and Non-goals, and meaning against Contract
     surfaces.
5. **No step checks a brief against the code and the plan before
   implementation.**
   - **The hypothesis nobody tests.** The workplan skill says a brief written
     ahead of its dependencies "is a hypothesis, and contact with the code is
     where it gets tested against reality" (`skills/workplan/SKILL.md:43-46`).
     No step does that testing.
   - **What orientation covers.** The task skill's orientation
     (`skills/task/SKILL.md:25-40`) reads the brief, checks dependencies, and
     runs the demonstration, then implementation begins.
   - **What it misses.** It never checks the brief against other briefs'
     deferrals, against where the code lives at HEAD, or against its own
     objective.
   - **The cost.** Every item in *What the plan got wrong* surfaced only after
     implementation, through two runs, a review loop, and two audits. The audits
     ended up deciding what the brief should have said: #55's audit graded
     against the objective because the criteria did not cover it. Judging the
     brief's quality is not an audit's job; an audit judges the criteria it is
     given.

## Why no gate caught it

- `make check` is green on both pull requests. Write sets, contract surfaces,
  and deferral tables are prose, and nothing in CI compares a diff against them.
- Nothing gates failure behavior. The standing scenario is the happy path.
- Nothing reviews a brief against the code and the plan before a task is
  dispatched (*Process failures* 5).
- Nothing checks a post-mortem's causal claim. The only gate that fired in this
  episode was `validate` on #55, and it fired on the task skill's own
  contradiction.

## Consequences if left as is

- P0-019 would own fixes to code that exists only on #53's branch.
- #55's three findings would have no owner. P0-008 could not complete without
  absorbing them, or would complete with them unrecorded.
- P0-009 and P0-011 are written to the same template: happy-path criteria over
  the same observation pump. P0-011 also carries the same `ps` line.
- Until briefs are checked before implementation, the next audit would again
  decide which failures count.

## Remediation

Applied with #55, pending merge:

1. **#55's audit findings have owners.** The decision to defer them was taken
   outside the audit. Each is a failure or recovery guarantee, which the
   model's split test admits as a separate task, and #55's main path met the
   brief. P0-008's Deferrals table names the three owners. It is the only
   change to that brief.

   | Finding in `4f61c0e` | Reproduction in its audit | Owner |
   |---|---|---|
   | A failed append is logged and dropped (`cmd/winchd/main.go:122-125`), and the run still completes | one-shot trigger on `run_events`: `final state=completed; stored events=5; output present=false` | P0-020 |
   | A start whose request is cancelled during lease acquisition is left `queued` (`internal/application/start.go:121-133`) | trigger delaying acquisition, request cancelled at 100 ms: `state=queued; retry status=409; code=run_state_conflict` | P0-021 |
   | Attempt transitions go through the unfenced `RunRepository.Save` (`start.go:307`) | replacement owner takes the lease: `old owner observation error: stale lease`, then `stale owner changed the attempt to completed` | P0-022 |

Proposed, not applied — corrections to P0-008's brief for items 1 and 3 to 5
of *What the plan got wrong*:
- **Scope:** carry P0-003's deferral.
- **Write set:** list the files the scope requires, and replace the
  "registration only" note.
- **Contract surfaces:** declare the refusal's problem response and the
  fake-harness configuration keys.
- **Traces to:** add `docs/contracts.md` §1–2.
- **Demonstration:** use a process check that runs in the deployment image.

Both implementations were built and audited against the brief as it stands.
Rewriting its scope and surfaces to match #55, inside #55's own pull request,
would repeat *Process failures* 2. These corrections belong to the refinement
step described under *Prevention*, when a task is next dispatched from this
brief. The same `ps` line in P0-011 is left for that task's refinement.

Not changed here:
- #53 and #54 are still open; closing them is a separate decision.
- #55's pull request body still names more than one task ID.
- *Process failures* 3, the task-ID contradiction, is recorded, not fixed.
  *Process failures* 4 and 5 are addressed by workflow improvements 1 to 5
  below, applied to `skills/`, `scripts/task-prompt.md`, and the runner in this
  change.

## Prevention

- **Blame the brief only with evidence the brief determines:** two sentences in
  it that cannot both hold, or the same defect in an independent
  implementation. A pattern in one implementation's history is an
  implementation finding. A claim that no alternative exists records the
  search for one.
- **An audit reports.** Creating tasks, adding deferral rows, and editing the
  brief under audit are separate changes, decided by a person.
- **Refine each brief against the code and the plan before implementation
  starts.** Every item in *What the plan got wrong* was checkable at that point:
  - the brief's assumptions about existing code hold at HEAD;
  - every deferral that names the task is carried into its scope;
  - every guarantee the objective states has failure criteria, or is narrowed
    out, or goes to a hardening task;
  - the demonstration can run where it will run.

  The refinement corrects the brief itself and makes its Scope concrete, as the
  first commit on the task's branch.
- **An audit judges the acceptance criteria, not the brief.** Whether the
  criteria cover the objective is settled before implementation. A gap an
  audit finds anyway is reported as a finding about the brief, not folded into
  the verdict.
- **Weight the write set as what it is: a merge-contention forecast.** A write
  outside it is disclosed so the plan's collision table stays true. Scope creep
  is judged against Scope and Non-goals, and changes of meaning against Contract
  surfaces.
- **Record a review loop where later readers can see it:** the reviewer's
  requests go on the pull request. Without them, the cause of a round is
  guesswork, and guesswork is what #54 wrote down.


## Workflow improvements

Items 1 to 5 are applied in this change: `skills/task/SKILL.md`,
`skills/workplan/SKILL.md`, `skills/shared/workplan-model.md`,
`skills/README.md`, `scripts/task-prompt.md`, `runner/control_panel.py`, and
`runner/README.md`. Items 6 and 7 are suggestions only.

  1. Brief refinement, a new first step of every task:
     - Dispatched by the scheduler and the control panel in place of implementation.
     - Checks the brief against HEAD and the plan: assumptions about existing code, deferrals into the task, the objective against the criteria, and whether the demonstration can run where it will run.
     - Makes the brief concrete in its existing sections, chiefly Scope, including items for work beyond the task's own code.
     - Produces one commit that changes only the brief, the first on the task's branch. Implementation starts from it.
     - Is done by neither the implementer nor the auditor.
  2. Briefs:
     - If an objective promises durability, fencing, recovery or no orphans, the criteria name the failures that would break it.
     - Derive each write-set entry from a scope bullet.
  3. Review:
     - The reviewer agent posts its requests on the PR.
     - If a concern reverses twice (#53's refusal code changed three times), escalate to a person instead of doing another round.
  4. Audit:
     - Judge the acceptance criteria one at a time, as written. The brief's quality is not the audit's question.
     - Report a gap between the objective and the criteria as a brief finding, not as a reason for the verdict.
     - The auditor only reports. Deferring findings or editing the plan is a decision a person makes.
  5. Write-set wording in the task skill:
     - State once what a write set is: a merge-contention forecast.
     - Drop "this change's boundary", "usually scope creep", and "stayed inside the declared write set" from orientation, implementation, audit item 5, and the final checklist.
     - Keep one instruction: list each write outside the set in the PR, so the collision table can be updated.
     - Keep contract surfaces at full weight.
  6. Post-mortems:
     - Blame the brief only if it contradicts itself on its face, or an independent implementation hits the same defect.
     - A claim that something is impossible must show the search for a counterexample.
     - For high-risk seams, compare two independent implementations, as this record did.
  7. Tooling:
     - Fix the task-ID contradiction: either have the scheduler read only the Task: line, or have PRs point to the brief's Deferrals table instead of listing IDs.
     - Add an informational CI note listing files a PR changes outside its brief's write set, to feed the collision table.