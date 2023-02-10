import { RemixBrowser } from "@remix-run/react";
import { startTransition, StrictMode } from "react";
import { hydrate } from "react-dom";

function _hydrate() {
  startTransition(() => {
    hydrate(<StrictMode><RemixBrowser /></StrictMode>, document);
  });
}

if (window.requestIdleCallback) {
  window.requestIdleCallback(_hydrate);
} else {
  // Safari doesn't support requestIdleCallback
  // https://caniuse.com/requestidlecallback
  window.setTimeout(_hydrate, 1);
}
