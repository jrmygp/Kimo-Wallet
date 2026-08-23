"use client";

import { useState } from "react";
import Page from "@/components/layout/Page";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  MdOutlineQrCode2,
  MdOutlineViewList,
  MdOutlineAddCard,
  MdSend,
  MdSettings,
  MdEmail,
  MdOutlineVisibility,
  MdOutlineVisibilityOff,
} from "react-icons/md";
import image1 from "@/public/images/carousel1.jpeg";
import image2 from "@/public/images/carousel2.jpeg";
import image3 from "@/public/images/carousel3.jpeg";
import Image from "next/image";
import Link from "next/link";
import { TransactionRow } from "@/features/transaction/components/transaction-row";
import type { Transaction } from "@/features/transaction/types";
import { useAppSelector } from "@/lib/store/hooks";

const menuItemClassName =
  "flex flex-col items-center gap-1.5 cursor-pointer rounded-md py-3 transition-all duration-300 hover:bg-white hover:shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-kimo-500";

const iconWrapperClassName = "flex size-11 items-center justify-center rounded-full bg-white text-kimo-500 shadow-sm";

const transactions: Transaction[] = [
  { id: "1", counterpartyName: "Jane Doe", occurredAt: "2026-08-15T17:20:00", amount: 50000, direction: "out" },
  { id: "2", counterpartyName: "John Smith", occurredAt: "2026-08-15T13:00:00", amount: 100000, direction: "in" },
  { id: "3", counterpartyName: "Jane Doe", occurredAt: "2026-08-15T17:20:00", amount: 50000, direction: "out" },
  { id: "4", counterpartyName: "Jane Doe", occurredAt: "2026-08-15T17:20:00", amount: 50000, direction: "out" },
  { id: "5", counterpartyName: "Jane Doe", occurredAt: "2026-08-15T17:20:00", amount: 50000, direction: "out" },
];
const HomePage = () => {
  const [balanceHidden, setBalanceHidden] = useState(false);
  const userData = useAppSelector((state) => state.user);
  console.log(userData.user)

  return (
    <Page>
      <div className="flex min-h-full w-full flex-col gap-8 items-center">
        {/* Profile section */}
        <section className="bg-kimo-500 w-full px-4 h-40 flex flex-col py-4 sm:flex-row sm:py-0 sm:items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <Avatar size="2xl">
              <AvatarImage src="https://github.com/shadcn.png" />
              <AvatarFallback>CN</AvatarFallback>
            </Avatar>

            <div className="flex flex-col">
              <p className="font-bold text-xl text-white">Jhon Doe</p>
              <p className="text-sm text-white/80">+6281234567890</p>
            </div>
          </div>

          <div className="flex flex-col items-start sm:items-end">
            <p className="font-bold text-xl text-white text-left sm:text-right">My Balance</p>
            <div className="flex items-center gap-2">
              <p className="text-2xl text-white text-left sm:text-right tabular-nums">
                {balanceHidden ? "Rp ••••••" : "Rp 100.000"}
              </p>
              <button
                type="button"
                onClick={() => setBalanceHidden((hidden) => !hidden)}
                aria-label={balanceHidden ? "Show balance" : "Hide balance"}
                aria-pressed={balanceHidden}
                className="cursor-pointer text-white/80 transition-colors hover:text-white"
              >
                {balanceHidden ? <MdOutlineVisibilityOff size={20} /> : <MdOutlineVisibility size={20} />}
              </button>
            </div>
          </div>
        </section>

        {/* Menus */}
        <section className="bg-white px-4 sm:px-10 flex flex-col w-full">
          <div className="rounded-md p-4 grid grid-cols-3 gap-y-4 bg-kimo-50">
            <button type="button" className={menuItemClassName}>
              <span className={iconWrapperClassName}>
                <MdOutlineAddCard size={22} />
              </span>
              <p className="text-sm text-center font-medium">Top Up</p>
            </button>

            <button type="button" className={menuItemClassName}>
              <span className={iconWrapperClassName}>
                <MdSend size={22} />
              </span>
              <p className="text-sm text-center font-medium">Transfer</p>
            </button>

            <Link href="/wallet/qr" className={menuItemClassName}>
              <span className={iconWrapperClassName}>
                <MdOutlineQrCode2 size={22} />
              </span>
              <p className="text-sm text-center font-medium">QR Code</p>
            </Link>

            <Link href="/wallet/history" className={menuItemClassName}>
              <span className={iconWrapperClassName}>
                <MdOutlineViewList size={22} />
              </span>
              <p className="text-sm text-center font-medium">History</p>
            </Link>

            <button type="button" className={menuItemClassName}>
              <span className={iconWrapperClassName}>
                <MdEmail size={22} />
              </span>
              <p className="text-sm text-center font-medium">Inbox</p>
            </button>

            <button type="button" className={menuItemClassName}>
              <span className={iconWrapperClassName}>
                <MdSettings size={22} />
              </span>
              <p className="text-sm text-center font-medium">Settings</p>
            </button>
          </div>
        </section>

        {/* Promotion carousel */}
        <section
          aria-label="Promotions"
          className="flex w-full min-w-0 snap-x snap-mandatory items-center gap-4 overflow-x-auto px-10 pb-1"
        >
          <div className="w-96 shrink-0 snap-center overflow-hidden rounded-md shadow-lg">
            <Image src={image1} alt="Promotion 1" />
          </div>
          <div className="w-96 shrink-0 snap-center overflow-hidden rounded-md shadow-lg">
            <Image src={image2} alt="Promotion 2" />
          </div>
          <div className="w-96 shrink-0 snap-center overflow-hidden rounded-md shadow-lg">
            <Image src={image3} alt="Promotion 3" />
          </div>
        </section>

        {/* Latest transactions */}
        <section className="w-full flex flex-col gap-6 px-4 sm:px-10 border-t border-border bg-muted/40 py-6">
          <p className="font-medium text-left text-foreground text-lg">Latest Transactions</p>

          <div className="flex flex-col gap-4">
            {transactions.map((transaction) => (
              <TransactionRow key={transaction.id} transaction={transaction} />
            ))}
          </div>
        </section>
      </div>
    </Page>
  );
};

export default HomePage;
