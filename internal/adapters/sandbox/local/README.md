# Local PTY sandbox

The local driver starts trusted commands in a host PTY and owns their process
groups. It reports `Isolation: unisolated`: it provides **no filesystem or
network isolation**, resource limits, or secret boundary. Deployments must only
enable it for trusted local development.

`Stop` sends `SIGTERM`, waits for the configured grace period, then sends
`SIGKILL` to the complete process group. `Cleanup` repeats that escalation and
is safe to retry. Operational errors contain stable codes and resource IDs but
never command arguments, environment values, PTY output, or secret content.
