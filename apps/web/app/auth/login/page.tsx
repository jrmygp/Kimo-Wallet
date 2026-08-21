"use client";

import kimo from "@/public/images/kimo.png";
import Image from "next/image";
import Page from "@/components/layout/Page";
import { Button } from "@/components/ui/button";
import { CountryCodeSelect } from "@/components/country-code-select";
import { Input } from "@/components/ui/input";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { useFormik } from "formik";
import { loginValidation } from "@/features/transaction/schemas/login.schema";
import { useRouter } from "next/navigation";

const LoginPage = () => {
  const router = useRouter();

  const formik = useFormik({
    initialValues: {
      country: "ID",
      number: "",
    },
    validationSchema: loginValidation,
    onSubmit: (values) => {
      router.push("/home")
    },
  });

  const countryInvalid = formik.touched.country && !!formik.errors.country;
  const numberInvalid = formik.touched.number && !!formik.errors.number;

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
        </section>

        <section className="flex flex-col gap-2 items-center justify-center">
          <p className="text-white text-sm text-center">By continuing, you are agree with our T&C and Privacy Notice</p>
          <Button className="sm:w-32 w-full" type="submit">
            Continue
          </Button>
        </section>
      </form>
    </Page>
  );
};

export default LoginPage;
