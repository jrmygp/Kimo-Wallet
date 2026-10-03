"use client";

import Link from "next/link";
import { useFormik } from "formik";
import { LoaderCircleIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Field, FieldLabel } from "@/components/ui/field";
import { AuthShell } from "@/features/auth/components/auth-shell";
import { FieldMessage, FormAlert } from "@/features/auth/components/form-feedback";
import { PhoneNumberField } from "@/features/auth/components/phone-number-field";
import { registerValidation } from "@/features/auth/schemas/register.schema";
import { useRegisterMutation } from "@/features/auth/hooks/use-register-mutation";
import { useRouter } from "next/navigation";
import { all } from "country-codes-list";
import { useRef, type FormEvent } from "react";

const countries = all();

const PHONE_ALREADY_REGISTERED = "phone number is already registered";

const RegisterPage = () => {
  const router = useRouter();
  const registerMutation = useRegisterMutation();
  const numberInputRef = useRef<HTMLInputElement>(null);
  const fullNameInputRef = useRef<HTMLInputElement>(null);
  const phoneNumberArr = localStorage.getItem("phoneNumber")?.split("-") ?? [];

  const formik = useFormik({
    enableReinitialize: true,
    initialValues: {
      country: phoneNumberArr[0] ?? "",
      number: phoneNumberArr[1] ?? "",
      fullName: "",
    },
    validationSchema: registerValidation,
    onSubmit: (values) => {
      const callingCode = countries.find((country) => country.countryCode === values.country)?.countryCallingCode;
      if (!callingCode) return;
      const phoneNumber = `+${callingCode}${values.number}`;

      registerMutation.mutate(
        { fullName: values.fullName, phoneNumber },
        {
          onSuccess: () => {
            // Registering no longer starts a session (see api-gateway's
            // Register handler) — the account exists, but the user must
            // log in separately with the same phone number. Keep it in
            // localStorage under the same key/format the login page
            // itself writes when it redirects *here* on "user not
            // found", so the login page can prefill it.
            localStorage.setItem("phoneNumber", `${values.country}-${values.number}`);
            router.push("/auth/login");
          },
          onError(error) {
            if (error.message === PHONE_ALREADY_REGISTERED) {
              formik.setFieldError("number", error.message);
            }
          },
        },
      );
    },
  });

  const countryInvalid = !!formik.touched.country && !!formik.errors.country;
  const numberInvalid = !!formik.touched.number && !!formik.errors.number;
  const fullNameInvalid = !!formik.touched.fullName && !!formik.errors.fullName;
  const phoneInvalid = countryInvalid || numberInvalid;
  const phoneError = (countryInvalid && formik.errors.country) || (numberInvalid && formik.errors.number) || "";

  // Formik's submit marks every field touched and validates; when that fails, move focus
  // to the first invalid field in visual order (docs/CLAUDE.md §5.3.6).
  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    formik.handleSubmit(event);
    const errors = await formik.validateForm();
    if (errors.country || errors.number) {
      numberInputRef.current?.focus();
    } else if (errors.fullName) {
      fullNameInputRef.current?.focus();
    }
  }

  return (
    <AuthShell onSubmit={handleSubmit}>
      <h1 className="text-[22px] leading-7 font-semibold text-foreground">Create your Kimo account</h1>
      <p className="mt-2 text-sm leading-5 text-muted-foreground">
        It only takes a minute. You&apos;ll log in with this number once your account is ready.
      </p>

      <PhoneNumberField
        id="register-number"
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

      <Field data-invalid={fullNameInvalid} className="mt-3 gap-2">
        <FieldLabel htmlFor="register-full-name" className="text-[13px] text-foreground">
          Full name
        </FieldLabel>
        <Input
          ref={fullNameInputRef}
          id="register-full-name"
          name="fullName"
          autoComplete="name"
          autoCapitalize="words"
          value={formik.values.fullName}
          onChange={formik.handleChange}
          onBlur={formik.handleBlur}
          placeholder="e.g. Budi Santoso"
          aria-invalid={fullNameInvalid}
          aria-describedby="register-full-name-message"
          className="h-13 rounded-xl border-input bg-card px-3 text-base shadow-none focus-visible:ring-ring/25 aria-invalid:ring-3 aria-invalid:ring-destructive/15 md:text-base"
        />
        <FieldMessage
          id="register-full-name-message"
          error={fullNameInvalid ? formik.errors.fullName : undefined}
          hint="10–50 characters, shown to senders."
        />
      </Field>

      {registerMutation.isError && registerMutation.error.message !== PHONE_ALREADY_REGISTERED && (
        <FormAlert message={registerMutation.error.message} />
      )}

      <div className="mt-8 flex flex-col gap-3">
        <Button
          type="submit"
          variant="primary"
          disabled={registerMutation.isPending}
          className="h-13 w-full rounded-xl text-base font-semibold"
        >
          {registerMutation.isPending ? (
            <>
              <LoaderCircleIcon className="size-5 animate-spin motion-reduce:animate-none" aria-hidden />
              Creating your account…
            </>
          ) : (
            "Create account"
          )}
        </Button>
        <p className="text-center text-xs leading-5 text-muted-foreground">
          By creating an account, you agree to Kimo&apos;s Terms &amp; Conditions and Privacy Notice.
        </p>
      </div>

      <p className="mt-4 text-center text-sm text-muted-foreground">
        Already have an account?{" "}
        <Link
          href="/auth/login"
          className="inline-flex min-h-11 items-center font-semibold text-primary underline-offset-4 hover:underline focus-visible:underline focus-visible:outline-none"
        >
          Log in
        </Link>
      </p>
    </AuthShell>
  );
};

export default RegisterPage;
