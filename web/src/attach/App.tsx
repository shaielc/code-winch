import { useEffect, useState } from "react";
import { Composer } from "./Composer";
import { sandboxURL } from "./paths";
import { Posture, type SessionPosture } from "./Posture";
export function App() {
  const [posture, setPosture] = useState<SessionPosture>();
  const [childStatus, setChildStatus] = useState("Unknown");
  const [error, setError] = useState(false);
  const [localMessages, setLocalMessages] = useState<string[]>([]);
  useEffect(() => {
    fetch(sandboxURL("api/session"))
      .then((response) => {
        if (!response.ok) throw new Error("request failed");
        return response.json() as Promise<unknown>;
      })
      .then((payload) => {
        if (isSessionPosture(payload)) setPosture(payload);
        if (isObject(payload) && isChildStatus(payload.childStatus)) {
          setChildStatus(payload.childStatus);
        }
      })
      .catch(() => setError(true));
  }, []);
  return (
    <main>
      <header>
        <span className="eyebrow">Code Winch</span>
        <h1>Sandbox attach</h1>
        <p>Direct, host-local access to this sandbox session.</p>
      </header>
      {error && <p role="alert">Sandbox posture is unavailable.</p>}
      {posture && <Posture posture={posture} />}
      <section aria-labelledby="child-status-heading">
        <h2 id="child-status-heading">Child status</h2>
        <p>{childStatus}</p>
      </section>
      <section aria-labelledby="local-messages-heading">
        <h2 id="local-messages-heading">Local messages</h2>
        {localMessages.length === 0 ? (
          <p className="empty-state">Messages you send will appear here.</p>
        ) : (
          <ul className="local-messages">
            {localMessages.map((message, index) => (
              <li key={index}>
                <span className="local-label">Local</span>
                <p>{message}</p>
              </li>
            ))}
          </ul>
        )}
      </section>
      <Composer
        onSubmit={(message) =>
          setLocalMessages((messages) => [...messages, message])
        }
      />
    </main>
  );
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isChildStatus(value: unknown): value is "Down" | "Running" | "Exited" {
  return value === "Down" || value === "Running" || value === "Exited";
}

function isSessionPosture(value: unknown): value is SessionPosture {
  return (
    isObject(value) &&
    typeof value.profile === "string" &&
    Array.isArray(value.unenforcedControls) &&
    value.unenforcedControls.every((control) => typeof control === "string")
  );
}
