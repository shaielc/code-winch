# P0-003: Session records persist, and a reload shows what the harness already said

**Phase:** 0 — A sandbox you can talk to
**Shape:** seam
**Dependencies:** P0-002 (contract: the session record shape and the runner-local
ordinal namespace this persists and serves)

## Objective

Reload the page after the harness has spoken and the earlier records are there, because
the runner wrote them to a store of its own and the page fetches from an ordinal before
joining the live stream.

## Scope

- `internal/runner/store.go` — the sandbox's private ordered record store. Choose the
  simplest substrate that persists a single session's ordered records, per roadmap
  deferred decision D3; a sandbox that needed PostgreSQL to start would not be a sandbox
  you can start by hand (`docs/architecture.md` §8).
- Every record the runner produces is appended before it is published, so a reader that
  arrives later sees what an earlier reader saw.
- `GET /api/session/records?after_ordinal=N` — the snapshot half of §8's "snapshot plus
  stream", with a bounded page size.
- The page loads the snapshot, then opens the stream, and shows one continuous list.
- `winch records --after <ordinal>`.
- `WINCH_SANDBOX_STORE_PATH` and a compose volume, so the store survives a container
  restart as well as a browser reload.

## Non-goals

- Mid-session resume with a no-gap guarantee, a caught-up marker, and heartbeats.
  `docs/roadmap.md` §2 puts "reconnect mid-session and lose nothing" at Stage 2; Stage 0
  is "the exchange survives a reload". A reload refetches from ordinal 0.
- History from a session that has already ended, and any question of whether the store
  outlives one session — roadmap deferred decision D3 names that as its own trigger.
- Cross-session history, search, export, and retention policy. Roadmap Stage 2 and §3's
  deliberate absences.
- Canonical sequence numbers. The control plane assigns those, roadmap Stage 4, and
  whether the envelope wraps or replaces this record shape is deferred decision D7.

## Runtime reachability

`cmd/winch-sandbox` constructs the store from `WINCH_SANDBOX_STORE_PATH` and hands it to
the runner; the `sandbox` service mounts the volume behind it.
`winch records --after 0` and reloading the page are the two hands-on paths.

## Write set

- `internal/runner/store.go`, `internal/runner/session.go`, `internal/runner/runner.go`
- `internal/adapters/transport/attach/records.go`, `.../server.go`
- `cmd/winch-sandbox/main.go`, `cmd/winch-sandbox/config.go`
- `cmd/winch/records.go`
- `web/src/attach/App.tsx`, `web/src/attach/useSessionRecords.ts`
- `deployments/compose.yml`
- `test/e2e/scenario_records_survive_reload_test.go`

## Contract surfaces

- port: the session record store — append and read-from-ordinal
- the durability of the ordinal: an ordinal, once issued, names the same record for the
  life of the session
- API: `GET /api/session/records?after_ordinal=`
- config key: `WINCH_SANDBOX_STORE_PATH`
- CLI: `winch records`

## Demonstration

    $ docker compose -f deployments/compose.yml up --build -d
    $ sleep 2 && ./bin/winch records --after 0
    → expect: the records the harness emitted while nothing was watching

    $ ./bin/winch records --after 3
    → expect: ordinals 4 onward only

    $ xdg-open http://127.0.0.1:8080     # then reload the page
    → expect: the same records are present after the reload, in the same order

    $ docker compose -f deployments/compose.yml restart sandbox
    $ ./bin/winch records --after 0
    → expect: the records from before the restart are still there

## Verification

- Standing scenario suite passes against the persistent store — the same scenarios
  P0-002 added still pass, which is the parity this task is accountable for.
- `scenario_records_survive_reload_test.go`: drive the harness, read the stream, drop the
  stream, refetch from ordinal 0, and assert the two agree.
- Store unit tests: append-then-read round trip, read from an ordinal beyond the end
  returns empty rather than erroring, read from a negative or non-numeric ordinal is
  rejected.
- `make check`, `make e2e`, `make test-cycle`, `cd web && npm test`.

## Acceptance criteria

- [ ] A record is in the store before it reaches any stream reader, so a reader that
      connects immediately after a record is published can refetch it. Inject by
      publishing a record with no reader attached, then fetching from ordinal 0.
- [ ] Reloading the page loses no record and duplicates none. The list after a reload
      equals the list before it.
- [ ] `after_ordinal` beyond the last issued ordinal returns an empty page with a
      success status, not an error.
- [ ] A restart of the `sandbox` service preserves records written before it.
- [ ] The store is reachable without any external service: `docker compose up` with no
      database and no network access to anything still serves records.
- [ ] Ordinals are still gap-free after a restart — a restart does not reissue an
      ordinal already used.
- [ ] P0-001's and P0-002's demonstrations still pass unchanged.

## Deferrals

| Deferred | Owning task |
|---|---|
| A submission of your own appearing in this record list | P0-005 |
| Diagnostic records from malformed harness output landing in this store | P0-007 |

## Traces to

`docs/roadmap.md` §3 Stage 0, §4 output path, §6 D3 and D7;
`docs/contracts.md` §8 (*Records are ordinal-ordered*, *Reading is snapshot plus
stream*); `docs/architecture.md` §5 Topology A, §8; `docs/code-structure.md` §3;
ADR-0005 (*No store in the sandbox* — rejected because reload is the cheapest
observation of persistence there is).
