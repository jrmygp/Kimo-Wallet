import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

// The signed-in app frame: full-bleed on phones, a 448px card centred on a tinted
// background from `sm` up. Pass height classes via `className` for screens that manage
// their own scroll container (see app/wallet/history/page.tsx).
export function AppColumn({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div className="flex min-h-dvh w-full justify-center bg-canvas sm:items-start sm:bg-kimo-50/60 sm:py-8">
      <main
        className={cn(
          "flex min-h-dvh w-full flex-col bg-canvas sm:min-h-0 sm:max-w-md sm:overflow-hidden sm:rounded-3xl sm:border sm:border-border sm:shadow-[0_12px_32px_-12px_rgb(16_24_40/0.12)]",
          className,
        )}
      >
        {children}
      </main>
    </div>
  );
}
