"use client";

import Page from "@/components/layout/Page";
import kimo from "@/public/images/kimo.png";
import { useFormik } from "formik";
import Image from "next/image";
import { Button } from "@/components/ui/button";
import { CountryCodeSelect } from "@/components/country-code-select";
import { Input } from "@/components/ui/input";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { registerValidation } from "@/features/auth/schemas/register.schema";
import { useRegisterMutation } from "@/features/auth/hooks/use-register-mutation";
import { useAppDispatch } from "@/lib/store/hooks";
import { setUser } from "@/features/auth/store/user-slice";
import { useRouter } from "next/navigation";
import { all } from "country-codes-list";

const countries = all();

const RegisterPage = () => {
  const router = useRouter();
  const dispatch = useAppDispatch();
  const registerMutation = useRegisterMutation();
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
          onSuccess: (data) => {
            localStorage.removeItem("phoneNumber");
            localStorage.setItem("token", data.accessToken);
            dispatch(setUser(data.user));
            router.push("/home");
          },
        },
      );
    },
  });

  const countryInvalid = formik.touched.country && !!formik.errors.country;
  const numberInvalid = formik.touched.number && !!formik.errors.number;
  const fullNameInvalid = formik.touched.fullName && !!formik.errors.fullName;

  return (
    <Page>
      <form
        className="bg-kimo-500 flex min-h-full w-full flex-col justify-between py-10 px-2 sm:px-0"
        onSubmit={formik.handleSubmit}
      >
        <section className="flex flex-col w-full items-center justify-center gap-4">
          <Image src={kimo} alt="kimo-logo" className="w-60 sm:w-80" />
          <p className="text-xl font-medium text-white text-center">Welcome to Kimo!</p>
          <p className="text-lg text-white text-center">Make your new account and start your journey</p>

          <div className="flex flex-col">
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

            <Field data-invalid={fullNameInvalid}>
              <FieldLabel className={fullNameInvalid ? "text-red-500" : undefined}>Full Name</FieldLabel>
              <Input
                name="fullName"
                value={formik.values.fullName}
                onChange={formik.handleChange}
                onBlur={formik.handleBlur}
                placeholder="John Doe"
                aria-invalid={fullNameInvalid}
                className="bg-white"
              />
              <FieldDescription>{fullNameInvalid && formik.errors.fullName}</FieldDescription>
            </Field>
          </div>

          <section className="flex flex-col gap-2 items-center justify-center">
            <p className="text-white text-sm text-center">
              By continuing, you are agree with our T&C and Privacy Notice
            </p>
            <Button className="sm:w-32 w-full" type="submit" disabled={registerMutation.isPending}>
              {registerMutation.isPending ? "Please wait..." : "Continue"}
            </Button>
          </section>
        </section>
      </form>
    </Page>
  );
};

export default RegisterPage;
