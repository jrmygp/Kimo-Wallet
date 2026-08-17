"use client";

import { useState } from "react";
import Page from "@/components/layout/Page";
import { Button } from "@/components/ui/button";
import QRCode from "react-qr-code";
import { MdArrowBack, MdOutlineFileUpload, MdOutlineContentCopy, MdCheck } from "react-icons/md";
import Link from "next/link";

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
    <Page>
      <div className="bg-kimo-500 flex min-h-full w-full flex-col items-center justify-between py-10 px-4">
        <div className="relative flex w-full items-center justify-center">
          <p className="text-white text-lg font-medium">Request from QRIS</p>

          <Link
            href={"/home"}
            className="absolute left-0 top-1/2 flex -translate-y-1/2 items-center gap-2 font-medium text-white sm:left-10"
          >
            <MdArrowBack size={24} />
            <p className="hidden sm:flex">Back</p>
          </Link>
        </div>

        <div className="flex flex-col items-center gap-4">
          <div className="max-w-72 rounded-lg bg-white p-4 shadow-lg">
            <QRCode
              size={256}
              style={{ height: "auto", maxWidth: "100%", width: "100%" }}
              value={QRIS_VALUE}
              viewBox={`0 0 256 256`}
            />
          </div>

          <p className="max-w-72 text-center text-sm text-white/80">
            Show this code so someone can scan it and pay you directly.
          </p>

          <Button onClick={handleCopy} variant={copied ? "outline" : "default"}>
            {copied ? <MdCheck size={20} /> : <MdOutlineContentCopy size={20} />}
            {copied ? "Copied" : "Copy code"}
          </Button>
        </div>

        <Button className="w-full" onClick={handleShare}>
          <MdOutlineFileUpload size={20} />
          Share QR code
        </Button>
      </div>
    </Page>
  );
};

export default QRISPage;
