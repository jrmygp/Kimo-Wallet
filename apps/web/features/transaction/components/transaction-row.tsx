import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { formatTransactionAmount, formatTransactionDate } from "@/features/transaction/format";
import type { Transaction } from "@/features/transaction/types";

export function TransactionRow({ transaction }: { transaction: Transaction }) {
  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center gap-2">
        <Avatar size="lg">
          <AvatarImage src="https://github.com/shadcn.png" />
          <AvatarFallback>{transaction.counterpartyName.slice(0, 2).toUpperCase()}</AvatarFallback>
        </Avatar>

        <div className="flex flex-col">
          <p className="text-foreground">{transaction.counterpartyName}</p>
          <p className="text-xs text-muted-foreground">{formatTransactionDate(transaction.occurredAt)}</p>
        </div>
      </div>

      <Badge
        variant={transaction.direction === "out" ? "destructive" : undefined}
        className={
          transaction.direction === "in"
            ? "bg-green-50 text-green-700 dark:bg-green-950 dark:text-green-300"
            : undefined
        }
      >
        {formatTransactionAmount(transaction)}
      </Badge>
    </div>
  );
}
