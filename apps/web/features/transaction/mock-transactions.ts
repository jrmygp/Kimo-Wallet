import type { Transaction, TransactionDirection } from "./types";

const NAMES = [
  "Jane Doe",
  "John Smith",
  "Siti Aminah",
  "Budi Santoso",
  "Coffee Shop",
  "Ahmad Fauzi",
  "Dewi Lestari",
  "Rian Pratama",
  "Nadia Putri",
  "Warung Makan Bu Tuti",
];

const AMOUNTS = [15000, 25000, 50000, 75000, 100000, 150000, 250000, 500000];

const TOTAL_TRANSACTIONS = 120;
const PAGE_SIZE = 15;
const SIMULATED_LATENCY_MS = 600;

function buildTransaction(index: number): Transaction {
  const direction: TransactionDirection = index % 3 === 0 ? "in" : "out";
  const daysAgo = Math.floor(index / 3);
  const occurredAt = new Date(Date.now() - daysAgo * 86_400_000 - (index % 24) * 3_600_000).toISOString();

  return {
    id: `txn_${index}`,
    counterpartyName: NAMES[index % NAMES.length],
    occurredAt,
    amount: AMOUNTS[index % AMOUNTS.length],
    direction,
  };
}

const ALL_TRANSACTIONS: Transaction[] = Array.from({ length: TOTAL_TRANSACTIONS }, (_, index) =>
  buildTransaction(index),
);

export interface TransactionsPage {
  items: Transaction[];
  hasMore: boolean;
}

/**
 * Simulates a paginated transaction-history API against a fixed in-memory pool.
 * Swap for a real fetch to the transaction service once it exists — callers
 * only depend on this function's signature, not its implementation.
 */
export async function fetchTransactionsPage(page: number): Promise<TransactionsPage> {
  await new Promise((resolve) => setTimeout(resolve, SIMULATED_LATENCY_MS));

  const start = page * PAGE_SIZE;
  const items = ALL_TRANSACTIONS.slice(start, start + PAGE_SIZE);

  return {
    items,
    hasMore: start + PAGE_SIZE < ALL_TRANSACTIONS.length,
  };
}
