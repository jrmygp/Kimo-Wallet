"use client";

import Image from "next/image";
import Link from "next/link";
import { HistoryIcon, PlusIcon, QrCodeIcon, SendIcon } from "lucide-react";
import image1 from "@/public/images/carousel1.jpeg";
import image2 from "@/public/images/carousel2.jpeg";
import image3 from "@/public/images/carousel3.jpeg";
import { AppColumn } from "@/components/layout/app-column";
import { RecentTransactions } from "@/features/transaction/components/recent-transactions";
import { TransferRecipientSheet } from "@/features/wallet/components/transfer-recipient-sheet";
import { WalletSummary } from "@/features/wallet/components/wallet-summary";
import { useAppSelector } from "@/lib/store/hooks";

const actionClassName =
  "relative flex min-h-[84px] flex-col items-center justify-center gap-1.5 rounded-xl px-1 py-2 text-[13px] leading-4 font-medium text-foreground transition-colors focus-visible:ring-3 focus-visible:ring-ring/40 focus-visible:outline-none";

const actionIconClassName = "flex size-11 items-center justify-center rounded-full bg-kimo-50 text-primary";

const promotions = [
  { src: image1, alt: "Transfer to all banks and e-wallets with no admin fee." },
  { src: image2, alt: "Kimo, rated number one in Asia for security and reach." },
  { src: image3, alt: "Coming soon: send funds to crypto wallets such as ETH and LTC." },
];

const HomePage = () => {
  const userData = useAppSelector((state) => state.user);

  return (
    <AppColumn className="pb-[max(2rem,env(safe-area-inset-bottom))]">
        <h1 className="sr-only">Home</h1>

        <WalletSummary user={userData.user} balance={userData.balance} />

        <nav
          aria-label="Wallet actions"
          className="relative mx-4 -mt-10 grid grid-cols-4 gap-1 rounded-2xl border border-border bg-card p-2 shadow-[0_1px_2px_rgb(16_24_40/0.06)]"
        >
          <button
            type="button"
            aria-disabled
            title="Top Up is coming soon"
            className={`${actionClassName} cursor-not-allowed text-muted-foreground`}
          >
            <span className={`${actionIconClassName} bg-muted text-muted-foreground`}>
              <PlusIcon className="size-5" aria-hidden />
            </span>
            Top Up
            <span className="absolute top-1 right-1 rounded-full bg-amber-100 px-1.5 text-xs leading-4 font-semibold text-amber-800">
              Soon
            </span>
            <span className="sr-only">(coming soon)</span>
          </button>

          <TransferRecipientSheet triggerClassName={`${actionClassName} cursor-pointer hover:bg-muted`}>
            <span className={actionIconClassName}>
              <SendIcon className="size-5" aria-hidden />
            </span>
            Transfer
          </TransferRecipientSheet>

          <Link href="/wallet/qr" className={`${actionClassName} hover:bg-muted`}>
            <span className={actionIconClassName}>
              <QrCodeIcon className="size-5" aria-hidden />
            </span>
            QRIS
          </Link>

          <Link href="/wallet/history" className={`${actionClassName} hover:bg-muted`}>
            <span className={actionIconClassName}>
              <HistoryIcon className="size-5" aria-hidden />
            </span>
            History
          </Link>
        </nav>

        <div className="mt-6">
          <RecentTransactions />
        </div>

        <section aria-labelledby="promotions-heading" className="mt-8">
          <h2 id="promotions-heading" className="px-4 text-lg leading-6 font-semibold text-foreground">
            For you
          </h2>
          {/* Focusable so keyboard users can scroll it with the arrow keys. */}
          <div
            tabIndex={0}
            aria-label="Promotions, scroll horizontally"
            className="mt-3 flex snap-x snap-mandatory scroll-px-4 gap-3 overflow-x-auto px-4 pb-2 focus-visible:ring-3 focus-visible:ring-ring/40 focus-visible:outline-none"
          >
            {promotions.map((promotion) => (
              <div
                key={promotion.alt}
                className="w-[85%] shrink-0 snap-start overflow-hidden rounded-2xl border border-border bg-card"
              >
                <Image src={promotion.src} alt={promotion.alt} sizes="(min-width: 640px) 380px, 85vw" />
              </div>
            ))}
          </div>
        </section>
    </AppColumn>
  );
};

export default HomePage;
