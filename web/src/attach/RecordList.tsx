import type { SessionRecord } from "./useSessionStream";

function recordText(record: SessionRecord) {
  if (record.kind === "stream.raw") {
    const data = String(record.payload.data ?? "");
    return record.payload.encoding === "base64" ? `[base64] ${data}` : data;
  }
  if (record.kind === "session.terminated") {
    const details = record.payload.signal ?? record.payload.exitCode;
    return `Harness ${String(record.payload.outcome)}${details === undefined ? "" : ` (${String(details)})`}`;
  }
  return JSON.stringify(record.payload);
}

export function RecordList({ records }: { records: SessionRecord[] }) {
  return (
    <section aria-labelledby="records-title">
      <h2 id="records-title">Live harness output</h2>
      <ol className="records" aria-live="polite">
        {records.map((record) => (
          <li key={record.ordinal}>
            <span className="ordinal">{record.ordinal}</span>
            <pre>{recordText(record)}</pre>
          </li>
        ))}
      </ol>
      {records.length === 0 && <p>Waiting for the harness…</p>}
    </section>
  );
}
