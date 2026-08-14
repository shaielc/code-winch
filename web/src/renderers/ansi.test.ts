import { describe, expect, it } from "vitest";
import { parseAnsi } from "./ansi";

describe("parseAnsi", () => {
  it("projects supported colors without producing markup", () => {
    expect(parseAnsi("plain \u001b[31mred\u001b[0m end")).toEqual([
      { text: "plain ", className: undefined },
      { text: "red", className: "ansi-red" },
      { text: " end", className: undefined },
    ]);
  });
  it("removes malicious OSC links and controls while leaving markup inert", () => {
    const segments = parseAnsi(
      "\u001b]8;;javascript:alert(1)\u0007click\u001b]8;;\u0007<script>alert(2)</script>\u001b[2J",
    );
    expect(segments.map(({ text }) => text).join("")).toBe(
      "click<script>alert(2)</script>",
    );
  });
});
