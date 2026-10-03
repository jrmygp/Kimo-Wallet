"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { CircleAlertIcon, ReceiptTextIcon } from "lucide-react";
import { TransactionRow } from "@/features/transaction/components/transaction-row";
import { fetchTransactionsPage } from "@/features/transaction/mock-transactions";

const RECENT_COUNT = 5;

// Reads the first page of the same (mock) source as /wallet/history, so the two screens
// always agree. Swap together with history once the transaction service exists.
export function RecentTransactions() {
  const { data, isPending, isError, refetch } = useQuery({
    queryKey: ["transactions", "page", 0],
    queryFn: () => fetchTransactionsPage(0),
  });

  const items = data?.items.slice(0, RECENT_COUNT) ?? [];

  return (
    <section aria-labelledby="recent-activity-heading" className="px-4">
      <div className="flex items-center justify-between">
        <h2 id="recent-activity-heading" className="text-lg leading-6 font-semibold text-foreground">
          Recent activity
        </h2>
        <Link
          href="/wallet/history"
          className="-mr-2 inline-flex min-h-11 items-center rounded-lg px-2 text-sm font-semibold text-primary hover:underline focus-visible:ring-3 focus-visible:ring-ring/40 focus-visible:outline-none"
        >
          See all
        </Link>
      </div>

      <div className="mt-2 overflow-hidden rounded-2xl border border-border bg-card">
        {isPending && (
          <ul aria-label="Loading recent activity" className="divide-y divide-border">
            {Array.from({ length: 3 }, (_, index) => (
              <li key={index} className="flex items-center gap-3 px-4 py-3">
                <span className="size-10 shrink-0 animate-pulse rounded-full bg-muted motion-reduce:animate-none" />
                <span className="flex flex-1 flex-col gap-2">
                  <span className="h-3.5 w-1/2 animate-pulse rounded bg-muted motion-reduce:animate-none" />
                  <span className="h-3 w-1/3 animate-pulse rounded bg-muted motion-reduce:animate-none" />
                </span>
                <span className="h-3.5 w-16 animate-pulse rounded bg-muted motion-reduce:animate-none" />
              </li>
            ))}
          </ul>
        )}

        {isError && (
          <div role="alert" className="flex flex-col items-center gap-3 px-4 py-8 text-center">
            <CircleAlertIcon className="size-6 text-destructive" aria-hidden />
            <p className="text-sm text-muted-foreground">We couldn&apos;t load your recent activity.</p>
            <button
              type="button"
              onClick={() => refetch()}
              className="min-h-11 rounded-lg px-4 text-sm font-semibold text-primary hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/40 focus-visible:outline-none"
            >
              Try again
            </button>
          </div>
        )}

        {!isPending && !isError && items.length === 0 && (
          <div className="flex flex-col items-center gap-2 px-4 py-8 text-center">
            <ReceiptTextIcon className="size-6 text-muted-foreground" aria-hidden />
            <p className="text-sm font-medium text-foreground">No activity yet</p>
            <p className="text-sm text-muted-foreground">Money you send and receive will show up here.</p>
          </div>
        )}

        {items.length > 0 && (
          <ul className="divide-y divide-border">
            {items.map((transaction) => (
              <li key={transaction.id} className="px-4 py-3">
                <TransactionRow transaction={transaction} />
              </li>
            ))}
          </ul>
        )}
      </div>
    </section>
  );
}
