type PageLocation = Pick<Location, "origin" | "pathname">;

// The page is served at {origin}/ or, behind proxies, at {origin}/{proxy/}+sandbox/.
export function sandboxURL(
  path: string,
  location: PageLocation = window.location,
): string {
  const segments = location.pathname.split("/");
  const sandbox = segments.indexOf("sandbox");
  const base =
    sandbox === -1 ? "/" : `${segments.slice(0, sandbox + 1).join("/")}/`;
  return new URL(path.replace(/^\/+/, ""), location.origin + base).toString();
}
