import { expect, test } from "vitest";
import { getBasePath } from "./basePath";

test.each([
  ["http://sandbox:8080/app", ""],
  ["http://sandbox:8080/app/", ""],
  ["http://sandbox:8080/", ""],
  ["http://sandbox:8080/main/app", "/main"],
  ["https://host/proxy/test/app/", "/proxy/test"],
  ["https://host/app/app", "/app"],
  ["https://host/main/app?x=1#y", "/main"],
  ["https://host/main/application", ""],
])("%s has base path %j", (href, expected) => {
  expect(getBasePath(href)).toBe(expected);
});
