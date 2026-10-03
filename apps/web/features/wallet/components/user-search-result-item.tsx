import { ChevronRightIcon } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import type { WalletUser } from "@/features/wallet/schemas/user.schema";
import { getInitials } from "@/lib/utils";

export function UserSearchResultItem({ user, onClick }: { user: WalletUser; onClick?: (userId: string) => void }) {
  return (
    <button
      type="button"
      role="option"
      aria-selected={false}
      onClick={() => onClick?.(user.kimoId)}
      className="flex min-h-16 w-full cursor-pointer items-center gap-3 rounded-xl px-3 py-2 text-left transition-colors hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/40 focus-visible:outline-none"
    >
      <Avatar size="lg">
        {user.profilePicture && <AvatarImage src={user.profilePicture} alt="" />}
        <AvatarFallback className="bg-kimo-100 font-semibold text-kimo-800">{getInitials(user.fullName)}</AvatarFallback>
      </Avatar>

      <div className="min-w-0 flex-1">
        <p className="truncate text-[15px] font-medium text-foreground">{user.fullName}</p>
        <p className="text-xs text-muted-foreground">KimoID {user.kimoId}</p>
      </div>

      <ChevronRightIcon className="size-5 shrink-0 text-muted-foreground" aria-hidden />
    </button>
  );
}
