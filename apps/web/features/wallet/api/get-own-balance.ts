import * as yup from "yup";
import { AxiosError } from "axios";
import { AxiosInstance } from "@/lib/axios";
import { apiEnvelopeSchema } from "@/lib/api-envelope";
import { userSchema } from "@/features/wallet/schemas/user.schema";
import { walletSchema, type WalletBalance } from "@/features/wallet/schemas/wallet.schema";

// Same GET /v1/user/{kimoId} endpoint as searchUserById, but called with
// the caller's OWN kimoId: the gateway only ever includes `wallet` in the
// response when the authenticated caller's id matches the looked-up
// user's (apps/api-gateway's walletForOwnProfile), so this is the one
// call site that can actually read it.
const ownProfileEnvelopeSchema = apiEnvelopeSchema(
  yup.object({
    user: userSchema.required(),
    wallet: walletSchema,
  }),
);

/**
 * Reads the caller's current wallet balance back from the server — the
 * only authoritative source for a displayed balance (docs/CLAUDE.md
 * §3.7.4: never compute it client-side). Used to refresh Redux's cached
 * balance after it goes stale, e.g. once a transfer settles, since
 * nothing else currently re-fetches it after login.
 */
export async function fetchOwnBalance(ownKimoId: string): Promise<WalletBalance> {
  try {
    const raw: unknown = await AxiosInstance.get(`/v1/user/${encodeURIComponent(ownKimoId)}`);
    const envelope = await ownProfileEnvelopeSchema.validate(raw);
    return envelope.data.wallet;
  } catch (error) {
    if (error instanceof AxiosError) {
      const message = error.response?.data?.message;
      if (typeof message === "string" && message.length > 0) {
        throw new Error(message);
      }
      throw new Error("Unable to reach the server. Please check your connection and try again.");
    }
    throw new Error("Something went wrong. Please try again.");
  }
}
