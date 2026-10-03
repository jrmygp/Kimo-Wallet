"use client";

import { useState } from "react";
import { CheckIcon, CopyIcon, ShareIcon, SunIcon } from "lucide-react";
import QRCode from "react-qr-code";
import { AppBar } from "@/components/layout/app-bar";
import { AppColumn } from "@/components/layout/app-column";
import { Button } from "@/components/ui/button";

const QRIS_VALUE =
  "00020101021226670015ID.SINGAPAY.WWW01189360783904122600420152886407023049170303UME51440014ID.CO.QRIS.WWW0215ID190900099006003UME6221051017389001790703CO153033605405110005502035702306015JAKARTASELATAN520458125914PAYCRED_ACQ_015802ID610458126304423E";

const QRISPage = () => {
  const [copied, setCopied] = useState(false);

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(QRIS_VALUE);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard unavailable (unsupported browser or denied permission) — nothing to confirm.
    }
  }

  async function handleShare() {
    if (navigator.share) {
      try {
        await navigator.share({
          title: "Kimo QRIS code",
          text: QRIS_VALUE,
        });
      } catch {
        // User dismissed the share sheet — not an error.
      }
      return;
    }
    handleCopy();
  }

  return (
    <AppColumn>
      <AppBar title="Receive money" backHref="/home" />

      <div className="flex flex-1 flex-col px-4 pt-6 pb-[max(1.5rem,env(safe-area-inset-bottom))]">
        <section
          aria-labelledby="qris-heading"
          className="flex flex-col items-center rounded-3xl border border-border bg-card px-6 pt-5 pb-6 text-center shadow-[0_1px_2px_rgb(16_24_40/0.06)]"
        >
          <span className="rounded-full bg-kimo-50 px-3 py-1 text-xs leading-4 font-bold tracking-widest text-primary">
            QRIS
          </span>
          <h2 id="qris-heading" className="mt-3 text-lg leading-6 font-semibold text-foreground">
            Scan to pay
          </h2>
          <p className="mt-1 text-sm leading-5 text-muted-foreground">
            Anyone with a QRIS-enabled app can scan this to send you money.
          </p>

          <div className="mt-5 w-full max-w-64 rounded-2xl border border-border bg-white p-3">
            <QRCode
              size={256}
              style={{ height: "auto", maxWidth: "100%", width: "100%" }}
              value={QRIS_VALUE}
              viewBox={`0 0 256 256`}
              title="QRIS payment code"
            />
          </div>
        </section>

        <div className="mt-4 grid grid-cols-2 gap-3">
          <Button
            variant="outline"
            onClick={handleCopy}
            className="h-12 rounded-xl border-border bg-card text-[15px] font-semibold"
          >
            {copied ? <CheckIcon className="size-5" aria-hidden /> : <CopyIcon className="size-5" aria-hidden />}
            {copied ? "Copied" : "Copy code"}
          </Button>
          <Button variant="primary" onClick={handleShare} className="h-12 rounded-xl text-[15px] font-semibold">
            <ShareIcon className="size-5" aria-hidden />
            Share
          </Button>
        </div>
        <p aria-live="polite" className="sr-only">
          {copied ? "QRIS code copied" : ""}
        </p>

        <p className="mt-auto flex items-start gap-2 pt-8 text-sm leading-5 text-muted-foreground">
          <SunIcon className="mt-0.5 size-4 shrink-0" aria-hidden />
          If the scan doesn&apos;t work, turn up your screen brightness and hold the phone steady.
        </p>
      </div>
    </AppColumn>
  );
};

export default QRISPage;
