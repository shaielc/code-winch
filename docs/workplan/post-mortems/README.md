# Plan post-mortems

Records of defects whose root cause is the **plan** rather than an implementation: a seam
between two briefs, an invariant with no owning task, an assumption one brief made about
another's output, a write collision whose reconciliation needed more than a rebase.

Ordinary implementation bugs do not belong here. The test is whether any single brief is
wrong on its face. If one is, fix that brief. If none is and the defect still happened, the
gap is between them, and that is what this directory is for.

One file per record, named `<date>-<slug>.md`, each stating:

1. the briefs involved, by ID;
2. why no single one of them is wrong read on its own;
3. what a person actually observed, with a command or a `file:line`;
4. **the plan rule that would have caught it.**

Point 4 is the reason the directory exists. Those rules are the most durable thing a plan
produces, and `skills/workplan/SKILL.md`'s close procedure restates every one of them in
`docs/state.md` before this directory is removed — which is the only reason removing it is
safe.

Empty is the expected state.
