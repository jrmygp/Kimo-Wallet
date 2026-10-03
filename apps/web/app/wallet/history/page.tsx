"use client";

import { useEffect, useRef, useState } from "react";
import InfiniteScroll from "react-infinite-scroll-component";
import { LoaderCircleIcon, ReceiptTextIcon } from "lucide-react";
import { AppBar } from "@/components/layout/app-bar";
import { AppColumn } from "@/components/layout/app-column";
import { TransactionRow } from "@/features/transaction/components/transaction-row";
import { formatTransactionDay, transactionDayKey } from "@/features/transaction/format";
import { fetchTransactionsPage } from "@/features/transaction/mock-transactions";
import type { Transaction } from "@/features/transaction/types";

const HistoryPage = () => {
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [hasMore, setHasMore] = useState(true);
  const [initialLoading, setInitialLoading] = useState(true);

  // Imperative mirrors of the state above, read by the scroll listener below.
  // A listener attached once on mount would otherwise close over the state
  // values from that first render forever — these refs are always current.
  const pageRef = useRef(0);
  const hasMoreRef = useRef(true);
  const loadingRef = useRef(false);

  async function loadMore() {
    if (loadingRef.current || !hasMoreRef.current) return;
    loadingRef.current = true;
    try {
      const result = await fetchTransactionsPage(pageRef.current);
      pageRef.current += 1;
      hasMoreRef.current = result.hasMore;
      setTransactions((prev) => [...prev, ...result.items]);
      setHasMore(result.hasMore);
    } finally {
      loadingRef.current = false;
    }
  }

  useEffect(() => {
    let cancelled = false;
    loadingRef.current = true;

    fetchTransactionsPage(0).then((result) => {
      if (cancelled) return;
      pageRef.current = 1;
      hasMoreRef.current = result.hasMore;
      setTransactions(result.items);
      setHasMore(result.hasMore);
      setInitialLoading(false);
      loadingRef.current = false;
    });

    return () => {
      cancelled = true;
    };
  }, []);

  // Safety net: react-infinite-scroll-component's sentinel only fires `next`
  // on an intersection *transition* into view. A fast or programmatic scroll
  // that lands exactly at the max scrollable position in one motion never
  // produces that transition — the sentinel is already intersecting and
  // stays that way — so the library silently stalls even though more data
  // exists. A direct scroll listener on the real container is immune to
  // that: it re-checks distance-from-bottom on every scroll tick, not just
  // when new data arrives.
  useEffect(() => {
    const container = document.getElementById("page-scroll-container");
    if (!container) return;

    function handleScroll() {
      if (loadingRef.current || !hasMoreRef.current) return;
      const distanceFromBottom = container!.scrollHeight - container!.scrollTop - container!.clientHeight;
      if (distanceFromBottom < container!.clientHeight) {
        loadMore();
      }
    }

    container.addEventListener("scroll", handleScroll, { passive: true });
    handleScroll();

    return () => container.removeEventListener("scroll", handleScroll);
  }, []);

  // Consecutive rows that share a calendar day go under one date header. The source is
  // already newest-first, so grouping in order is enough.
  const groups: { key: string; label: string; items: Transaction[] }[] = [];
  for (const transaction of transactions) {
    const key = transactionDayKey(transaction.occurredAt);
    const last = groups[groups.length - 1];
    if (last && last.key === key) {
      last.items.push(transaction);
    } else {
      groups.push({ key, label: formatTransactionDay(transaction.occurredAt), items: [transaction] });
    }
  }

  return (
    // Fixed height + an inner scroll container (not window scroll): the scroll listener
    // above and InfiniteScroll's scrollableTarget both depend on #page-scroll-container.
    <AppColumn className="h-dvh sm:h-[min(880px,calc(100dvh-4rem))]">
      <AppBar title="Transaction history" backHref="/home" />

      <div id="page-scroll-container" className="min-h-0 flex-1 overflow-y-auto">
        <div className="px-4 pt-4 pb-[max(1.5rem,env(safe-area-inset-bottom))]">
          {initialLoading ? (
            <div aria-label="Loading transactions" className="overflow-hidden rounded-2xl border border-border bg-card">
              {Array.from({ length: 6 }, (_, index) => (
                <div key={index} className="flex items-center gap-3 border-b border-border px-4 py-3 last:border-b-0">
                  <span className="size-10 shrink-0 animate-pulse rounded-full bg-muted motion-reduce:animate-none" />
                  <span className="flex flex-1 flex-col gap-2">
                    <span className="h-3.5 w-1/2 animate-pulse rounded bg-muted motion-reduce:animate-none" />
                    <span className="h-3 w-1/3 animate-pulse rounded bg-muted motion-reduce:animate-none" />
                  </span>
                  <span className="h-3.5 w-16 animate-pulse rounded bg-muted motion-reduce:animate-none" />
                </div>
              ))}
            </div>
          ) : transactions.length === 0 ? (
            <div className="flex flex-col items-center gap-2 px-4 py-16 text-center">
              <ReceiptTextIcon className="size-8 text-muted-foreground" aria-hidden />
              <p className="text-base font-semibold text-foreground">No transactions yet</p>
              <p className="text-sm text-muted-foreground">Money you send and receive will show up here.</p>
            </div>
          ) : (
            <InfiniteScroll
              dataLength={transactions.length}
              next={loadMore}
              hasMore={hasMore}
              scrollableTarget="page-scroll-container"
              className="flex flex-col gap-5"
              loader={
                <p className="flex items-center justify-center gap-2 py-4 text-sm text-muted-foreground">
                  <LoaderCircleIcon className="size-4 animate-spin motion-reduce:animate-none" aria-hidden />
                  Loading more…
                </p>
              }
              endMessage={
                <p className="py-4 text-center text-sm text-muted-foreground">
                  That&apos;s everything. You&apos;ve reached your first transaction.
                </p>
              }
            >
              {groups.map((group) => (
                <section key={group.key} aria-labelledby={`day-${group.key}`}>
                  <h2
                    id={`day-${group.key}`}
                    className="px-1 pb-2 text-[13px] leading-4 font-semibold tracking-wide text-muted-foreground uppercase"
                  >
                    {group.label}
                  </h2>
                  <ul className="divide-y divide-border overflow-hidden rounded-2xl border border-border bg-card">
                    {group.items.map((transaction) => (
                      <li key={transaction.id} className="px-4 py-3">
                        <TransactionRow transaction={transaction} dateStyle="time" />
                      </li>
                    ))}
                  </ul>
                </section>
              ))}
            </InfiniteScroll>
          )}
        </div>
      </div>
    </AppColumn>
  );
};

export default HistoryPage;
