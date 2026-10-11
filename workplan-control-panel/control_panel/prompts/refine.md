Refine the brief for $id: $title.

Read the system record at docs/state.md, the task brief selected from the active
tracker at docs/workplan/$brief, skills/refine-task/SKILL.md,
and all applicable AGENTS.md files.
Create a new branch from `$branch`. Refine the brief as
skills/refine-task/SKILL.md describes, change only docs/workplan/$brief, and commit
there. Do not implement the task. Open a pull request whose base is `$branch`.

Include `Task: $id` in the pull request body and no other task ID. Do not edit
status fields in docs/workplan/tasks.json; automation stamps `completed` only when the
final task pull request into main is approved.
