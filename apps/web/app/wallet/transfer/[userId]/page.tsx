"use client";

import { use, useEffect, useRef, useState, type ReactNode, type Ref } from "react";
import Link from "next/link";
import { useFormik } from "formik";
import { AlertCircleIcon, CheckIcon, ClockIcon, InfoIcon, LoaderCircleIcon, WifiOffIcon } from "lucide-react";
import { AppBar } from "@/components/layout/app-bar";
import { AppColumn } from "@/components/layout/app-column";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { TransferError } from "@/features/transaction/api/transfer";
import { useTransferMutation } from "@/features/transaction/hooks/use-transfer-mutation";
import { createTransferFormSchema, type TransferFormValues } from "@/features/transaction/schemas/transfer-form.schema";
import type { TransferTransaction } from "@/features/transaction/schemas/transfer.schema";
import { intentFor, type TransferIntent } from "@/features/transaction/transfer-intent";
import type { WalletUser } from "@/features/wallet/schemas/user.schema";
import { useSearchUserQuery } from "@/features/wallet/hooks/use-search-user-query";
import { useAppSelector } from "@/lib/store/hooks";
import { currencySymbol, formatMoney, parseAmountInput } from "@/lib/money";
import { useOnlineStatus } from "@/lib/use-online-status";
import { getInitials } from "@/lib/utils";

// Transfer flow: form → review → result. Money rules this page relies on
// (docs/CLAUDE.md §3.7, §5.3):
// - The amount is a string in form state, converted once by parseAmountInput.
// - The idempotency key belongs to a TransferIntent kept outside Formik. It's
//   minted when the user reaches review and reused for every retry of the same
//   transfer; it's only discarded after a terminal response (success or 409).
// - Confirm is disabled while a request is in flight and while offline.
// - A missing or unreadable response is shown as "confirming", never "failed".
// Not yet implemented: PIN entry before sending (PRD §7.5) and status polling
// (there's no endpoint to read a transaction back yet).

type Step =
  | { name: "form" }
  | { name: "review"; intent: TransferIntent }
  | { name: "unconfirmed"; intent: TransferIntent }
  | { name: "submitted"; transaction: TransferTransaction }
  | { name: "duplicate" };

const OFFLINE_MESSAGE =
  "You're offline. Please check your connection and try again.";

const TransferPage = ({ params }: { params: Promise<{ userId: string }> }) => {
  const { userId } = use(params);
  const { data: recipient, isFetching, isError, error } = useSearchUserQuery(userId);
  const balance = useAppSelector((state) => state.user.balance);
  const online = useOnlineStatus();
  console.log(online)
  const transferMutation = useTransferMutation();

  const [step, setStep] = useState<Step>({ name: "form" });
  const [rejection, setRejection] = useState<string | null>(null);
  // The intent from the last time the user reached review. Kept across
  // "Edit" so coming back with the same amount reuses the same key.
  const lastIntentRef = useRef<TransferIntent | null>(null);
  const stepHeadingRef = useRef<HTMLHeadingElement>(null);
  const currency = balance?.currency ?? "IDR";

  const formik = useFormik<TransferFormValues>({
    initialValues: { amount: "", note: "" },
    validationSchema: createTransferFormSchema({ balance: balance?.balance ?? null }),
    onSubmit: (values) => {
      const amount = parseAmountInput(values.amount);
      if (!recipient || amount === null) return;
      const intent = intentFor(
        lastIntentRef.current,
        { receiverUserId: recipient.id, amount, currency, note: values.note },
        () => crypto.randomUUID(),
      );
      lastIntentRef.current = intent;
      setRejection(null);
      setStep({ name: "review", intent });
    },
  });

  // Move focus to the new step's heading so screen-reader and keyboard users
  // land on what changed.
  useEffect(() => {
    if (step.name !== "form") stepHeadingRef.current?.focus();
  }, [step.name]);

  function send(intent: TransferIntent) {
    if (!online || transferMutation.isPending) return;
    setRejection(null);
    transferMutation.mutate(
      {
        idempotencyKey: intent.idempotencyKey,
        receiverUserId: intent.receiverUserId,
        currency: intent.currency,
        amount: intent.amount,
        note: intent.note
      },
      {
        onSuccess: (transaction) => {
          lastIntentRef.current = null;
          formik.resetForm();
          setStep({ name: "submitted", transaction });
        },
        onError: (err) => {
          if (err instanceof TransferError && err.kind === "duplicate") {
            lastIntentRef.current = null;
            formik.resetForm();
            setStep({ name: "duplicate" });
            return;
          }
          if (err instanceof TransferError && err.kind === "rejected") {
            // Nothing was created; stay on review with the same intent.
            setRejection(err.message);
            return;
          }
          // Unknown outcome: keep the intent (and its key) so "Try again"
          // can only ever repeat this same transfer.
          setStep({ name: "unconfirmed", intent });
        },
      },
    );
  }

  // Strip anything but digits as the user types/pastes, instead of letting a
  // bad character in and only flagging it on blur/submit — the Yup schema
  // (transfer-form.schema.ts) still re-validates on top of this, since this
  // is UX only, never the authority (docs/CLAUDE.md §5.3.5).
  function handleAmountChange(e: React.ChangeEvent<HTMLInputElement>) {
    formik.setFieldValue("amount", e.target.value.replace(/\D/g, ""));
  }

  const amountInvalid = !!formik.touched.amount && !!formik.errors.amount;
  const noteInvalid = !!formik.touched.note && !!formik.errors.note;
  const sending = transferMutation.isPending;

  return (
    <AppColumn>
      <AppBar title="Transfer" backHref="/home" />

      <div className="flex flex-1 flex-col px-4 pt-5 pb-[max(1.5rem,env(safe-area-inset-bottom))]">
        {!online && (
          <div
            role="status"
            className="mb-4 flex items-start gap-2 rounded-xl bg-amber-50 px-3 py-3 text-sm text-amber-900"
          >
            <WifiOffIcon className="mt-0.5 size-4 shrink-0" aria-hidden />
            <p>{OFFLINE_MESSAGE}</p>
          </div>
        )}

        {step.name === "form" && (
          <>
            <RecipientCard recipient={recipient} loading={isFetching} />

            {isError && (
              <div
                role="alert"
                className="mt-3 flex items-start gap-2 rounded-xl bg-destructive/10 px-3 py-3 text-sm text-destructive"
              >
                <AlertCircleIcon className="mt-0.5 size-4 shrink-0" aria-hidden />
                <p>{error.message}</p>
              </div>
            )}

            <form className="mt-6 flex flex-col" onSubmit={formik.handleSubmit} noValidate>
              <Field data-invalid={amountInvalid}>
                <FieldLabel htmlFor="amount" className="text-[13px] text-foreground">
                  Amount
                </FieldLabel>
                <div
                  className={
                    amountInvalid
                      ? "flex items-center gap-2 rounded-2xl border border-destructive bg-card px-4 py-4 ring-3 ring-destructive/15"
                      : "flex items-center gap-2 rounded-2xl border border-input bg-card px-4 py-4 transition-[border-color,box-shadow] focus-within:border-ring focus-within:ring-3 focus-within:ring-ring/25"
                  }
                >
                  <span className="text-2xl font-semibold text-muted-foreground">{currencySymbol(currency)}</span>
                  <Input
                    id="amount"
                    name="amount"
                    value={formik.values.amount}
                    onChange={handleAmountChange}
                    onBlur={formik.handleBlur}
                    placeholder="0"
                    inputMode="numeric"
                    autoComplete="off"
                    aria-invalid={amountInvalid}
                    aria-describedby="amount-message"
                    className="h-auto border-0 bg-transparent p-0 text-3xl font-bold tabular-nums shadow-none focus-visible:ring-0 md:text-3xl"
                  />
                </div>
                <FieldDescription
                  id="amount-message"
                  className={`flex min-h-5 items-center gap-1.5 ${amountInvalid ? "text-destructive" : "text-muted-foreground"}`}
                >
                  {amountInvalid ? (
                    <>
                      <AlertCircleIcon className="size-4 shrink-0" aria-hidden />
                      {formik.errors.amount}
                    </>
                  ) : (
                    balance && <>Balance {formatMoney(balance.balance, balance.currency)}</>
                  )}
                </FieldDescription>
              </Field>

              <Field data-invalid={noteInvalid} className="mt-4">
                <FieldLabel htmlFor="note" className="text-[13px] text-foreground">
                  Note <span className="font-normal text-muted-foreground">(optional)</span>
                </FieldLabel>
                <Input
                  id="note"
                  name="note"
                  value={formik.values.note}
                  onChange={formik.handleChange}
                  onBlur={formik.handleBlur}
                  placeholder="What's this for?"
                  aria-invalid={noteInvalid}
                  aria-describedby="note-error"
                />
                <FieldDescription id="note-error" className="flex min-h-5 items-center gap-1.5">
                  {noteInvalid && (
                    <>
                      <AlertCircleIcon className="size-4 shrink-0" aria-hidden />
                      {formik.errors.note}
                    </>
                  )}
                </FieldDescription>
              </Field>

              <Button
                type="submit"
                variant="primary"
                className="mt-2 h-13 rounded-xl text-[15px] font-semibold"
                disabled={!recipient || isFetching || !online}
              >
                Review transfer
              </Button>
            </form>
          </>
        )}

        {step.name === "review" && (
          <section aria-labelledby="review-heading" className="flex flex-col">
            <h2
              id="review-heading"
              ref={stepHeadingRef}
              tabIndex={-1}
              className="text-lg font-semibold text-foreground outline-none"
            >
              Review transfer
            </h2>

            <dl className="mt-4 divide-y divide-border rounded-2xl border border-border bg-card">
              <div className="flex items-center justify-between gap-3 px-4 py-3">
                <dt className="text-sm text-muted-foreground">To</dt>
                <dd className="min-w-0 text-right">
                  <p className="truncate text-[15px] font-semibold text-foreground">{recipient?.fullName}</p>
                  <p className="text-xs text-muted-foreground">KimoID {recipient?.kimoId}</p>
                </dd>
              </div>
              <div className="flex items-center justify-between gap-3 px-4 py-3">
                <dt className="text-sm text-muted-foreground">Amount</dt>
                <dd className="text-xl font-bold text-foreground tabular-nums">
                  {formatMoney(step.intent.amount, step.intent.currency)}
                </dd>
              </div>
              <div className="flex items-center justify-between gap-3 px-4 py-3">
                <dt className="text-sm text-muted-foreground">From</dt>
                <dd className="text-[15px] font-medium text-foreground">Kimo balance</dd>
              </div>
            </dl>

            {rejection && (
              <div
                role="alert"
                className="mt-3 flex items-start gap-2 rounded-xl bg-destructive/10 px-3 py-3 text-sm text-destructive"
              >
                <AlertCircleIcon className="mt-0.5 size-4 shrink-0" aria-hidden />
                <p>{rejection}</p>
              </div>
            )}

            <div className="mt-6 grid grid-cols-[auto_1fr] gap-3">
              <Button
                type="button"
                variant="outline"
                onClick={() => setStep({ name: "form" })}
                disabled={sending}
                className="h-13 rounded-xl border-border bg-card px-5 text-[15px] font-semibold"
              >
                Edit
              </Button>
              <Button
                type="button"
                variant="primary"
                onClick={() => send(step.intent)}
                disabled={sending || !online}
                className="h-13 rounded-xl text-[15px] font-semibold"
              >
                {sending ? (
                  <>
                    <LoaderCircleIcon className="size-5 animate-spin motion-reduce:animate-none" aria-hidden />
                    Sending…
                  </>
                ) : (
                  "Confirm and send"
                )}
              </Button>
            </div>
          </section>
        )}

        {step.name === "unconfirmed" && (
          <ResultPanel
            headingRef={stepHeadingRef}
            icon={<ClockIcon className="size-7" aria-hidden />}
            tone="pending"
            title="We're confirming your transfer"
          >
            <p>
              We didn&apos;t get an answer in time, so your transfer of{" "}
              <span className="font-semibold text-foreground">
                {formatMoney(step.intent.amount, step.intent.currency)}
              </span>{" "}
              may still go through. Check your history before sending it again.
            </p>
            <p className="mt-2">Trying again from here repeats this same transfer, so you won&apos;t be charged twice.</p>
            <div className="mt-6 grid grid-cols-2 gap-3">
              <Link
                href="/wallet/history"
                className="flex h-13 items-center justify-center rounded-xl border border-border bg-card text-[15px] font-semibold text-foreground hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/40 focus-visible:outline-none"
              >
                Check history
              </Link>
              <Button
                type="button"
                variant="primary"
                onClick={() => send(step.intent)}
                disabled={sending || !online}
                className="h-13 rounded-xl text-[15px] font-semibold"
              >
                {sending ? (
                  <>
                    <LoaderCircleIcon className="size-5 animate-spin motion-reduce:animate-none" aria-hidden />
                    Checking…
                  </>
                ) : (
                  "Try again"
                )}
              </Button>
            </div>
          </ResultPanel>
        )}

        {step.name === "submitted" && (
          <ResultPanel
            headingRef={stepHeadingRef}
            icon={<CheckIcon className="size-7" aria-hidden />}
            tone="success"
            title="Transfer submitted"
          >
            <p className="text-2xl font-bold text-foreground tabular-nums">
              {formatMoney(step.transaction.amount, step.transaction.currency)}
            </p>
            <p className="mt-1">to {recipient?.fullName ?? "your recipient"}</p>
            <p className="mt-4">We&apos;re processing it now. It will show in your history once it completes.</p>
            <p className="mt-2 text-xs break-all">Transaction ID {step.transaction.id}</p>
            <ResultLinks />
          </ResultPanel>
        )}

        {step.name === "duplicate" && (
          <ResultPanel
            headingRef={stepHeadingRef}
            icon={<InfoIcon className="size-7" aria-hidden />}
            tone="info"
            title="Already submitted"
          >
            <p>This transfer was already submitted, so we didn&apos;t send it again. Check your history for its status.</p>
            <ResultLinks />
          </ResultPanel>
        )}
      </div>
    </AppColumn>
  );
};

function RecipientCard({ recipient, loading }: { recipient: WalletUser | undefined; loading: boolean }) {
  return (
    <section
      aria-label="Recipient"
      className="flex min-h-[72px] items-center gap-3 rounded-2xl border border-border bg-card p-4 shadow-[0_1px_2px_rgb(16_24_40/0.06)]"
    >
      {loading ? (
        <>
          <span className="size-12 shrink-0 animate-pulse rounded-full bg-muted motion-reduce:animate-none" />
          <span className="flex flex-1 flex-col gap-2">
            <span className="h-4 w-32 animate-pulse rounded bg-muted motion-reduce:animate-none" />
            <span className="h-3 w-24 animate-pulse rounded bg-muted motion-reduce:animate-none" />
          </span>
        </>
      ) : recipient ? (
        <>
          <Avatar size="xl">
            {recipient.profilePicture && <AvatarImage src={recipient.profilePicture} alt="" />}
            <AvatarFallback className="bg-kimo-100 font-semibold text-kimo-800">
              {getInitials(recipient.fullName)}
            </AvatarFallback>
          </Avatar>
          <div className="min-w-0 flex-1">
            <p className="truncate text-base font-semibold text-foreground">{recipient.fullName}</p>
            <p className="text-sm text-muted-foreground">KimoID {recipient.kimoId}</p>
          </div>
        </>
      ) : (
        <p className="text-sm text-muted-foreground">Recipient not found.</p>
      )}
    </section>
  );
}

const resultToneClassName = {
  success: "bg-emerald-50 text-emerald-700",
  pending: "bg-amber-50 text-amber-700",
  info: "bg-kimo-50 text-primary",
} as const;

function ResultPanel({
  headingRef,
  icon,
  tone,
  title,
  children,
}: {
  headingRef: Ref<HTMLHeadingElement>;
  icon: ReactNode;
  tone: keyof typeof resultToneClassName;
  title: string;
  children: ReactNode;
}) {
  return (
    <section aria-labelledby="result-heading" className="flex flex-col items-center pt-6 text-center">
      <span className={`flex size-14 items-center justify-center rounded-full ${resultToneClassName[tone]}`}>
        {icon}
      </span>
      <h2
        id="result-heading"
        ref={headingRef}
        tabIndex={-1}
        className="mt-4 text-xl font-semibold text-foreground outline-none"
      >
        {title}
      </h2>
      <div className="mt-2 w-full text-sm leading-5 text-muted-foreground">{children}</div>
    </section>
  );
}

function ResultLinks() {
  return (
    <div className="mt-6 grid grid-cols-2 gap-3">
      <Link
        href="/wallet/history"
        className="flex h-13 items-center justify-center rounded-xl border border-border bg-card text-[15px] font-semibold text-foreground hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/40 focus-visible:outline-none"
      >
        View history
      </Link>
      <Link
        href="/home"
        className="flex h-13 items-center justify-center rounded-xl bg-primary text-[15px] font-semibold text-primary-foreground hover:bg-kimo-800 focus-visible:ring-3 focus-visible:ring-ring/40 focus-visible:outline-none"
      >
        Back to home
      </Link>
    </div>
  );
}

export default TransferPage;
