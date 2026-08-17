export type TransactionDirection = "in" | "out";

export interface Transaction {
  id: string;
  counterpartyName: string;
  /** ISO 8601 timestamp. */
  occurredAt: string;
  /** Whole Rupiah, always positive — sign is derived from `direction`. */
  amount: number;
  direction: TransactionDirection;
}
