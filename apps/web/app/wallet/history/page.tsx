"use client";

import { useEffect, useRef, useState } from "react";
import InfiniteScroll from "react-infinite-scroll-component";
import Link from "next/link";
import { MdArrowBack, MdAutorenew } from "react-icons/md";
import Page from "@/components/layout/Page";
import { TransactionRow } from "@/features/transaction/components/transaction-row";
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

  return (
    <Page>
      <div className="flex min-h-full w-full flex-col">
        <div className="sticky top-0 z-10 flex items-center gap-4 border-b border-border bg-white px-4 py-4">
          <Link href="/home" aria-label="Back" className="text-foreground">
            <MdArrowBack size={24} />
          </Link>
          <p className="text-lg font-medium text-foreground">Transaction History</p>
        </div>

        <div className="flex flex-col gap-4 px-4 py-6 sm:px-10">
          {initialLoading ? (
            <div className="flex items-center justify-center gap-2 py-10 text-muted-foreground">
              <MdAutorenew size={20} className="animate-spin" />
              <p className="text-sm">Loading transactions…</p>
            </div>
          ) : (
            <InfiniteScroll
              dataLength={transactions.length}
              next={loadMore}
              hasMore={hasMore}
              scrollableTarget="page-scroll-container"
              className="flex flex-col gap-4"
              loader={
                <div className="flex items-center justify-center gap-2 py-4 text-muted-foreground">
                  <MdAutorenew size={18} className="animate-spin" />
                  <p className="text-sm">Loading more…</p>
                </div>
              }
              endMessage={
                <p className="py-4 text-center text-sm text-muted-foreground">
                  You&apos;ve reached the end of your transaction history.
                </p>
              }
            >
              {transactions.map((transaction) => (
                <TransactionRow key={transaction.id} transaction={transaction} />
              ))}
            </InfiniteScroll>
          )}
        </div>
      </div>
    </Page>
  );
};

export default HistoryPage;
