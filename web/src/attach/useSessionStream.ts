import { useEffect, useState } from "react";
import { getBasePath } from "./basePath";

export interface SessionRecord {
  ordinal: number;
  kind: string;
  occurredAt: string;
  sensitivity: string;
  payload: Record<string, unknown>;
}

export function useSessionStream() {
  const [records, setRecords] = useState<SessionRecord[]>([]);
  const [streamError, setStreamError] = useState(false);
  useEffect(() => {
    const protocol = location.protocol === "https:" ? "wss:" : "ws:";
    const socket = new WebSocket(
      `${protocol}//${location.host}${getBasePath()}/api/session/stream`,
    );
    socket.onmessage = (event) => {
      const record = JSON.parse(String(event.data)) as SessionRecord;
      setRecords((current) => [...current, record]);
    };
    socket.onerror = () => setStreamError(true);
    return () => socket.close();
  }, []);
  return { records, streamError };
}
