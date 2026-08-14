import { useState } from "react";
import { RunPage } from "../features/runs/RunPage";
import "./app.css";

export function App() {
  const [runId, setRunId] = useState(() =>
    new URLSearchParams(location.search).get("run"),
  );
  if (runId)
    return (
      <RunPage
        runId={runId}
        onBack={() => {
          history.pushState(null, "", location.pathname);
          setRunId(null);
        }}
      />
    );
  return (
    <main className="shell">
      <header>
        <div>
          <p className="eyebrow">Code Winch</p>
          <h1>Runs</h1>
        </div>
      </header>
      <section className="empty">
        <h2>Open a terminal run</h2>
        <p>Enter a run ID to reconnect to its durable output.</p>
        <form
          onSubmit={(event) => {
            event.preventDefault();
            const form = new FormData(event.currentTarget);
            const id = String(form.get("runId") ?? "").trim();
            if (id) {
              history.pushState(null, "", `?run=${encodeURIComponent(id)}`);
              setRunId(id);
            }
          }}
        >
          <label htmlFor="run-id">Run ID</label>
          <input
            id="run-id"
            name="runId"
            required
            pattern="[0-9A-HJKMNP-TV-Z]{26}"
          />
          <button>Open run</button>
        </form>
      </section>
    </main>
  );
}
