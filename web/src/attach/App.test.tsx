import { afterEach, expect, test, vi } from "vitest";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { App } from "./App";

afterEach(cleanup);
test("shows unknown child status independently of posture", async () => {
  globalThis.fetch = vi.fn().mockImplementation(() => new Promise(() => {}));
  render(<App />);

  let region = screen.getByRole("region", { name: "Child status" });
  expect(region).toHaveTextContent(/^Child statusUnknown$/);
  expect(screen.getByText("Unknown", { selector: "p" })).toBeInTheDocument();
  expect(region).not.toHaveTextContent(/Down|Running|ready|exit/i);
  expect(
    screen.getByRole("region", { name: "Local messages" }),
  ).toBeInTheDocument();

  cleanup();
  globalThis.fetch = vi.fn().mockRejectedValue(new Error("request failed"));
  render(<App />);

  expect(await screen.findByRole("alert")).toHaveTextContent(
    "Sandbox posture is unavailable.",
  );
  region = screen.getByRole("region", { name: "Child status" });
  expect(region).toHaveTextContent(/^Child statusUnknown$/);
  expect(screen.getByText("Unknown", { selector: "p" })).toBeInTheDocument();
  expect(region).not.toHaveTextContent(/Down|Running|ready|exit/i);
});

test("shows effective and unenforced posture", async () => {
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

test.each(["Down", "Running", "Exited"])(
  "shows %s only after the sandbox reports the exact supported status",
  async (childStatus) => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        profile: "container-standard",
        unenforcedControls: ["network-egress"],
        childStatus,
      }),
    });
    render(<App />);

    const region = screen.getByRole("region", { name: "Child status" });
    expect(region).toHaveTextContent(/^Child statusUnknown$/);
    expect(
      await screen.findByText(childStatus, { selector: "p" }),
    ).toBeInTheDocument();
    expect(region).toHaveTextContent(`Child status${childStatus}`);
    expect(region).not.toHaveTextContent(/ready|fake/i);
  },
);

test.each([
  ["missing", undefined],
  ["wrong case", "down"],
  ["unsupported", "Started"],
])("keeps child status Unknown for a %s status", async (_name, childStatus) => {
  globalThis.fetch = vi.fn().mockResolvedValue({
    ok: true,
    json: async () => ({
      profile: "container-standard",
      unenforcedControls: ["network-egress"],
      ...(childStatus === undefined ? {} : { childStatus }),
    }),
  });
  render(<App />);

  expect(await screen.findByText("container-standard")).toBeInTheDocument();
  const region = screen.getByRole("region", { name: "Child status" });
  expect(region).toHaveTextContent(/^Child statusUnknown$/);
  expect(region).not.toHaveTextContent(/Down|Running|ready|fake|exit/i);
});

test.each([
  [
    "invalid JSON",
    { ok: true, json: async () => Promise.reject(new Error("invalid JSON")) },
  ],
  [
    "rejected response",
    { ok: false, json: async () => ({ childStatus: "Down" }) },
  ],
])(
  "request failure from %s leaves child status Unknown",
  async (_name, response) => {
    globalThis.fetch = vi.fn().mockResolvedValue(response);
    render(<App />);

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Sandbox posture is unavailable.",
    );
    const region = screen.getByRole("region", { name: "Child status" });
    expect(region).toHaveTextContent(/^Child statusUnknown$/);
    expect(region).not.toHaveTextContent(/Down|Running|ready|fake|exit/i);
  },
);

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
