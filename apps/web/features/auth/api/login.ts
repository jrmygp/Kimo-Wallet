import { AxiosError } from "axios";
import { AxiosInstance } from "@/lib/axios";
import { apiEnvelopeSchema } from "@/lib/api-envelope";
import { userResponseSchema, type UserResponseData } from "@/features/auth/schemas/user-response.schema";

const loginEnvelopeSchema = apiEnvelopeSchema(userResponseSchema);

/**
 * Calls POST /v1/auth/login on the API Gateway. phoneNumber must already be
 * a full E.164 string (e.g. "+6281234567890") — combining a country code
 * with a local number is the caller's job (the form's), not this function's.
 *
 * Throws a plain Error with a safe, user-facing message on any failure —
 * either the gateway's own message (validation error, not-found — see
 * apps/api-gateway/internal/handler/respond.go, these are curated to be
 * safe to show a user) or a generic connectivity message if the gateway
 * never responded at all.
 */
export async function loginUser(phoneNumber: string): Promise<UserResponseData> {
  try {
    const raw: unknown = await AxiosInstance.post("/v1/auth/login", { phoneNumber });
    const envelope = await loginEnvelopeSchema.validate(raw);
    return envelope.data;
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
