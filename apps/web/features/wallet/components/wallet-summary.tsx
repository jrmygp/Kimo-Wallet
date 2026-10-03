"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAppDispatch } from "@/lib/store/hooks";
import { clearUser } from "@/features/auth/store/user-slice";
import { BellIcon, CheckIcon, CopyIcon, EyeIcon, EyeOffIcon, LogOutIcon, SettingsIcon } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import type { UserBalance, UserProfile } from "@/features/auth/store/user-slice";
import { currencySymbol, formatAmount } from "@/lib/money";
import { getInitials } from "@/lib/utils";
import { Dialog, DialogTrigger } from "@/components/ui/dialog";
import LogoutConfirmationModal from "./logout-confirmation-modal";

const iconButtonClassName =
  "flex size-11 items-center justify-center rounded-full transition-colors focus-visible:ring-3 focus-visible:ring-white/60 focus-visible:outline-none";

// The one teal surface on Home: who you are, what you have, and the KimoID people need
// to pay you. Text sits on teal-700/800 only — white on teal-500 fails contrast.
export function WalletSummary({ user, balance }: { user: UserProfile | null; balance: UserBalance | null }) {
  const dispatch = useAppDispatch();
  const router = useRouter();
  const [balanceHidden, setBalanceHidden] = useState(false);
  const [copied, setCopied] = useState(false);

  const logoutHandler = () => {
    dispatch(clearUser());
    router.push("/auth/login");
  };

  useEffect(() => {
    if (!copied) return;
    const timeout = setTimeout(() => setCopied(false), 2000);
    return () => clearTimeout(timeout);
  }, [copied]);

  async function copyKimoId() {
    if (!user) return;
    try {
      await navigator.clipboard.writeText(user.kimoId);
      setCopied(true);
    } catch {
      // Clipboard unavailable (unsupported browser or denied permission) — the ID stays
      // visible on screen, so there's nothing else to recover.
    }
  }

  const firstName = user?.fullName.trim().split(/\s+/)[0] ?? "";
  const symbol = currencySymbol(balance?.currency ?? "IDR");

  return (
    <section className="bg-linear-to-br from-kimo-700 to-kimo-800 px-4 pt-[max(1rem,env(safe-area-inset-top))] pb-14 text-white">
      <Dialog>
        <div className="flex items-center gap-3">
          <Avatar size="xl" className="after:border-white/30">
            {user?.profilePicture && <AvatarImage src={user.profilePicture} alt="" />}
            <AvatarFallback className="bg-white/15 font-semibold text-white">
              {user ? getInitials(user.fullName) : ""}
            </AvatarFallback>
          </Avatar>

          <div className="min-w-0 flex-1">
            <p className="text-sm leading-5 text-kimo-100">Hi, {firstName}</p>
            <p className="truncate text-base leading-6 font-semibold">{user?.fullName}</p>
          </div>

          <button
            type="button"
            aria-disabled
            title="Inbox is coming soon"
            className={`${iconButtonClassName} cursor-not-allowed text-white/60`}
          >
            <BellIcon className="size-5" aria-hidden />
            <span className="sr-only">Inbox (coming soon)</span>
          </button>
          <button
            type="button"
            aria-disabled
            title="Settings is coming soon"
            className={`${iconButtonClassName} cursor-not-allowed text-white/60`}
          >
            <SettingsIcon className="size-5" aria-hidden />
            <span className="sr-only">Settings (coming soon)</span>
          </button>

          <DialogTrigger
            render={
              <button
                type="button"
                aria-label="Log out"
                className={`${iconButtonClassName} -mr-2 cursor-pointer text-white/80 hover:bg-white/10 hover:text-white`}
              />
            }
          >
            <LogOutIcon className="size-5" aria-hidden />
          </DialogTrigger>
        </div>

        <LogoutConfirmationModal onLogout={logoutHandler} />
      </Dialog>

      <div className="mt-6">
        <p className="text-sm leading-5 text-kimo-100">Kimo balance</p>

        <div className="mt-1 flex items-center gap-1">
          {balance ? (
            <p className="flex items-baseline gap-1.5 tabular-nums">
              <span className="text-xl font-semibold">{symbol}</span>
              <span className="text-[36px] leading-11 font-bold tracking-tight">
                {balanceHidden ? "••••••" : formatAmount(balance.balance)}
              </span>
              {balanceHidden && <span className="sr-only">Balance hidden</span>}
            </p>
          ) : (
            <p className="text-base leading-11 font-medium text-kimo-100">Your wallet is being set up…</p>
          )}

          {balance && (
            <button
              type="button"
              onClick={() => setBalanceHidden((hidden) => !hidden)}
              aria-label={balanceHidden ? "Show balance" : "Hide balance"}
              aria-pressed={balanceHidden}
              className={`${iconButtonClassName} cursor-pointer text-kimo-100 hover:bg-white/10 hover:text-white`}
            >
              {balanceHidden ? (
                <EyeOffIcon className="size-5" aria-hidden />
              ) : (
                <EyeIcon className="size-5" aria-hidden />
              )}
            </button>
          )}
        </div>

        {user && (
          <div className="mt-3 flex items-center gap-1 text-sm text-kimo-100">
            <span>
              KimoID <span className="font-semibold text-white tabular-nums">{user.kimoId}</span>
            </span>
            <button
              type="button"
              onClick={copyKimoId}
              aria-label={copied ? "KimoID copied" : "Copy KimoID"}
              className={`${iconButtonClassName} -my-3 cursor-pointer hover:bg-white/10 hover:text-white`}
            >
              {copied ? <CheckIcon className="size-4" aria-hidden /> : <CopyIcon className="size-4" aria-hidden />}
            </button>
            <span aria-live="polite" className="text-xs font-medium text-white">
              {copied ? "Copied" : ""}
            </span>
          </div>
        )}
      </div>
    </section>
  );
}
