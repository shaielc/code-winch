import { afterEach, expect, test, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { App } from "./App";

afterEach(cleanup);
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

test("shows a submitted message as local without an ordinal", () => {
  globalThis.fetch = vi.fn().mockImplementation(() => new Promise(() => {}));
  const { container } = render(<App />);
  const composer = screen.getByRole("textbox", { name: "Message" });

  fireEvent.change(composer, { target: { value: "a local message" } });
  fireEvent.keyDown(composer, { key: "Enter" });

  const region = screen.getByRole("region", { name: "Local messages" });
  expect(region).toHaveTextContent("Local");
  expect(region).toHaveTextContent("a local message");
  expect(container).not.toHaveTextContent(/ordinal/i);
});
