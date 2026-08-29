"use client";

import { all } from "country-codes-list";
import kimo from "@/public/images/kimo.png";
import Image from "next/image";
import Page from "@/components/layout/Page";
import { Button } from "@/components/ui/button";
import { CountryCodeSelect } from "@/components/country-code-select";
import { Input } from "@/components/ui/input";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { useFormik } from "formik";
import { loginValidation } from "@/features/auth/schemas/login.schema";
import { useLoginMutation } from "@/features/auth/hooks/use-login-mutation";
import { useRouter } from "next/navigation";
import { useAppDispatch } from "@/lib/store/hooks";
import { setUser } from "@/features/auth/store/user-slice";
import { useEffect } from "react";
import { useAppSelector } from "@/lib/store/hooks";

const countries = all();

const LoginPage = () => {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const loginMutation = useLoginMutation();
  const userData = useAppSelector((state) => state.user);

  const formik = useFormik({
    initialValues: {
      country: "ID",
      number: "",
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
          localStorage.setItem("token", data.accessToken);
          dispatch(setUser(data.user));
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

  const countryInvalid = formik.touched.country && !!formik.errors.country;
  const numberInvalid = formik.touched.number && !!formik.errors.number;

  useEffect(() => {
    if (userData.user?.id) {
      router.push("/home")
    }
  }, [userData.user?.id]);

  return (
    <Page>
      <form
        className="bg-kimo-500 flex min-h-full w-full flex-col justify-between py-10 px-2 sm:px-0"
        onSubmit={formik.handleSubmit}
      >
        <section className="flex flex-col w-full items-center justify-center gap-4">
          <Image src={kimo} alt="kimo-logo" className="w-60 sm:w-80" />

          <p className="text-lg font-medium text-white text-center">Enter your mobile number to continue</p>

          <div className="flex w-full max-w-sm gap-2">
            <Field data-invalid={countryInvalid} className={countryInvalid ? "w-56" : "w-30"}>
              <FieldLabel className={countryInvalid ? "text-red-500" : undefined}>Country</FieldLabel>
              <CountryCodeSelect
                value={formik.values.country}
                onChange={(code) => {
                  formik.setFieldValue("country", code);
                  formik.setFieldTouched("country", true);
                }}
              />
              <FieldDescription>{countryInvalid && formik.errors.country}</FieldDescription>
            </Field>

            <Field data-invalid={numberInvalid}>
              <FieldLabel className={numberInvalid ? "text-red-500" : undefined}>Number</FieldLabel>
              <Input
                name="number"
                value={formik.values.number}
                onChange={formik.handleChange}
                onBlur={formik.handleBlur}
                placeholder="812-3456-7890"
                inputMode="numeric"
                aria-invalid={numberInvalid}
                className="bg-white"
              />
              <FieldDescription>{numberInvalid && formik.errors.number}</FieldDescription>
            </Field>
          </div>

          <div className="flex items-center gap-2">
            <p className="text-xs text-white">Lost or inactive number?</p>
            <Button className="rounded-2xl" size="sm">
              Change number
            </Button>
          </div>

          {loginMutation.isError && (
            <p role="alert" className="text-sm text-red-200 text-center max-w-sm">
              {loginMutation.error.message}
            </p>
          )}
        </section>

        <section className="flex flex-col gap-2 items-center justify-center">
          <p className="text-white text-sm text-center">By continuing, you are agree with our T&C and Privacy Notice</p>
          <Button className="sm:w-32 w-full" type="submit" disabled={loginMutation.isPending}>
            {loginMutation.isPending ? "Please wait..." : "Continue"}
          </Button>
        </section>
      </form>
    </Page>
  );
};

export default LoginPage;
