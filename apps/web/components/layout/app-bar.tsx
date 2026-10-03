import Link from "next/link";
import { ArrowLeftIcon } from "lucide-react";

// Sticky top bar for every screen below Home: one back target, one title. The title is
// the page's h1.
export function AppBar({ title, backHref }: { title: string; backHref: string }) {
  return (
    <header className="sticky top-0 z-10 grid h-14 shrink-0 grid-cols-[44px_1fr_44px] items-center border-b border-border bg-card/95 px-2 pt-[env(safe-area-inset-top)] backdrop-blur supports-backdrop-filter:bg-card/85">
      <Link
        href={backHref}
        aria-label="Back"
        className="flex size-11 items-center justify-center rounded-full text-foreground hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/40 focus-visible:outline-none"
      >
        <ArrowLeftIcon className="size-5" aria-hidden />
      </Link>
      <h1 className="truncate text-center text-base font-semibold text-foreground">{title}</h1>
    </header>
  );
}
