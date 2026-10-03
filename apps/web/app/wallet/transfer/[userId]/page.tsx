"use client";

import { use } from "react";
import { useFormik } from "formik";
import { AlertCircleIcon } from "lucide-react";
import { AppBar } from "@/components/layout/app-bar";
import { AppColumn } from "@/components/layout/app-column";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useSearchUserQuery } from "@/features/wallet/hooks/use-search-user-query";
import { currencySymbol } from "@/lib/money";
import { getInitials } from "@/lib/utils";

// NOTE: this is a visual restyle only — validationSchema, onSubmit and all
// other form logic are unchanged from before. There is no real
// transfer-creation endpoint on api-gateway yet, so submit is still a
// stub. See docs/agent-logs for the recorded gap: before this page is
// wired to a real endpoint, it still needs a real Yup schema, an
// idempotency key minted outside Formik state, and a disabled-while-
// submitting button per docs/CLAUDE.md §3.7/§5.3 — none of that changed here.
const TransferPage = ({ params }: { params: Promise<{ userId: string }> }) => {
  const { userId } = use(params);
  const { data, isFetching, isError, error } = useSearchUserQuery(userId);
  const formik = useFormik({
    initialValues: {
      amount: "",
      note: "",
    },
    validationSchema: null,
    onSubmit: (values) => {
      console.log(values);
    },
  });

  const amountInvalid = formik.touched.amount && !!formik.errors.amount;
  const noteInvalid = formik.touched.note && !!formik.errors.note;

  return (
    <AppColumn>
      <AppBar title="Transfer" backHref="/home" />

      <div className="flex flex-1 flex-col px-4 pt-5 pb-[max(1.5rem,env(safe-area-inset-bottom))]">
        <section
          aria-label="Recipient"
          className="flex min-h-[72px] items-center gap-3 rounded-2xl border border-border bg-card p-4 shadow-[0_1px_2px_rgb(16_24_40/0.06)]"
        >
          {isFetching ? (
            <>
              <span className="size-12 shrink-0 animate-pulse rounded-full bg-muted motion-reduce:animate-none" />
              <span className="flex flex-1 flex-col gap-2">
                <span className="h-4 w-32 animate-pulse rounded bg-muted motion-reduce:animate-none" />
                <span className="h-3 w-24 animate-pulse rounded bg-muted motion-reduce:animate-none" />
              </span>
            </>
          ) : data ? (
            <>
              <Avatar size="xl">
                {data.profilePicture && <AvatarImage src={data.profilePicture} alt="" />}
                <AvatarFallback className="bg-kimo-100 font-semibold text-kimo-800">
                  {getInitials(data.fullName)}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0 flex-1">
                <p className="truncate text-base font-semibold text-foreground">{data.fullName}</p>
                <p className="text-sm text-muted-foreground">KimoID {data.kimoId}</p>
              </div>
            </>
          ) : (
            <p className="text-sm text-muted-foreground">Recipient not found.</p>
          )}
        </section>

        {isError && (
          <div
            role="alert"
            className="mt-3 flex items-start gap-2 rounded-xl bg-destructive/10 px-3 py-3 text-sm text-destructive"
          >
            <AlertCircleIcon className="mt-0.5 size-4 shrink-0" aria-hidden />
            <p>{error.message}</p>
          </div>
        )}

        <form className="mt-6 flex flex-col" onSubmit={formik.handleSubmit}>
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
              <span className="text-2xl font-semibold text-muted-foreground">{currencySymbol("IDR")}</span>
              <Input
                id="amount"
                name="amount"
                value={formik.values.amount}
                onChange={formik.handleChange}
                onBlur={formik.handleBlur}
                placeholder="0"
                inputMode="numeric"
                aria-invalid={amountInvalid}
                aria-describedby="amount-error"
                className="h-auto border-0 bg-transparent p-0 text-3xl font-bold tabular-nums shadow-none focus-visible:ring-0 md:text-3xl"
              />
            </div>
            <FieldDescription id="amount-error" className="flex min-h-5 items-center gap-1.5">
              {amountInvalid && (
                <>
                  <AlertCircleIcon className="size-4 shrink-0" aria-hidden />
                  {formik.errors.amount}
                </>
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

          <Button type="submit" variant="primary" className="mt-2 h-13 rounded-xl text-[15px] font-semibold">
            Send Now
          </Button>
        </form>
      </div>
    </AppColumn>
  );
};

export default TransferPage;
