"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAppDispatch, useAppSelector } from "@/lib/store/hooks";
import { clearUser } from "@/features/auth/store/user-slice";
import { decodeJwtExpiry } from "@/lib/jwt";

/**
 * Proactively logs the user out the moment their access token's own `exp`
 * claim says it's expired — a UX nicety layered on top of the actually
 * authoritative check (lib/axios.ts's response interceptor reacting to a
 * real 401 from the gateway). Without this, an idle user on a page that
 * makes no requests would just sit on a stale-looking screen until their
 * next API call finally 401s.
 *
 * A JS timer can't itself be persisted across a reload — only a target
 * timestamp can, and the token already carries one (its `exp` claim), so
 * there's nothing extra to store. This re-arms whenever the logged-in user
 * changes (i.e. right after login sets it), since mounting once at the
 * root layout means the effect otherwise never re-runs on a client-side
 * navigation from /auth/login to /home.
 */
export function useAutoLogout() {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.user.user);

  useEffect(() => {
    if (!user) return;

    const token = localStorage.getItem("token");
    if (!token) return;

    const exp = decodeJwtExpiry(token);
    if (exp === null) return;

    const logout = () => {
      localStorage.removeItem("token");
      dispatch(clearUser());
      router.push("/auth/login");
    };

    const msUntilExpiry = exp * 1000 - Date.now();
    if (msUntilExpiry <= 0) {
      logout();
      return;
    }

    const timeoutId = setTimeout(logout, msUntilExpiry);
    return () => clearTimeout(timeoutId);
  }, [user, dispatch, router]);
}
