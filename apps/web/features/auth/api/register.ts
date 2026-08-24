import { AxiosError } from "axios";
import { AxiosInstance } from "@/lib/axios";
import { apiEnvelopeSchema } from "@/lib/api-envelope";
import { userResponseSchema, type UserResponseData } from "@/features/auth/schemas/user-response.schema";

const registerEnvelopeSchema = apiEnvelopeSchema(userResponseSchema);

export async function registerUser(phoneNumber: string, fullName: string): Promise<UserResponseData> {
  try {
    const raw: unknown = await AxiosInstance.post("/v1/auth/register", { phoneNumber, fullName });
    const envelope = await registerEnvelopeSchema.validate(raw);
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
