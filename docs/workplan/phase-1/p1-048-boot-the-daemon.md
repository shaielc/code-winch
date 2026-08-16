# P1-048: Boot the daemon

**Phase:** 1 — Local single-user vertical slice
**Shape:** seam
**Dependencies:** P1-011 (compile: `postgres.MigrateUp` and the pool the root constructs), P1-017 (compile: the `httpapi` handler and its `Config` the root will mount)

## Objective

`docker compose up` leaves a `winchd` process running that answers
`GET /api/v1/health`, having applied its migrations, instead of exiting
immediately.

## Scope

- `internal/platform/config`: layered resolution as `docs/code-structure.md` §6
  describes — compiled safe defaults, then an optional file, then environment —
  with startup validation that refuses missing or under-length secrets and
  reports every invalid field at once.
- `internal/platform/telemetry`: an `slog` handler restricted to the bounded
  attribute set in `docs/security.md` §5, a redaction helper that drops rather
  than logs a field it cannot redact, and a metric registry pre-declaring the
  `docs/architecture.md` §7 metric names with bounded labels.
- `cmd/winchd/main.go`: config, telemetry, database pool, `MigrateUp`, ID and
  clock sources, event stream, health handler, static asset serving, and
  graceful shutdown with a bounded drain deadline.
- Deliver the CSRF token to the browser by injecting it into the served
  `index.html`, not through a cookie and not through a new API path.
- `make run`, and replace the "the daemon is not wired yet" section of
  `deployments/README.md` with what actually runs.

## Non-goals

- Run use cases, driver registration, and the run API surface — P1-050.
- Dashboards, alerts, SLOs, and runbooks — P5-047. This task ships the
  primitives those consume, not the operational product.

## Runtime reachability

The `winchd` composition root, `compose.yml` as shipped. Every later Phase 1
task extends this root rather than creating one.

## Owned surfaces

`cmd/winchd/`, `internal/platform/config/`, `internal/platform/telemetry/`,
`deployments/compose.yml`, `deployments/README.md`, `Makefile` (`run` and
`web-build` targets), `web/index.html`.

Also, because booting is what first exercises them: `internal/adapters/postgres/`
migration re-runnability (see the 2026-08-16 post-mortem), and the log attribute
keys in `internal/adapters/transport/httpapi/` — the redaction allowlist here and
the keys emitted there have to agree or the fields are silently dropped.

## Demonstration

    $ docker compose -f deployments/compose.yml up --build -d
    $ curl -fsS http://localhost:8080/api/v1/health
    → expect: {"status":"ok"}, and the container still running afterwards

    $ docker compose -f deployments/compose.yml logs winchd
    → expect: migration applied at version 5; no message content, path, or
      secret in any line

    $ WINCH_TOKEN=short docker compose -f deployments/compose.yml up winchd
    → expect: exit before listening, naming the invalid field and not its value

## Verification

- Configuration precedence and validation tests, including a case where several
  fields are invalid at once.
- Telemetry tests asserting that a free-form field and a known secret canary are
  both dropped, and that metric labels stay inside their declared sets.
- Migration bootstrap test: a second start against a migrated database is a
  no-op.

## Acceptance criteria

- [ ] The daemon serves `/api/v1/health` and survives `compose up` without
      exiting.
- [ ] Migrations run at startup; the daemon refuses to serve if they fail.
- [ ] Startup fails closed on weak, missing, or malformed configuration, and the
      error names fields rather than values.
- [ ] Shutdown closes listeners, drains subscribers, and returns within its
      configured deadline.
- [ ] `deployments/README.md` describes the stack that exists, with no
      "not wired yet" caveat remaining.

## Deferrals

| Deferred | Owning task |
|---|---|
| `/api/v1/runs*` endpoints; the health handler is the only route until then | P1-050 |
| Metric emission from the run and outbox paths | P1-050, P2-059 |
| SLOs, dashboards, alerts, runbooks over these metrics | P5-047 |

## Traces to

`docs/architecture.md` §5 Phase A, §7; `docs/code-structure.md` §1, §6;
`docs/security.md` §5; `docs/roadmap.md` Phase 1
