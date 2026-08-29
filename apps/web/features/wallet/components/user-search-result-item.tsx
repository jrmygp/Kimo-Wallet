import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import type { WalletUser } from "@/features/wallet/schemas/user.schema";

export function UserSearchResultItem({ user, onClick }: { user: WalletUser; onClick?: () => void }) {
  return (
    <button
      type="button"
      role="option"
      aria-selected={false}
      onClick={onClick}
      className="flex w-full items-center gap-3 rounded-md px-2 py-2 text-left transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-kimo-500"
    >
      <Avatar size="lg">
        <AvatarImage src="https://github.com/shadcn.png" alt="" />
        <AvatarFallback>{user.fullName.slice(0, 2).toUpperCase()}</AvatarFallback>
      </Avatar>

      <div className="flex flex-col">
        <p className="text-foreground">{user.fullName}</p>
        <p className="text-xs text-muted-foreground">{user.phoneNumber}</p>
      </div>
    </button>
  );
}
