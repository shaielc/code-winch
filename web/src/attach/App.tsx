import { useEffect, useState } from "react";
import { Posture, type SessionPosture } from "./Posture";
export function App() {
  const [posture, setPosture] = useState<SessionPosture>();
  const [error, setError] = useState(false);
  useEffect(() => {
    fetch("/api/session")
      .then((response) => {
        if (!response.ok) throw new Error("request failed");
        return response.json() as Promise<SessionPosture>;
      })
      .then(setPosture)
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
    </main>
  );
}
