"use client";

import { useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useAppDispatch, useAppSelector } from "@/lib/store/hooks";
import { clearUser } from "@/features/auth/store/user-slice";
import { decodeJwtExpiry } from "@/lib/jwt";

// Routes reachable without being logged in. AuthWatcher (which calls this
// hook) is mounted once at the root layout, so it runs on these too — a
// missing user here is expected, not something to redirect away from.
const PUBLIC_PATHS = new Set(["/", "/auth/login", "/auth/register"]);

/**
 * Two jobs, both ending in the same redirect-to-login:
 *
 * 1. Proactively logs the user out the moment their access token's own
 *    `exp` claim says it's expired — a UX nicety layered on top of the
 *    actually authoritative check (lib/axios.ts's response interceptor
 *    reacting to a real 401 from the gateway). Without this, an idle user
 *    on a page that makes no requests would just sit on a stale-looking
 *    screen until their next API call finally 401s.
 * 2. Kicks an unauthenticated visitor off any route that isn't in
 *    PUBLIC_PATHS — this app has no middleware-based route protection
 *    (see apps/web/AGENTS.md: this Next.js version's APIs shouldn't be
 *    assumed from memory, and no middleware.ts exists here), so this
 *    global hook is the only thing enforcing it.
 *
 * A JS timer can't itself be persisted across a reload — only a target
 * timestamp can, and the token already carries one (its `exp` claim), so
 * there's nothing extra to store. This re-arms whenever the logged-in user
 * or the route changes (i.e. right after login sets it, or on client-side
 * navigation), since mounting once at the root layout means the effect
 * otherwise never re-runs on a client-side navigation from /auth/login to
 * /home.
 */
export function useAutoLogout() {
  const router = useRouter();
  const pathname = usePathname();
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.user.user);
  // redux-provider.tsx's PersistGate uses loading={null} — it renders
  // children immediately rather than blocking on rehydration — so `user`
  // starts out null on every fresh page load, even for an already-logged-in
  // visitor, until redux-persist finishes reading it back from localStorage
  // a moment later. Without waiting for this, job 2 above would
  // flash-redirect every logged-in user on every refresh.
  const rehydrated = useAppSelector((state) => state._persist?.rehydrated ?? false);

  useEffect(() => {
    if (!rehydrated) return;

    const logout = () => {
      localStorage.removeItem("token");
      dispatch(clearUser());
      router.push("/auth/login");
    };

    const isPublicRoute = PUBLIC_PATHS.has(pathname);

    if (!user?.id) {
      if (!isPublicRoute) logout();
      return;
    }

    const token = localStorage.getItem("token");
    if (!token) {
      if (!isPublicRoute) logout();
      return;
    }

    const exp = decodeJwtExpiry(token);
    if (exp === null) return;

    const msUntilExpiry = exp * 1000 - Date.now();
    if (msUntilExpiry <= 0) {
      logout();
      return;
    }

    const timeoutId = setTimeout(logout, msUntilExpiry);
    return () => clearTimeout(timeoutId);
  }, [user, pathname, rehydrated, dispatch, router]);
}
