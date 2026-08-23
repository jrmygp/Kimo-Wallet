"use client";

import { useState } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

export function QueryProvider({ children }: { children: React.ReactNode }) {
  // Created inside useState (not at module scope) so each browser session
  // gets its own client instance instead of one shared across requests —
  // matters even though this app is client-rendered end to end today.
  const [queryClient] = useState(() => new QueryClient());

  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}
