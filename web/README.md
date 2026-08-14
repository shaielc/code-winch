# Code Winch web workspace

The web application requires Node.js 20 or newer. Direct dependency versions
are fixed in `package.json`.

```sh
npm install
npm run format:check
npm run lint
npm run typecheck
npm test
npm run build
npm run api:check
```

Open a terminal run with `?run=<run-id>`. The page fetches durable events before
opening its resumable WebSocket and caches the terminal projection and cursor in
session storage. Terminal output remains text: only a small ANSI color subset is
interpreted, OSC URLs and unsupported controls are removed, and CSP prohibits
inline script and object execution.

API types in
`src/api/schema.ts` are generated from `../api/openapi/code-winch.yaml`; update
the contract first, then run `npm run api:generate` and commit the result.
