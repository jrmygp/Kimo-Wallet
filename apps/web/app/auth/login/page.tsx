"use client";

import { all } from "country-codes-list";
import { LoaderCircleIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { AuthShell } from "@/features/auth/components/auth-shell";
import { FormAlert } from "@/features/auth/components/form-feedback";
import { PhoneNumberField } from "@/features/auth/components/phone-number-field";
import { useFormik } from "formik";
import { loginValidation } from "@/features/auth/schemas/login.schema";
import { useLoginMutation } from "@/features/auth/hooks/use-login-mutation";
import { useRouter } from "next/navigation";
import { useAppDispatch } from "@/lib/store/hooks";
import { setUser, setBalance } from "@/features/auth/store/user-slice";
import { useEffect, useRef, type FormEvent } from "react";
import { useAppSelector } from "@/lib/store/hooks";

const countries = all();

const LoginPage = () => {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const loginMutation = useLoginMutation();
  const userData = useAppSelector((state) => state.user);
  const numberInputRef = useRef<HTMLInputElement>(null);
  // Set by the register page on a successful registration (redirecting
  // here, since Register no longer starts a session — see api-gateway's
  // Register handler) and, symmetrically, by this page itself below when
  // it redirects to /auth/register on "user not found". Same key/format
  // both ways round.
  const phoneNumberArr = localStorage.getItem("phoneNumber")?.split("-") ?? [];

  const formik = useFormik({
    enableReinitialize: true,
    initialValues: {
      country: phoneNumberArr[0] ?? "ID",
      number: phoneNumberArr[1] ?? "",
    },
    validationSchema: loginValidation,
    onSubmit: (values) => {
      // CountryCodeSelect only ever sets `country` to a code present in
      // this same `countries` list, so a missing calling code here can't
      // actually happen through the UI — guarded anyway rather than
      // building a phone number with a literal "undefined" in it.
      const callingCode = countries.find((country) => country.countryCode === values.country)?.countryCallingCode;
      if (!callingCode) return;
      const phoneNumber = `+${callingCode}${values.number}`;

      loginMutation.mutate(phoneNumber, {
        onSuccess: (data) => {
          localStorage.removeItem("phoneNumber");
          localStorage.setItem("token", data.accessToken);
          dispatch(setUser(data.user));
          dispatch(setBalance(data.wallet));
          router.push("/home");
        },
        onError: (error) => {
          if (error.message === "user not found") {
            router.push("/auth/register");
            localStorage.setItem("phoneNumber", `${values.country}-${values.number}`);
          }
        },
      });
    },
  });

  const countryInvalid = !!formik.touched.country && !!formik.errors.country;
  const numberInvalid = !!formik.touched.number && !!formik.errors.number;
  const phoneInvalid = countryInvalid || numberInvalid;
  const phoneError = (countryInvalid && formik.errors.country) || (numberInvalid && formik.errors.number) || "";

  useEffect(() => {
    if (userData.user?.id) {
      router.push("/home");
    }
  }, [userData.user?.id]);

  // Formik's submit marks every field touched and validates; when that fails, move focus
  // to the invalid field (docs/CLAUDE.md §5.3.6) so keyboard and screen-reader users land
  // on the error instead of the button.
  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    formik.handleSubmit(event);
    const errors = await formik.validateForm();
    if (errors.country || errors.number) {
      numberInputRef.current?.focus();
    }
  }

  return (
    <AuthShell onSubmit={handleSubmit}>
      <h1 className="text-[22px] leading-7 font-semibold text-foreground">Enter your mobile number</h1>
      <p className="mt-2 text-sm leading-5 text-muted-foreground">
        Log in with the number linked to your Kimo account. New to Kimo? We&apos;ll help you sign up next.
      </p>

      <PhoneNumberField
        id="login-number"
        className="mt-8 gap-2"
        inputRef={numberInputRef}
        name="number"
        value={formik.values.number}
        onChange={formik.handleChange}
        onBlur={formik.handleBlur}
        country={formik.values.country}
        onCountryChange={(code) => {
          formik.setFieldValue("country", code);
          formik.setFieldTouched("country", true);
        }}
        invalid={phoneInvalid}
        error={phoneError}
      />

      <p className="mt-3 text-sm text-muted-foreground">
        Lost access to your number?{" "}
        <button
          type="button"
          className="-my-3 inline-flex min-h-11 cursor-pointer items-center font-semibold text-primary underline-offset-4 hover:underline focus-visible:underline focus-visible:outline-none"
        >
          Change number
        </button>
      </p>

      {loginMutation.isError && loginMutation.error.message !== "user not found" && (
        <FormAlert message={loginMutation.error.message} />
      )}

      <div className="mt-8 flex flex-col gap-3">
        <Button
          type="submit"
          variant="primary"
          disabled={loginMutation.isPending}
          className="h-13 w-full rounded-xl text-base font-semibold"
        >
          {loginMutation.isPending ? (
            <>
              <LoaderCircleIcon className="size-5 animate-spin motion-reduce:animate-none" aria-hidden />
              Checking your number…
            </>
          ) : (
            "Continue"
          )}
        </Button>
        <p className="text-center text-xs leading-5 text-muted-foreground">
          By continuing, you agree to Kimo&apos;s Terms &amp; Conditions and Privacy Notice.
        </p>
      </div>
    </AuthShell>
  );
};

export default LoginPage;
