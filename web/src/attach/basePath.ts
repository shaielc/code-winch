// The page is served at {protocol}://{domain}/{proxy...}/app; the API lives under {proxy...}.
export function getBasePath(href: string = location.href): string {
  const match = /^(.*)\/app\/?$/.exec(new URL(href).pathname);
  return match ? match[1] : "";
}
