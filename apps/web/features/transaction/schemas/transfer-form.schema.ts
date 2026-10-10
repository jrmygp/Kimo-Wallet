import * as yup from "yup";
import { formatMoney, parseAmountInput } from "@/lib/money";

export const MIN_TRANSFER_AMOUNT = 10000;

/**
 * What the user types on the transfer form. The amount stays a string in
 * form state and is converted once, by parseAmountInput, on the way to the
 * API (docs/CLAUDE.md §5.3.7). `balance` is the last balance the client
 * knows (from login), so this is a UX hint only — the authoritative check is
 * wallet-service's locked debit (§3.4.4). Pass null when it's unknown.
 */
export function createTransferFormSchema({ balance }: { balance: number | null }) {
  return yup.object({
    amount: yup
      .string()
      .trim()
      .required("Enter an amount")
      .matches(/^\d+$/, "Use digits only, without dots or commas")
      .test("safe-integer", "That amount is too large", (value) => value === undefined || parseAmountInput(value) !== null)
      .test(
        "min",
        `The minimum is ${formatMoney(MIN_TRANSFER_AMOUNT, "IDR")}`,
        (value) => {
          const amount = value === undefined ? null : parseAmountInput(value);
          return amount === null || amount >= MIN_TRANSFER_AMOUNT;
        },
      )
      .test("balance", "Insufficient balance", (value) => {
        const amount = value === undefined ? null : parseAmountInput(value);
        return amount === null || balance === null || amount <= balance;
      }),
    note: yup.string().max(100, "Keep the note under 100 characters"),
  });
}

export type TransferFormValues = yup.InferType<ReturnType<typeof createTransferFormSchema>>;
