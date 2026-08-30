import { createSlice, type PayloadAction } from "@reduxjs/toolkit";
import type { UserResponseData } from "@/features/auth/schemas/user-response.schema";

// Reuses the shape already validated by userResponseSchema instead of
// hand-declaring a duplicate interface, so the two can't drift.
export type UserProfile = UserResponseData["user"];

interface UserState {
  user: UserProfile | null;
}

const initialState: UserState = {
  user: null,
};

const userSlice = createSlice({
  name: "user",
  initialState,
  reducers: {
    setUser(state, action: PayloadAction<UserProfile>) {
      state.user = action.payload;
    },
    clearUser(state) {
      state.user = null;
    },
  },
});

export const { setUser, clearUser } = userSlice.actions;
export default userSlice.reducer;
