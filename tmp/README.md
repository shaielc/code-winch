# Interim API scripts — delete under P1-051

**This directory should not exist.** It is an unowned surface that duplicates
`cmd/winch`, kept only until the operator CLI can drive the run API. P1-051
deletes it; that is an acceptance criterion of the brief, not a suggestion.

## Why it is here

P1-050 must show that "published events reach a subscribed WebSocket in
sequence order", and nothing in this repository can subscribe to that
WebSocket: the client is `winch run watch`, which P1-051 builds *after*
P1-050. The criterion has no witness inside the task that carries it. See
[the post-mortem](../docs/workplan/post-mortems/2026-08-17-unverifiable-stream-criterion.md).

Rather than mark P1-050 complete on evidence that cannot be reproduced, the
evidence lives here, in the shape the CLI will take, so each script is replaced
one-for-one and the deletion is mechanical.

| Script | Replaced by |
|---|---|
| `create.sh` | `winch run create` |
| `get.sh` | `winch run get` |
| `start.sh` | `winch run start` |
| `input.sh` | `winch run input` |
| `stop.sh` | `winch run stop` |
| `events.sh` | `winch run watch --after-sequence` |
| `stream.py` | `winch run watch` |
| `common.sh` | `internal/platform/config` and the CLI's shared client |

## Using them

They read `WINCH_ENDPOINT`, `WINCH_TOKEN`, `WINCH_CSRF_TOKEN`, and
`WINCH_ALLOWED_ORIGIN`, defaulting to the development values in
`deployments/compose.yml`. They need `curl`, `jq`, and Python 3. Start the
stack first:

```sh
docker compose -f deployments/compose.yml up --build -d
```

Each script takes the run ID as an argument and otherwise reuses the last run
`create.sh` made, recorded in `tmp/.last-run`:

```sh
./tmp/create.sh                       # → run ID, state "created"
./tmp/stream.py "$(cat tmp/.last-run)" 30 &   # watch it live
./tmp/start.sh                        # → state "running", a harness exists
./tmp/input.sh 'echo hello'           # → "hello" appears in the stream
./tmp/stop.sh                         # → state "stopping", then "completed"
./tmp/events.sh                       # the persisted history it all came from
```

`stream.py` speaks just enough of RFC 6455 to subscribe and print. It has no
dependencies and is not a client library: no continuation frames, no
compression, no reconnect. `winch run watch` owns those properly, including
resume from `after_sequence`.

## What they are not

Not tested, not CI-wired, and not a supported interface. Nothing in the
repository may depend on them — no test, no Makefile target, no documented
procedure outside this file and the post-mortem.
