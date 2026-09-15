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
| [2026-09-15 a coordinator with nowhere to live](2026-09-15-a-coordinator-with-nowhere-to-live.md) | P0-008's scope requires an orchestrator its write set gives no package, and a refusal its contract surfaces cannot express |
