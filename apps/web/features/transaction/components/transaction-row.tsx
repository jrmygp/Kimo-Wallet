import { ArrowDownLeftIcon, ArrowUpRightIcon } from "lucide-react";
import { formatTransactionAmount, formatTransactionDate, formatTransactionTime } from "@/features/transaction/format";
import type { Transaction } from "@/features/transaction/types";
import { cn } from "@/lib/utils";

// Direction is carried by the icon, the "Received/Sent" label and the sign — colour only
// reinforces it. Outgoing money is not red: red is reserved for failures.
// `dateStyle="time"` is for lists that already group rows under a date header.
export function TransactionRow({
  transaction,
  dateStyle = "full",
}: {
  transaction: Transaction;
  dateStyle?: "full" | "time";
}) {
  const incoming = transaction.direction === "in";
  const Icon = incoming ? ArrowDownLeftIcon : ArrowUpRightIcon;

  return (
    <div className="flex items-center gap-3">
      <span
        className={cn(
          "flex size-10 shrink-0 items-center justify-center rounded-full",
          incoming ? "bg-emerald-50 text-emerald-700" : "bg-muted text-foreground",
        )}
      >
        <Icon className="size-5" aria-hidden />
      </span>

      <div className="min-w-0 flex-1">
        <p className="truncate text-[15px] leading-5 font-medium text-foreground">{transaction.counterpartyName}</p>
        <p className="text-xs leading-4 text-muted-foreground">
          {incoming ? "Received" : "Sent"} ·{" "}
          {dateStyle === "time"
            ? formatTransactionTime(transaction.occurredAt)
            : formatTransactionDate(transaction.occurredAt)}
        </p>
      </div>

      <p
        className={cn(
          "shrink-0 text-[15px] font-semibold tabular-nums",
          incoming ? "text-emerald-700" : "text-foreground",
        )}
      >
        {formatTransactionAmount(transaction)}
      </p>
    </div>
  );
}
