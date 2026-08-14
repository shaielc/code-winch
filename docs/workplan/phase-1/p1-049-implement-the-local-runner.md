# P1-049: Implement the local runner

**Phase:** 1 — Local single-user vertical slice
**Shape:** seam
**Dependencies:** P1-013 (compile: the local driver holds the PTY this task exposes), P1-014 (compile: `HarnessDriver`/`HarnessCodec` the pump drives), P0-006 (compile: `protocol.RunnerMessage`), P0-008 (contract: the shared sandbox suite this task extends)

## Objective

`winch dev run` launches the fake harness under the local sandbox, streams its
output to the terminal, forwards typed input to it, and reports its exit — the
first time any byte moves between a harness and Code Winch.

## Scope

- Add attached I/O to the sandbox port: one operation returning an
  `io.ReadWriteCloser` bound to a single execution handle, plus an explicit
  capability flag. Implement it in the local driver (the PTY it already owns)
  and the fake driver (a scripted pipe). Update `docs/code-structure.md` §3 in
  the same change, per the contract rule.
- Extend `test/contract/sandbox` so every driver is checked for read-after-write,
  close semantics, behavior after process exit, and whether attach is single-use
  or idempotent. A driver that advertises the capability and returns a nil
  stream fails the suite.
- `internal/runner/local`: the only component holding execution handles and
  codecs. It implements `application.RunnerGateway` and the
  `ReconciliationRunner` that P1-020 assumed and no code supplies. It handles
  `prepare`, `start`, `input`, `resize`, `stop`, `inspect`, and `cleanup`,
  rejects unknown kinds and stale lease tokens, and assigns a monotonic
  runner-local ordinal — never a canonical sequence.
- `cmd/winch` with a `dev run` command that constructs the drivers and runner
  in process, with no daemon and no database, so the pump is observable by hand.

## Non-goals

- Persisting anything. `dev run` prints observations; the supervisor and store
  are P1-050's concern.
- Remote transport and runner registration — P5-041, P5-042.

## Runtime reachability

`cmd/winch`, `winch dev run --harness fake --sandbox local`. The same
`internal/runner/local` instance is registered into `cmd/winchd` by P1-050.

## Owned surfaces

`internal/runner/local/`, `cmd/winch/`, `internal/application/adapters.go`
(sandbox port), `internal/adapters/sandbox/local/`,
`internal/adapters/sandbox/fake/`, `test/contract/sandbox/`,
`docs/code-structure.md` §3.

## Demonstration

    $ go build -o /tmp/winch ./cmd/winch && go build -o /tmp/fake-harness ./cmd/fake-harness
    $ PATH=/tmp:$PATH /tmp/winch dev run --harness fake --sandbox local
    > echo hello
    → expect: "hello" printed back, as a decoded event line rather than raw bytes
    > exit
    → expect: an exit observation with a successful outcome, and the command returns

    $ PATH=/tmp:$PATH /tmp/winch dev run --harness fake --sandbox local --stop-after 1s
    → expect: stop escalation runs; `pgrep -f fake-harness` finds nothing afterwards

## Verification

- Standing sandbox contract suite passes for the local and fake drivers,
  including the new attach cases.
- Runner test against the fake harness asserting ordered observations for start,
  output, and exit, and that input sent before the process is ready fails with a
  stable code instead of being dropped.
- Race-enabled pump test; orphan-cleanup test after a forced stop.

## Acceptance criteria

- [ ] A person sees harness output and gets a reply to typed input, by hand,
      from one command.
- [ ] `Flush` runs before the exit observation, so trailing output is never lost.
- [ ] Stale lease tokens and unknown message kinds are rejected with stable,
      content-free errors.
- [ ] Forced stop leaves no descendant process.
- [ ] The runner assigns no canonical sequence numbers and writes no events.

## Deferrals

| Deferred | Owning task |
|---|---|
| Registering this runner in the daemon composition root | P1-050 |
| Docker driver implementing the same attach contract | P3-029 |
| Hosting the runner in a standalone `cmd/winch-runner` binary | P5-041 |

## Traces to

`docs/architecture.md` §3, §4 (sandbox drivers, event pipeline);
`docs/code-structure.md` §3, §5; `docs/contracts.md` §4;
`docs/decisions/0003-capability-based-adapters.md`
