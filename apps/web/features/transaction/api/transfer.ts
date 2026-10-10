import { AxiosError } from "axios";
import { AxiosInstance } from "@/lib/axios";
import { apiEnvelopeSchema } from "@/lib/api-envelope";
import {
  transferResponseSchema,
  type TransferRequestData,
  type TransferTransaction,
} from "@/features/transaction/schemas/transfer.schema";

const transferEnvelopeSchema = apiEnvelopeSchema(transferResponseSchema);

// Long enough for a slow mobile connection, short enough that the user isn't
// left staring at a spinner; a timeout is reported as "unconfirmed", never
// as a failure.
const TRANSFER_TIMEOUT_MS = 15_000;

/**
 * - `unconfirmed`: we don't know whether the transfer was created — no
 *   response, a timeout, a 5xx, or a 2xx we couldn't read. Never shown as
 *   "failed" (docs/CLAUDE.md §3.7.5); a retry must reuse the same key.
 * - `duplicate`: the server already has a transfer with this idempotency
 *   key (409). Terminal for this key.
 * - `rejected`: the server refused it (4xx) — nothing was created.
 */
export type TransferErrorKind = "unconfirmed" | "duplicate" | "rejected";

export class TransferError extends Error {
  readonly kind: TransferErrorKind;

  constructor(kind: TransferErrorKind, message: string) {
    super(message);
    this.name = "TransferError";
    this.kind = kind;
  }
}

const UNCONFIRMED_MESSAGE = "We couldn't confirm your transfer yet.";

export async function transferFunds(params: TransferRequestData): Promise<TransferTransaction> {
  let raw: unknown;
  try {
    raw = await AxiosInstance.post("/v1/transaction", params, { timeout: TRANSFER_TIMEOUT_MS });
  } catch (error) {
    if (error instanceof AxiosError && error.response) {
      const status = error.response.status;
      const message = error.response.data?.message;
      const serverMessage = typeof message === "string" && message.length > 0 ? message : undefined;
      if (status === 409) {
        throw new TransferError("duplicate", serverMessage ?? "This transfer was already submitted.");
      }
      if (status >= 400 && status < 500) {
        throw new TransferError("rejected", serverMessage ?? "This transfer couldn't be sent.");
      }
    }
    // No response (offline, timeout, connection reset) or a 5xx: the request
    // may or may not have created the transfer.
    throw new TransferError("unconfirmed", UNCONFIRMED_MESSAGE);
  }

  try {
    const envelope = await transferEnvelopeSchema.validate(raw);
    return envelope.data.transaction;
  } catch {
    // The server answered 2xx, so the transfer was very likely created — an
    // unreadable body must not be reported as a failure.
    throw new TransferError("unconfirmed", UNCONFIRMED_MESSAGE);
  }
}
