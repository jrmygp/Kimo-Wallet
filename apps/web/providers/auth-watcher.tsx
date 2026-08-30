"use client";

import { useAutoLogout } from "@/features/auth/hooks/use-auto-logout";

/**
 * Mounted once at the root layout so it's active on every page — see
 * useAutoLogout's doc comment for what it actually does. Renders nothing.
 */
export function AuthWatcher() {
  useAutoLogout();
  return null;
}
