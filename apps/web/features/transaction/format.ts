import type { Transaction } from "./types";

const dateFormatter = new Intl.DateTimeFormat("en-GB", {
  day: "2-digit",
  month: "short",
  year: "numeric",
  hour: "2-digit",
  minute: "2-digit",
  hour12: false,
});

export function formatTransactionDate(occurredAt: string): string {
  return dateFormatter.format(new Date(occurredAt));
}

export function formatTransactionAmount(transaction: Transaction): string {
  const sign = transaction.direction === "in" ? "+" : "-";
  return `${sign} Rp ${transaction.amount.toLocaleString("id-ID")}`;
}
