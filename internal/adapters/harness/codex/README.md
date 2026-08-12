# Codex harness adapter

This adapter launches `codex exec --json -`, sends the initial prompt on
standard input, and incrementally consumes the CLI's JSON-lines output. It does
not read provider credentials: authentication must already be available inside
the selected sandbox. Supplying resolved credentials is rejected rather than
silently placing secrets in the environment.

The descriptor advertises structured output and explicitly reports resume and
incremental follow-up input as unsupported. Recognized agent-message records
become canonical message events; the sanitized native record remains under the
`openai.codex/v1` extension namespace. Other records are retained as
confidential base64 `raw.output` events. Malformed or truncated records produce
both lossless raw output and a content-free operational diagnostic.

`Config.Executable` selects the binary (default `codex`) and `Config.Model`
adds `--model`. NUL and line breaks are rejected. Non-zero exits map to the
stable `PROCESS_FAILED` result; operator-facing errors contain an exit code but
never provider output or secrets.
