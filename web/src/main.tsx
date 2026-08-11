import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import { App } from "./app/App";

const root = document.getElementById("root");

if (root === null) {
  throw new Error('Unable to start Code Winch: required element "#root" was not found.');
}

createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
