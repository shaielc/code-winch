# P1-062: Run against the Codex harness

**Phase:** 1 — Local single-user vertical slice
**Shape:** swap
**Dependencies:** P1-014 (compile: the `codex.Driver` and `Config` this constructs and registers), P1-050 (compile: `DefaultDrivers.RegisterHarness` and the profile resolution that turns a run's `harnessProfile` into a driver), P1-053 (semantic: a swap is accepted by re-running the standing suite against the new substrate, and there is no suite to re-run until this lands)

## Objective

`harnessProfile: "codex"` resolves to the Codex adapter and produces a run whose
events came out of the real Codex JSON-lines protocol, instead of being refused
as an unknown profile.

## Scope

- `internal/adapters/harness/codex/register.go`, the one-line `init()` its
  siblings already have, plus the blank import in `cmd/winchd`. P1-014 built the
  descriptor, launch construction, codec, and exit mapping; nothing has ever
  reached them, because the package is imported by nothing outside its own
  tests. The registry P1-050 builds is append-only precisely so this costs two
  lines — but nobody has appended them, so the only harness a running daemon can
  resolve is `fake`.
- Derive input capabilities from the resolved harness descriptor as well as the
  run state. `execution.Capabilities` reads state alone, so it offers `text` to
  every `Running` run. Codex advertises `incremental-input: false` — `codex
  exec -` reads one prompt and stops reading — while its codec's `Encode` will
  frame a follow-up regardless. Registering the adapter without this change
  accepts input, records it durably, and delivers it into a pipe no process is
  reading. `InputErrorUnsupported` already exists and is already refused at
  acceptance; only the capability computation is missing.
- Give that refusal its own problem code. It currently maps through to
  `run_state_conflict`, whose detail tells the caller to retry when the state
  changes — advice that will never come true for a capability the adapter does
  not have. An operator has to be able to tell "not yet" from "not ever".
- A `codex` profile for the standing suite that puts a protocol-faithful fake
  `codex` on `PATH` — the `testdata/fake-codex.sh` fixture P1-014 already ships,
  promoted out of `testdata` so the suite can reach it. Every line of adapter
  code runs; only the vendor process is substituted, so CI still needs no
  account.
- Select scenarios by declared capability rather than by substrate name, so the
  codex profile skips the input step because the descriptor says so. P1-053
  forbids substrate conditionals inside scenarios, and a capability-limited
  adapter in the registry is the first thing that would tempt one.
- A missing executable must produce a terminal failed run naming the binary,
  within the start timeout, rather than a run parked in `Preparing`.
- Record in the adapter's README what the fake CLI does not prove — its record
  shapes are ours, not the vendor's — and what the profile requires of its host.

## Non-goals

- Model pinning, or any per-run harness option. The run contract carries a
  profile name and nothing else, and that name is the selector: a second model
  is a second registered profile, not a new request field.
- Credential acquisition or injection. `BuildLaunch` rejects resolved
  credentials outright, so this profile works only where the host user has
  already logged the CLI in — the same `local-trusted` precondition P1-050
  records. Acquisition is P2-058, injection P3-032.
- Disabling unsupported controls in the web app, and the two-adapter capability
  matrix — P2-025. This task refuses at the API; that one makes the refusal
  visible before a user tries.
- Installing the vendor CLI into the shipped image. `deployments/` is P1-048's
  surface, and a vendor binary in the default image is a supply-chain decision
  no task owns.

## Runtime reachability

`cmd/winchd` with `harnessProfile: "codex"` on `POST /api/v1/runs`, on the
compose stack; `make e2e HARNESS=codex` for the suite. The compose image carries
no `codex`, so there a started run fails at launch — which is what the
missing-executable criterion pins down. The real-provider check runs against a
daemon on a host that has one.

## Owned surfaces

`internal/adapters/harness/codex/` (`register.go`, `README.md`, and the fixture
moved out of `testdata/`), `internal/execution/capabilities.go` and its tests,
the `codex` profile under `test/e2e/`.

Two lines land in surfaces other tasks own, and both are why those tasks are
dependencies rather than peers: the blank import in `cmd/winchd/main.go`
(P1-050) and the unsupported-input problem code in `api/openapi/components/`
(P1-050's split). The e2e profile is appended through the switch P1-053 builds,
which is append-only, so it is not a claim on that task's files.

## Demonstration

    $ ID=$(curl -fsS -X POST localhost:8080/api/v1/runs -H "X-CSRF-Token: $T" … \
        -d '{"workspacePath":"/workspace","harnessProfile":"codex","sandboxProfile":"local"}' | jq -r .id)
    → expect: 201

P1-050 refuses an unresolvable profile at creation with 422, so a created run is
itself the proof that the registry resolved this one — no separate listing
endpoint is needed to see that registration worked.

    $ curl -sS -o /dev/null -w '%{http_code}\n' -X POST localhost:8080/api/v1/runs/$ID/input \
        -d '{"kind":"text","text":"hi"}' …
    → expect: a refusal naming the unsupported input kind, distinct from
      run_state_conflict; no command recorded against the run

    $ curl -fsS localhost:8080/api/v1/runs/$ID/events | jq -r '.events[].source.adapter' | sort -u
    → expect: codex

    $ make e2e HARNESS=codex
    → expect: the standing scenarios pass unchanged, and each skipped scenario
      names `incremental-input` as the capability that excluded it

    $ make e2e
    → expect: the fake profile still passes, unchanged

On a host with the real CLI logged in, with the daemon run outside the image:

    $ codex --version
    $ … create and start a run with harnessProfile codex …
    → expect: model output in the event stream carrying the `openai.codex/v1`
      extension, and a zero exit mapped to a successful terminal state

    $ … start a codex run with the binary absent from PATH …
    → expect: a terminal failed run naming the executable, and pgrep finds
      nothing left behind

## Verification

- Standing scenario suite passes against the `codex` profile and still passes
  against `fake`.
- A registry test asserting `codex` resolves from a composed daemon. This is the
  only check that catches the blank import being dropped by a later import
  tidy — which is how the adapter became unreachable in the first place.
- Capability tests both ways: a run whose harness advertises
  `incremental-input: false` is refused text input; one that advertises it is
  not.
- The existing codex contract, transcript, and fake-CLI tests still pass after
  the fixture moves.

## Acceptance criteria

- [ ] A run created with `harnessProfile: "codex"` starts, and its events carry
      `source.adapter: "codex"`.
- [ ] Every harness driver in the registry is reachable from a running daemon —
      no harness package is imported only by its own tests.
- [ ] An input kind the resolved descriptor does not advertise is refused at
      acceptance with a code distinguishable from a state conflict.
- [ ] The standing suite runs against the codex adapter with no vendor account,
      and what that substitution does not prove is written down in the same
      change.
- [ ] A missing executable produces a terminal failed run naming it and leaves
      no process behind.

## Deferrals

| Deferred | Owning task |
|---|---|
| Login modes in this descriptor, and acquiring the credential | P2-058 |
| Credential values reaching `BuildLaunch` | P3-032 |
| Capability-driven UI and the two-adapter capability matrix | P2-025 |
| Running this adapter under Docker isolation | P3-029 |

## Traces to

`docs/roadmap.md` Phase 1 ("One harness adapter, start/input/resize/stop");
`docs/architecture.md` §4 (harness adapters); `docs/code-structure.md` §4;
`docs/decisions/0003-capability-based-adapters.md` ("reject unsupported requests
rather than silently degrading")
