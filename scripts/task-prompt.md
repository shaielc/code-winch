Implement $id: $title.

Read skills/task/SKILL.md, then your brief at docs/workplan/$brief, then all
applicable AGENTS.md files. The skill states which sections of the brief bind
you and what your task is not answerable for. Stay inside that boundary: work
another task owns, or incompleteness your change did not introduce, is reported
rather than absorbed.

Run the verification your brief requires and its demonstration by hand, commit,
and open a pull request saying what you ran and what you observed. Report
failures as failures.

Include `Task: $id` in the pull request body and no other task ID. Do not edit
status fields in docs/workplan/tasks.json; automation stamps `completed` when the
pull request is approved.
