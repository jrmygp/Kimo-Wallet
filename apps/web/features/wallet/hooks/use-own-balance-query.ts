import { useQuery } from "@tanstack/react-query";
import { fetchOwnBalance } from "@/features/wallet/api/get-own-balance";

/**
 * Re-fetches the logged-in user's own balance from the server. Read-only,
 * so useQuery, not useMutation. `enabled` gates it until a kimoId is
 * actually known (e.g. before redux-persist has rehydrated the user).
 *
 * This is a safety net, not the primary update path: a successful
 * transfer calls fetchOwnBalance directly and dispatches the result
 * immediately (see app/wallet/transfer/[userId]/page.tsx) rather than
 * waiting for whatever page happens to mount this hook next.
 */
export function useOwnBalanceQuery(kimoId: string | undefined) {
  return useQuery({
    queryKey: ["user", "own-balance", kimoId],
    queryFn: () => fetchOwnBalance(kimoId!),
    enabled: !!kimoId,
  });
}
