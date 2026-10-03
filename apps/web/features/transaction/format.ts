import { formatMoney } from "@/lib/money";
import type { Transaction } from "./types";

const timeOptions: Intl.DateTimeFormatOptions = { hour: "2-digit", minute: "2-digit", hour12: false };
const thisYearFormatter = new Intl.DateTimeFormat("en-GB", { day: "numeric", month: "short", ...timeOptions });
const otherYearFormatter = new Intl.DateTimeFormat("en-GB", {
  day: "numeric",
  month: "short",
  year: "numeric",
  ...timeOptions,
});

// The year is only shown when it isn't the current one, so a row's subtitle fits on one
// line at 360px next to the amount.
export function formatTransactionDate(occurredAt: string): string {
  const date = new Date(occurredAt);
  const formatter = date.getFullYear() === new Date().getFullYear() ? thisYearFormatter : otherYearFormatter;
  return formatter.format(date);
}

const timeFormatter = new Intl.DateTimeFormat("en-GB", timeOptions);
const dayFormatter = new Intl.DateTimeFormat("en-GB", { weekday: "short", day: "numeric", month: "short" });
const dayOtherYearFormatter = new Intl.DateTimeFormat("en-GB", {
  weekday: "short",
  day: "numeric",
  month: "short",
  year: "numeric",
});

export function formatTransactionTime(occurredAt: string): string {
  return timeFormatter.format(new Date(occurredAt));
}

/** Local calendar day, used to group transactions under one date header. */
export function transactionDayKey(occurredAt: string): string {
  const date = new Date(occurredAt);
  return `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`;
}

/** "Today", "Yesterday", "Wed, 1 Oct", or "Wed, 1 Oct 2025" for another year. */
export function formatTransactionDay(occurredAt: string): string {
  const date = new Date(occurredAt);
  const today = new Date();
  const yesterday = new Date(today.getFullYear(), today.getMonth(), today.getDate() - 1);
  const key = transactionDayKey(occurredAt);
  if (key === transactionDayKey(today.toISOString())) return "Today";
  if (key === transactionDayKey(yesterday.toISOString())) return "Yesterday";
  return (date.getFullYear() === today.getFullYear() ? dayFormatter : dayOtherYearFormatter).format(date);
}

// U+2212 (minus sign), not a hyphen, so "−" lines up with "+" in tabular figures.
export function formatTransactionAmount(transaction: Transaction): string {
  const sign = transaction.direction === "in" ? "+" : "−";
  return `${sign}${formatMoney(transaction.amount, "IDR")}`;
}
