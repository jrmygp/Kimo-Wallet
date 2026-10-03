import type { Metadata } from "next";
import { jakarta } from "@/lib/fonts";
import "./globals.css";
import { cn } from "@/lib/utils";
import { QueryProvider } from "@/providers/query-provider";
import { ReduxProvider } from "@/providers/redux-provider";
import { AuthWatcher } from "@/providers/auth-watcher";

export const metadata: Metadata = {
  title: "Kimo Wallet",
  description: "Send, receive and pay with your Kimo digital wallet.",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en"
      className={cn("h-full", "antialiased", jakarta.variable, "font-sans")}
    >
      <body className="min-h-full flex flex-col">
        <ReduxProvider>
          <AuthWatcher />
          <QueryProvider>{children}</QueryProvider>
        </ReduxProvider>
      </body>
    </html>
  );
}
