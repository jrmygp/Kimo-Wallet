import * as yup from "yup";

export const transferSchema = yup.object().shape({
  idempotencyKey: yup.string().required(),
  receiverUserId: yup.string().required(),
  currency: yup.string().required(),
  amount: yup.number().required(),
});

export type TransferRequestData = yup.InferType<typeof transferSchema>;

// The transaction as POST /v1/transaction returns it (api-gateway's
// createTransactionResponseBody). No wallet ids: the gateway never sends them.
export const transferTransactionSchema = yup.object({
  id: yup.string().required(),
  idempotencyKey: yup.string().required(),
  senderUserId: yup.string().required(),
  receiverUserId: yup.string().required(),
  currency: yup.string().required(),
  amount: yup.number().integer().required(),
  status: yup.string().required(),
  createdAt: yup.string().required(),
});

export type TransferTransaction = yup.InferType<typeof transferTransactionSchema>;

// The envelope's `data` wraps the transaction: `{ transaction: {...} }`.
export const transferResponseSchema = yup.object({
  transaction: transferTransactionSchema.required(),
});