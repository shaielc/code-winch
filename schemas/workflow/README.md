# Workflow definition schema

`v1/workflow-definition.schema.json` is the source-of-truth wire schema for
declarative workflow definitions. The Go validator additionally checks graph
cycles, reference targets/output types, duration syntax, and policy bounds.

Definitions contain data only: step `type` is a closed enum and unknown fields
are rejected. `parallel` and `foreach` always require `maxConcurrency` (1–100).
Instances must persist the definition ID/version and named harness and sandbox
profiles using `workflow.InstancePins`; profile names are safe identifiers and
must never contain credentials.

Validation diagnostics expose a stable code and JSON path. They intentionally
exclude literal values, messages, artifacts, and secrets, so they are safe to
attach to logs and traces alongside a caller-supplied workflow resource ID.

