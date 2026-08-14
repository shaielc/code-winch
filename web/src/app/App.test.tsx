import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";

class FakeSocket {
  static instances: FakeSocket[] = [];
  onopen?: () => void;
  onmessage?: (event: { data: string }) => void;
  onerror?: () => void;
  onclose?: () => void;
  constructor(readonly url: string) {
    FakeSocket.instances.push(this);
  }
  close() {
    /* deterministic fake */
  }
  emit(value: unknown) {
    this.onmessage?.({ data: JSON.stringify(value) });
  }
}

const run = {
  id: "01J00000000000000000000000",
  state: "running",
  version: 1,
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
  workspacePath: "/work",
  harnessProfile: "fake-terminal",
  sandboxProfile: "local-trusted",
  lastSequence: 2,
};
const event = (sequence: number, text: string) => ({
  eventId: `event-${sequence}`,
  runId: run.id,
  sequence,
  occurredAt: run.createdAt,
  kind: "terminal.output",
  schemaVersion: 1,
  source: {},
  sensitivity: "user-content",
  payload: { text },
});

describe("terminal run browser slice", () => {
  beforeEach(() => {
    history.replaceState(null, "", `/?run=${run.id}`);
    sessionStorage.clear();
    FakeSocket.instances = [];
    vi.stubGlobal("WebSocket", FakeSocket);
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: string | URL | Request, init?: RequestInit) => {
        const url = String(input);
        if (url.includes("/events?"))
          return new Response(
            JSON.stringify({
              events: [event(1, "hello "), event(2, "world")],
              nextAfterSequence: 2,
              hasMore: false,
            }),
            { status: 200 },
          );
        if (init?.method === "POST")
          return new Response(JSON.stringify({ accepted: true }), {
            status: 202,
          });
        return new Response(JSON.stringify(run), {
          status: 200,
          headers: { etag: '"1"' },
        });
      }),
    );
  });
  afterEach(() => vi.unstubAllGlobals());

  it("loads, reconnects without duplicates, sends input, and stops", async () => {
    const first = render(<App />);
    expect(await screen.findByText("hello world")).toBeInTheDocument();
    expect(FakeSocket.instances[0].url).toContain("after_sequence=2");
    FakeSocket.instances[0].emit({
      type: "event",
      event: event(2, "duplicate"),
    });
    expect(screen.queryByText(/duplicate/)).not.toBeInTheDocument();
    first.unmount();
    render(<App />);
    expect(await screen.findByText("hello world")).toBeInTheDocument();
    await waitFor(() =>
      expect(FakeSocket.instances[1].url).toContain("after_sequence=2"),
    );
    fireEvent.change(screen.getByLabelText("Send input"), {
      target: { value: "continue" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Send" }));
    await waitFor(() =>
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining("/input"),
        expect.objectContaining({ method: "POST" }),
      ),
    );
    fireEvent.click(screen.getByRole("button", { name: "Stop run" }));
    await waitFor(() =>
      expect(fetch).toHaveBeenCalledWith(
        expect.stringContaining("/stop"),
        expect.objectContaining({ method: "POST" }),
      ),
    );
    expect(screen.getByText("Local trusted execution")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Resize unavailable" }),
    ).toBeDisabled();
  });
});
