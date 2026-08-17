# Post-mortems

Records of *plan* failures: defects whose root cause is the workplan rather than
an implementation. Write one when the cause is a seam between two briefs, an
invariant with no owning task, or an assumption one brief made about another's
output. Ordinary implementation bugs do not belong here.

Each record names the briefs involved, explains why no single one is wrong on
its face, and states what plan rule would have caught it.

| Record | Failure |
|---|---|
| [2026-08-16 migration re-runnability](2026-08-16-migration-rerunnability.md) | Migrations were planned to run once; P1-048 made them run on every boot |
| [2026-08-17 unverifiable stream criterion](2026-08-17-unverifiable-stream-criterion.md) | P1-050 spans five shapes; relieving its size by splitting the CLI downstream left one of its criteria with no witness |
