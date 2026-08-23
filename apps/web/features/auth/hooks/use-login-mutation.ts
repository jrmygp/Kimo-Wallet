import { useMutation } from "@tanstack/react-query";
import { loginUser } from "@/features/auth/api/login";

/**
 * Wraps loginUser as a TanStack Query mutation. Login is a mutation, not a
 * query — it has a side effect (issues a session token) and isn't safe to
 * retry/cache like a GET — so this uses useMutation, not useQuery. A future
 * read-only endpoint (e.g. fetching the user's own profile) should use
 * useQuery instead, following the same one-hook-per-API-call pattern.
 */
export function useLoginMutation() {
  return useMutation({
    mutationFn: loginUser,
  });
}
