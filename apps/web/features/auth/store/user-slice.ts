import { createSlice, type PayloadAction } from "@reduxjs/toolkit";
import type { UserResponseData } from "@/features/auth/schemas/user-response.schema";

// Reuses the shape already validated by userResponseSchema instead of
// hand-declaring a duplicate interface, so the two can't drift.
export type UserProfile = UserResponseData["user"];
export type UserBalance = UserResponseData["wallet"];

interface UserState {
  user: UserProfile | null;
  balance: UserBalance | null;
}

const initialState: UserState = {
  user: null,
  balance: null,
};

const userSlice = createSlice({
  name: "user",
  initialState,
  reducers: {
    setUser(state, action: PayloadAction<UserProfile>) {
      state.user = action.payload;
    },
    setBalance(state, action: PayloadAction<UserBalance>) {
      state.balance = action.payload;
    },
    clearUser(state) {
      state.user = null;
      state.balance = null;
    },
  },
});

export const { setUser, setBalance, clearUser } = userSlice.actions;
export default userSlice.reducer;
