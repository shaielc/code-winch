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
| [2026-09-16 a plan defect read from one implementation](2026-09-16-a-plan-defect-read-from-one-implementation.md) | An audit blamed P0-008's brief for one implementation's review churn; an independent implementation refutes it, and what both share — a deferral P0-008's scope never carried, and a durability objective with no failure-case criteria — is the plan's |
