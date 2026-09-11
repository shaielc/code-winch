# Post-mortems

Records of *plan* failures: defects whose root cause is the workplan rather than
an implementation. Write one when the cause is a seam between two briefs, an
invariant with no owning task, or an assumption one brief made about another's
output. Ordinary implementation bugs do not belong here.

Each record names the briefs involved, explains why no single one is wrong on
its face, and states what plan rule would have caught it.

| Record | Failure |
|---|---|
| [2026-09-09 idempotency with nowhere to live](2026-09-09-idempotency-with-nowhere-to-live.md) | P0-006 owns an API operation whose specification requires durable deduplication, and declares no storage to keep it in |
| [2026-09-11 a store profile derived from test doubles](2026-09-11-a-store-profile-derived-from-test-doubles.md) | P0-012 read its port list off `internal/adapters/memory`'s contents rather than off the application's ports, wiring a port P0-009 owns and omitting two the run path needs; P0-013 to P0-018 inherited it |
