# Fake sandbox

The fake sandbox provides an in-memory lifecycle and an attached echo pipe. It
is deterministic and useful for contract and runner tests, but it does **not**
prove operating-system process behavior, PTY semantics, isolation, resource
limits, or descendant cleanup. Use the local driver tests for those guarantees.
