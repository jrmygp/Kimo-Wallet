import { useMutation } from "@tanstack/react-query";
import { registerUser } from "@/features/auth/api/register";

interface RegisterPayload {
  phoneNumber: string;
  fullName: string;
}

/**
 * Wraps registerUser as a TanStack Query mutation, following the same
 * one-hook-per-API-call pattern as useLoginMutation. registerUser takes two
 * positional args, but mutationFn can only take a single variables argument,
 * so this bundles them into one object.
 */
export function useRegisterMutation() {
  return useMutation({
    mutationFn: ({ phoneNumber, fullName }: RegisterPayload) => registerUser(phoneNumber, fullName),
  });
}
