import { useQuery } from "@tanstack/react-query";
import { searchUserById } from "@/features/wallet/api/search-user";

/**
 * Looks up a user by KimoID via GET /v1/user/{id}. Read-only (GetUserByID
 * has no side effect) — this uses useQuery, not useMutation, per the
 * read-vs-write split noted in useLoginMutation's doc comment.
 *
 * `id` is expected to be the *submitted* id, not the raw input on every
 * keystroke — the caller only updates it on Enter (see app/home/page.tsx),
 * which is what makes this fire once per submission instead of per
 * keystroke. `enabled` gates the very first render before anything has
 * been submitted. `retry: false` because a miss here is an expected "not
 * found" outcome the user should see immediately, not something to retry.
 */
export function useSearchUserQuery(id: string) {
  return useQuery({
    queryKey: ["user", "search", id],
    queryFn: () => searchUserById(id),
    enabled: id.length > 0,
    retry: false,
  });
}
