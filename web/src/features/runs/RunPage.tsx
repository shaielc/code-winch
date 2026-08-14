import { FormEvent, useCallback, useEffect, useRef, useState } from "react";

import {
  ApiError,
  Run,
  RunEvent,
  runClient,
  streamUrl,
} from "../../api/runClient";
import { Terminal } from "../../renderers/Terminal";

const terminalKinds = new Set([
  "terminal.output",
  "stream.stdout",
  "stream.stderr",
  "output",
]);
const cacheKey = (id: string) => `code-winch.run.${id}.projection.v1`;

interface Projection {
  sequence: number;
  output: string;
}

function payloadText(event: RunEvent): string {
  if (!terminalKinds.has(event.kind)) return "";
  const value = event.payload.text ?? event.payload.data ?? event.payload.bytes;
  return typeof value === "string" ? value : "";
}

function readProjection(id: string): Projection {
  try {
    const value = JSON.parse(
      sessionStorage.getItem(cacheKey(id)) ?? "null",
    ) as Partial<Projection> | null;
    if (
      value &&
      Number.isSafeInteger(value.sequence) &&
      typeof value.output === "string"
    ) {
      return { sequence: value.sequence!, output: value.output };
    }
  } catch {
    /* Invalid browser state is rebuilt from durable events. */
  }
  return { sequence: 0, output: "" };
}

export function RunPage({
  runId,
  onBack,
}: {
  runId: string;
  onBack: () => void;
}) {
  const initial = useRef(readProjection(runId));
  const [run, setRun] = useState<Run>();
  const [etag, setEtag] = useState("");
  const [output, setOutput] = useState(initial.current.output);
  const outputRef = useRef(initial.current.output);
  const [sequence, setSequence] = useState(initial.current.sequence);
  const cursor = useRef(initial.current.sequence);
  const [connection, setConnection] = useState("Connecting");
  const [error, setError] = useState("");
  const [input, setInput] = useState("");

  const apply = useCallback(
    (events: RunEvent[]) => {
      const ordered = [...events].sort((a, b) => a.sequence - b.sequence);
      for (const event of ordered) {
        if (event.sequence <= cursor.current) continue;
        if (event.sequence !== cursor.current + 1)
          throw new Error(
            "Event history has a sequence gap; reconnect to repair it.",
          );
        outputRef.current += payloadText(event);
        cursor.current = event.sequence;
      }
      setOutput(outputRef.current);
      sessionStorage.setItem(
        cacheKey(runId),
        JSON.stringify({
          sequence: cursor.current,
          output: outputRef.current,
        }),
      );
      setSequence(cursor.current);
    },
    [runId],
  );

  useEffect(() => {
    let active = true;
    let socket: WebSocket | undefined;
    let retry: number | undefined;
    const connect = async () => {
      try {
        const detail = await runClient.get(runId);
        if (!active) return;
        setRun(detail.body);
        setEtag(detail.etag ?? `"${detail.body.version}"`);
        let page;
        do {
          page = (await runClient.events(runId, cursor.current)).body;
          apply(page.events);
        } while (page.hasMore && active);
        if (!active) return;
        socket = new WebSocket(streamUrl(runId, cursor.current));
        socket.onopen = () => setConnection("Live");
        socket.onmessage = (message) => {
          const frame = JSON.parse(String(message.data)) as {
            type: string;
            event?: RunEvent;
            lastSequence?: number;
          };
          if (frame.type === "event" && frame.event) apply([frame.event]);
          if (frame.type === "caught_up") setConnection("Live");
        };
        socket.onerror = () => setConnection("Reconnecting");
        socket.onclose = () => {
          if (active) {
            setConnection("Reconnecting");
            retry = window.setTimeout(connect, 500);
          }
        };
      } catch (reason) {
        if (!active) return;
        const message =
          reason instanceof ApiError
            ? `${reason.message} (${reason.code}${reason.requestId ? `, request ${reason.requestId}` : ""})`
            : reason instanceof Error
              ? reason.message
              : "Unable to load the run.";
        setError(message);
        setConnection("Disconnected");
      }
    };
    void connect();
    return () => {
      active = false;
      socket?.close();
      if (retry) clearTimeout(retry);
    };
  }, [apply, runId]);

  const running = run?.state === "running";
  const send = async (event: FormEvent) => {
    event.preventDefault();
    if (!input.trim()) return;
    try {
      await runClient.command(runId, etag, "text", { text: input });
      setInput("");
    } catch (reason) {
      setError(
        reason instanceof Error ? reason.message : "Input was not accepted.",
      );
    }
  };
  const stop = async () => {
    try {
      await runClient.command(runId, etag, "stop", {
        reason: "Stopped from web terminal",
      });
    } catch (reason) {
      setError(
        reason instanceof Error ? reason.message : "Stop was not accepted.",
      );
    }
  };

  return (
    <main className="shell">
      <button className="link" onClick={onBack}>
        ← Runs
      </button>
      <header>
        <div>
          <p className="eyebrow">Terminal run</p>
          <h1>{runId}</h1>
        </div>
        <span className={`connection ${connection.toLowerCase()}`}>
          {connection}
        </span>
      </header>
      <aside className="warning" role="alert">
        <strong>Local trusted execution</strong> The{" "}
        <code>{run?.sandboxProfile ?? "local-trusted"}</code> sandbox does not
        isolate this run from your host.
      </aside>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      <section className="metadata" aria-label="Run capabilities">
        <div>
          <span>Status</span>
          <strong>{run?.state ?? "loading"}</strong>
        </div>
        <div>
          <span>Harness</span>
          <strong>{run?.harnessProfile ?? "—"}</strong>
        </div>
        <div>
          <span>Sandbox</span>
          <strong>{run?.sandboxProfile ?? "—"}</strong>
        </div>
        <div>
          <span>Cursor</span>
          <strong>sequence {sequence}</strong>
        </div>
      </section>
      <Terminal output={output} />
      <form className="controls" onSubmit={send}>
        <label htmlFor="terminal-input">Send input</label>
        <div>
          <input
            id="terminal-input"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            disabled={!running}
            placeholder={
              running ? "Type a message" : "Input requires a running run"
            }
          />
          <button disabled={!running || !input.trim()}>Send</button>
        </div>
        <div className="actions">
          <button
            type="button"
            disabled
            title="This harness does not advertise terminal resize"
          >
            Resize unavailable
          </button>
          <button
            className="danger"
            type="button"
            disabled={!running}
            onClick={stop}
          >
            Stop run
          </button>
        </div>
        {!running && (
          <p className="hint">
            Input and stop are unavailable while the run is{" "}
            {run?.state ?? "loading"}.
          </p>
        )}
        <p className="hint">
          Terminal resize is disabled because capability metadata is not
          advertised by this harness.
        </p>
      </form>
    </main>
  );
}
