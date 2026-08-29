import * as yup from "yup";
import { AxiosError } from "axios";
import { AxiosInstance } from "@/lib/axios";
import { apiEnvelopeSchema } from "@/lib/api-envelope";
import { userSchema, type WalletUser } from "@/features/wallet/schemas/user.schema";

const searchUserEnvelopeSchema = apiEnvelopeSchema(
  yup.object({
    user: userSchema.required(),
  }),
);

/**
 * Calls GET /v1/user/{id} on the API Gateway (auth-required — the request
 * interceptor in lib/axios.ts attaches the bearer token). This is an exact
 * lookup by id ("KimoID"), not a fuzzy/partial search — a miss surfaces as
 * a 404, translated below into the gateway's "user not found" message.
 *
 * Throws a plain Error with a safe, user-facing message on any failure —
 * same convention as loginUser/registerUser (see features/auth/api/).
 */
export async function searchUserById(id: string): Promise<WalletUser> {
  try {
    const raw: unknown = await AxiosInstance.get(`/v1/user/${encodeURIComponent(id)}`);
    const envelope = await searchUserEnvelopeSchema.validate(raw);
    return envelope.data.user;
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
