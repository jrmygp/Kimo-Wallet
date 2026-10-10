import * as yup from "yup";

// Shared by any response that enriches a user with their own wallet
// (features/auth/schemas/user-response.schema.ts's login response, and
// GET /v1/user/{kimoId} when the caller looks up their own profile) — kept
// in one place so the two can't drift.
//
// Nullable, not just optional: wallet enrichment is always best-effort
// (apps/api-gateway's fetchWallet never fails the request), and a brand
// new user can legitimately have no wallet yet (async Kafka
// provisioning) — the backend sends a literal JSON `null` for that case.
export const walletSchema = yup
  .object({
    balance: yup.number().required(),
    currency: yup.string().required(),
    status: yup.string().required(),
  })
  .nullable()
  .defined();

export type WalletBalance = yup.InferType<typeof walletSchema>;
