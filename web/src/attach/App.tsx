import { useEffect, useState } from "react";
import { getBasePath } from "./basePath";
import { Posture, type SessionPosture } from "./Posture";
import { RecordList } from "./RecordList";
import { useSessionStream } from "./useSessionStream";
export function App() {
  const [posture, setPosture] = useState<SessionPosture>();
  const [error, setError] = useState(false);
  const { records, streamError } = useSessionStream();
  useEffect(() => {
    fetch(`${getBasePath()}/api/session`)
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
      {streamError && <p role="alert">Harness output is unavailable.</p>}
      <RecordList records={records} />
    </main>
  );
}
