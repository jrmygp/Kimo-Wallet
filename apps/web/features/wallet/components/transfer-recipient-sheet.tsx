"use client";

import { useState, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { CircleAlertIcon, LoaderCircleIcon, SearchIcon, XIcon } from "lucide-react";
import { Sheet, SheetClose, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { useSearchUserQuery } from "@/features/wallet/hooks/use-search-user-query";
import { UserSearchResultItem } from "@/features/wallet/components/user-search-result-item";

export function TransferRecipientSheet({ triggerClassName, children }: { triggerClassName?: string; children: ReactNode }) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  // Only set on Enter (see the input's onKeyDown below), never on
  // every keystroke — that's what makes the search fire once per
  // submission instead of once per character typed.
  const [searchId, setSearchId] = useState("");
  const {
    data: matchedUser,
    isFetching: isSearching,
    isError: searchFailed,
    error: searchError,
  } = useSearchUserQuery(searchId);

  const onClickUser = (userId: string) => {
    setOpen(false);
    router.push(`/wallet/transfer/${userId}`);
  };

  return (
    <Sheet
      open={open}
      onOpenChange={(nextOpen) => {
        setOpen(nextOpen);
        if (!nextOpen) {
          // Start clean next time it's opened, rather than showing
          // a stale search from the previous visit.
          setQuery("");
          setSearchId("");
        }
      }}
    >
      <SheetTrigger className={triggerClassName}>{children}</SheetTrigger>

      <SheetContent
        side="bottom"
        showCloseButton={false}
        className="gap-0 rounded-t-3xl p-0 data-[side=bottom]:h-[80dvh] sm:mx-auto sm:max-w-md"
      >
        <SheetHeader className="gap-4 border-b border-border px-4 pt-3 pb-4">
          <span aria-hidden className="mx-auto h-1 w-10 rounded-full bg-border" />

          <div className="flex items-start justify-between gap-3">
            <div className="flex flex-col gap-1">
              <SheetTitle className="text-lg font-semibold">Send money</SheetTitle>
              <SheetDescription>Find the person you&apos;re paying by their KimoID.</SheetDescription>
            </div>
            <SheetClose
              aria-label="Close"
              className="-mt-1 -mr-2 flex size-11 shrink-0 cursor-pointer items-center justify-center rounded-full hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none"
            >
              <XIcon className="size-5" aria-hidden />
            </SheetClose>
          </div>

          <div className="flex h-12 items-center gap-2 rounded-xl border border-input bg-card px-3 transition-[border-color,box-shadow] focus-within:border-ring focus-within:ring-3 focus-within:ring-ring/25">
            <SearchIcon className="size-5 shrink-0 text-muted-foreground" aria-hidden />
            <input
              autoFocus
              type="search"
              enterKeyHint="search"
              aria-label="Recipient's KimoID"
              aria-describedby="transfer-search-hint"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              onKeyDown={(event) => {
                if (event.key !== "Enter") return;
                event.preventDefault();
                setSearchId(query.trim());
              }}
              placeholder="Enter KimoID"
              className="h-full min-w-0 flex-1 bg-transparent text-base outline-none placeholder:text-muted-foreground"
            />
          </div>
        </SheetHeader>

        <div role="listbox" aria-label="Search results" className="min-h-0 flex-1 overflow-y-auto px-2 py-2">
          {isSearching && (
            <p className="flex items-center justify-center gap-2 px-2 py-8 text-sm text-muted-foreground">
              <LoaderCircleIcon className="size-4 animate-spin motion-reduce:animate-none" aria-hidden />
              Searching…
            </p>
          )}

          {!isSearching && searchFailed && (
            <p role="alert" className="flex items-center justify-center gap-2 px-2 py-8 text-center text-sm text-destructive">
              <CircleAlertIcon className="size-4 shrink-0" aria-hidden />
              {searchError.message}
            </p>
          )}

          {!isSearching && !searchFailed && matchedUser && (
            <UserSearchResultItem user={matchedUser} onClick={(userId: string) => onClickUser(userId)} />
          )}

          {!isSearching && !searchFailed && !matchedUser && searchId.length === 0 && (
            <p id="transfer-search-hint" className="px-2 py-8 text-center text-sm text-muted-foreground">
              Type their KimoID, then press Search on your keyboard.
            </p>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}
