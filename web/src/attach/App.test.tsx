import { expect, test, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { App } from "./App";
test("shows effective and unenforced posture", async () => {
  vi.stubGlobal(
    "WebSocket",
    vi.fn(() => ({ close: vi.fn() })),
  );
  globalThis.fetch = vi.fn().mockResolvedValue({
    ok: true,
    json: async () => ({
      profile: "container-standard",
      unenforcedControls: ["network-egress"],
    }),
  });
  render(<App />);
  expect(await screen.findByText("container-standard")).toBeInTheDocument();
  expect(screen.getByText("network-egress")).toBeInTheDocument();
  expect(screen.getByText(/not enforced by this profile/i)).toBeInTheDocument();
});
