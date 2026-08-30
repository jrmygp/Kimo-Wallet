import axios, { AxiosError } from "axios";
import { store } from "@/lib/store/store";
import { clearUser } from "@/features/auth/store/user-slice";

const AxiosInstance = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API,
});

AxiosInstance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("token");
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

AxiosInstance.interceptors.response.use(
  (response) => {
    return response.data;
  },
  (error) => {
    // The authoritative logout trigger: the gateway itself said this
    // token is missing/invalid/expired. A proactive client-side timer
    // (see features/auth/hooks/use-auto-logout.ts) is only a UX nicety on
    // top of this — this is what actually enforces it. Skip the redirect
    // if we're already on the login page, so an unauthenticated request
    // made from there (there isn't one today, but a future one might not
    // be) can't loop.
    if (
      error instanceof AxiosError &&
      error.response?.status === 401 &&
      window.location.pathname !== "/auth/login"
    ) {
      localStorage.removeItem("token");
      store.dispatch(clearUser());
      // A full reload, not router.push: this file is a plain module, not
      // a component, so useRouter() isn't available here — and a hard
      // reset is the right behavior for a forced logout anyway, clearing
      // TanStack Query's in-memory cache and any other stale client state
      // a soft navigation would leave behind.
      // eslint-disable-next-line @next/next/no-location-assign-relative-destination
      window.location.href = "/auth/login";
    }
    return Promise.reject(error);
  }
);

export { AxiosInstance };
