import { useDispatch, useSelector } from "react-redux";
import type { AppDispatch, RootState } from "@/lib/store/store";

// Typed wrappers over react-redux's hooks so callers don't need `: RootState`
// annotations at every use site.
export const useAppDispatch = useDispatch.withTypes<AppDispatch>();
export const useAppSelector = useSelector.withTypes<RootState>();
