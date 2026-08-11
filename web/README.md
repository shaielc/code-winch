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

The application shell deliberately contains no product screens. API types in
`src/api/schema.ts` are generated from `../api/openapi/code-winch.yaml`; update
the contract first, then run `npm run api:generate` and commit the result.
