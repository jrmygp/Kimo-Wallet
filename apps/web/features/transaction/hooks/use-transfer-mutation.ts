import { useMutation } from "@tanstack/react-query";
import { transferFunds } from "@/features/transaction/api/transfer";

export function useTransferMutation() {
  return useMutation({
    mutationFn: transferFunds,
  });
}