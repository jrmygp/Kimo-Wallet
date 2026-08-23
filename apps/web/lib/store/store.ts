import { combineReducers, configureStore } from "@reduxjs/toolkit";
import { FLUSH, PAUSE, PERSIST, PURGE, REGISTER, REHYDRATE, persistReducer, persistStore } from "redux-persist";
import storage from "redux-persist/lib/storage"; // localStorage
import userReducer from "@/features/auth/store/user-slice";

const rootReducer = combineReducers({
  user: userReducer,
});

const persistConfig = {
  key: "kimo-root",
  storage,
  // Only persist what's actually meant to survive a reload — add a slice
  // name here deliberately when it needs persisting, don't default to
  // persisting everything.
  whitelist: ["user"],
};

const persistedReducer = persistReducer(persistConfig, rootReducer);

export const store = configureStore({
  reducer: persistedReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      // redux-persist dispatches these with non-serializable payloads by
      // design; without this, RTK's default serializableCheck logs a
      // warning for every one of them on every app load.
      serializableCheck: {
        ignoredActions: [FLUSH, REHYDRATE, PAUSE, PERSIST, PURGE, REGISTER],
      },
    }),
});

export const persistor = persistStore(store);

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
