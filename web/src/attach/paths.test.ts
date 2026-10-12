import { expect, test } from "vitest";
import { sandboxURL } from "./paths";

const origin = "https://host.test";

test.each([
  ["/", "/api/session"],
  ["/index.html", "/api/session"],
  ["/main/sandbox/", "/main/sandbox/api/session"],
  ["/main/sandbox", "/main/sandbox/api/session"],
  ["/outer/inner/sandbox/", "/outer/inner/sandbox/api/session"],
  ["/main/sandbox/some/route", "/main/sandbox/api/session"],
])("resolves api/session from %s", (pathname, want) => {
  expect(sandboxURL("api/session", { origin, pathname })).toBe(origin + want);
});

test("treats a leading slash as relative to the sandbox base", () => {
  expect(
    sandboxURL("/api/session", { origin, pathname: "/main/sandbox/" }),
  ).toBe(`${origin}/main/sandbox/api/session`);
});
