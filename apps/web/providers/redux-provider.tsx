"use client";

import { Provider } from "react-redux";
import { PersistGate } from "redux-persist/integration/react";
import { persistor, store } from "@/lib/store/store";

export function ReduxProvider({ children }: { children: React.ReactNode }) {
  // loading={null} rather than a spinner: this app's persisted state
  // (user profile) isn't needed to render the first paint of any current
  // page, so there's nothing worth blocking on until rehydration finishes.
  return (
    <Provider store={store}>
      <PersistGate loading={null} persistor={persistor}>
        {children}
      </PersistGate>
    </Provider>
  );
}
