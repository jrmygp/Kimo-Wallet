/**
 * Decodes (never verifies) a JWT's payload to read its `exp` claim, purely
 * for client-side UX timing (see features/auth/hooks/use-auto-logout.ts).
 * The server is the only authority on whether a token is actually valid —
 * this never gates an actual request, it only decides when to proactively
 * show the user a logged-out state.
 */
function base64UrlDecode(segment: string): string {
  const base64 = segment.replace(/-/g, "+").replace(/_/g, "/");
  const padded = base64 + "=".repeat((4 - (base64.length % 4)) % 4);
  return atob(padded);
}

/**
 * Returns the token's `exp` claim (seconds since epoch), or null if the
 * token isn't a well-formed JWT or has no numeric `exp`.
 */
export function decodeJwtExpiry(token: string): number | null {
  const parts = token.split(".");
  if (parts.length !== 3) return null;

  try {
    const payload: unknown = JSON.parse(base64UrlDecode(parts[1]));
    if (typeof payload === "object" && payload !== null && "exp" in payload) {
      const exp = (payload as { exp: unknown }).exp;
      if (typeof exp === "number") return exp;
    }
    return null;
  } catch {
    return null;
  }
}
