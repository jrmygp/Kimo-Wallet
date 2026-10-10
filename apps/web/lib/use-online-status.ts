"use client";

import { useSyncExternalStore } from "react";

function subscribe(onChange: () => void) {
  window.addEventListener("online", onChange);
  window.addEventListener("offline", onChange);
  return () => {
    window.removeEventListener("online", onChange);
    window.removeEventListener("offline", onChange);
  };
}

/**
 * Whether the browser reports a network connection. `navigator.onLine` can
 * say "online" on a captive or dead network, so treat `true` as "maybe" — but
 * `false` is reliable, which is what blocking money actions offline needs
 * (docs/CLAUDE.md §3.7.6). Assumes online during server rendering.
 */
export function useOnlineStatus() {
  return useSyncExternalStore(
    subscribe,
    () => navigator.onLine,
    () => true,
  );
}
