import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import { RecordList } from "./RecordList";

test("renders raw and terminal records", () => {
  render(
    <RecordList
      records={[
        {
          ordinal: 1,
          kind: "stream.raw",
          occurredAt: "2026-01-01T00:00:00Z",
          sensitivity: "user-content",
          payload: { stream: "stdout", encoding: "utf-8", data: "hello" },
        },
        {
          ordinal: 2,
          kind: "session.terminated",
          occurredAt: "2026-01-01T00:00:01Z",
          sensitivity: "operational",
          payload: { outcome: "failed", exitCode: 1 },
        },
      ]}
    />,
  );
  expect(screen.getByText("hello")).toBeInTheDocument();
  expect(screen.getByText("Harness failed (1)")).toBeInTheDocument();
});
