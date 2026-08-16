# Fake harness scenarios

Version 1 scenarios are ordered JSON actions validated by
`v1/fake-harness-scenario.schema.json`. `emit` writes a normal JSON-lines
record, `wait_input` blocks until an input's text matches a Go regular
expression, `sleep` waits for a number of milliseconds, `exit` selects the
process status, and `malformed` writes deliberately invalid but newline-framed
data. String values may contain `{{random}}`; the default seed is deterministic,
an integer selects a repeatable sequence, and `--seed off` uses system entropy.

The flags `--latency`, `--fail-on-input`, `--truncate-record`,
`--oversized-record-bytes`, `--early-exit`, and `--ignore-sigterm` inject faults
without changing a fixture. Durations use Go syntax such as `250ms` or `2s`.

This profile proves orchestration and parser behavior against output shapes we
chose. It is not a provider and does not reproduce model latency distributions,
rate limits, authentication, provider semantics, or provider output diversity.
