/**
 * One transfer the user means to make. Its idempotency key is minted when the
 * user reaches the review step and reused for every retry of the *same*
 * intent, so a retry after a timeout can never become a second transfer
 * (docs/CLAUDE.md §3.7.2, §5.3.9). It lives outside Formik state on purpose.
 */
export interface TransferIntent {
  idempotencyKey: string;
  receiverUserId: string;
  amount: number;
  currency: string;
}

/**
 * Returns `current` unchanged when the user is reviewing the same transfer
 * again (same receiver, amount and currency); otherwise a new intent with a
 * freshly minted key. Changing the amount or recipient is a different
 * transfer, so it must not reuse the old key.
 */
export function intentFor(
  current: TransferIntent | null,
  next: Omit<TransferIntent, "idempotencyKey">,
  mintKey: () => string,
): TransferIntent {
  if (
    current &&
    current.receiverUserId === next.receiverUserId &&
    current.amount === next.amount &&
    current.currency === next.currency
  ) {
    return current;
  }
  return { ...next, idempotencyKey: mintKey() };
}
