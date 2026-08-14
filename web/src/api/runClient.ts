import type { components } from "./schema";

export type Run = components["schemas"]["Run"];
export type RunEvent = components["schemas"]["Event"];
export type EventPage = components["schemas"]["EventPage"];
export type InputKind = components["schemas"]["RunInputRequest"]["kind"];

export interface StreamMessage {
  type: "event" | "caught_up" | "heartbeat" | "disconnect";
  event?: RunEvent;
  lastSequence?: number;
}

export class ApiError extends Error {
  constructor(
    message: string,
    readonly code: string,
    readonly requestId?: string,
  ) {
    super(message);
  }
}

async function request<T>(
  path: string,
  init?: RequestInit,
): Promise<{ body: T; etag: string | null }> {
  const response = await fetch(`/api/v1${path}`, init);
  if (!response.ok) {
    const problem = (await response.json().catch(() => ({}))) as Partial<
      components["schemas"]["Problem"]
    >;
    throw new ApiError(
      problem.detail ?? `The server returned HTTP ${response.status}.`,
      problem.code ?? "unexpected_response",
      problem.requestId,
    );
  }
  return {
    body: (await response.json()) as T,
    etag: response.headers.get("etag"),
  };
}

export const runClient = {
  get: (id: string) => request<Run>(`/runs/${encodeURIComponent(id)}`),
  events: (id: string, after: number) =>
    request<EventPage>(
      `/runs/${encodeURIComponent(id)}/events?after_sequence=${after}&limit=200`,
    ),
  command: (
    id: string,
    etag: string,
    kind: InputKind | "stop",
    data: Record<string, unknown> = {},
  ) =>
    request<unknown>(
      `/runs/${encodeURIComponent(id)}/${kind === "stop" ? "stop" : "input"}`,
      {
        method: "POST",
        headers: {
          "content-type": "application/json",
          "Idempotency-Key": crypto.randomUUID(),
          "If-Match": etag,
        },
        body: JSON.stringify(kind === "stop" ? data : { kind, ...data }),
      },
    ),
};

export function streamUrl(runId: string, after: number): string {
  const url = new URL(
    `/api/v1/runs/${encodeURIComponent(runId)}/events/stream`,
    window.location.href,
  );
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.searchParams.set("after_sequence", String(after));
  return url.toString();
}
