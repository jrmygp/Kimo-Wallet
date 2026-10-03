import Image from "next/image";
import type { FormEventHandler, ReactNode } from "react";
import kimo from "@/public/images/kimo.png";

// Shared frame for the login and register screens: full-bleed on phones, a centred
// card on the canvas from `sm` up. The only teal surface is the brand band at the top.
export function AuthShell({
  onSubmit,
  children,
}: {
  onSubmit: FormEventHandler<HTMLFormElement>;
  children: ReactNode;
}) {
  return (
    <div className="flex min-h-dvh w-full justify-center bg-canvas sm:items-center sm:px-4 sm:py-10">
      <form
        className="flex min-h-dvh w-full flex-col overflow-hidden bg-card sm:min-h-0 sm:max-w-md sm:rounded-3xl sm:border sm:border-border sm:shadow-[0_12px_32px_-12px_rgb(16_24_40/0.12)]"
        onSubmit={onSubmit}
      >
        <header className="flex justify-center bg-linear-to-br from-kimo-700 to-kimo-600 pt-6 pb-12">
          {/* The asset is a white mark with wide transparent padding; the negative
              margin trims that padding so the band stays compact. */}
          <Image src={kimo} alt="Kimo" priority className="-my-6 w-44" />
        </header>

        <div className="-mt-6 flex flex-1 flex-col rounded-t-3xl bg-card px-6 pt-8 pb-[max(1.5rem,env(safe-area-inset-bottom))]">
          {children}
        </div>
      </form>
    </div>
  );
}
