"use client";

import { use } from "react";
import Page from "@/components/layout/Page";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { useSearchUserQuery } from "@/features/wallet/hooks/use-search-user-query";
import { Input } from "@/components/ui/input";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { useFormik } from "formik";
import { Button } from "@/components/ui/button";
import { MdArrowBack } from "react-icons/md";
import Link from "next/link";

const TransferPage = ({ params }: { params: Promise<{ userId: string }> }) => {
  const { userId } = use(params);
  const { data, isFetching, isError, error } = useSearchUserQuery(userId);
  const formik = useFormik({
    initialValues: {
      amount: "",
      note: "",
    },
    validationSchema: null,
    onSubmit: (values) => {
      console.log(values);
    },
  });

  const amountInvalid = formik.touched.amount && !!formik.errors.amount;
  const noteInvalid = formik.touched.note && !!formik.errors.note;

  return (
    <Page>
      <div className="flex min-h-full w-full flex-col gap-8 items-center relative bg-[#F5F5F5]">
        <section className="bg-kimo-500 w-full px-4 h-40 flex py-4 items-center justify-center relative">
          <Link
            href="/home"
            aria-label="Back"
            className="flex items-center gap-2 absolute top-4 left-4 text-white"
          >
            <MdArrowBack size={24} />
            <p>Back</p>
          </Link>

          <p className="text-2xl font-medium text-white">Send to Friend</p>
        </section>

        <div className="absolute top-30 w-[80%] flex flex-col gap-4 bg-white rounded-md p-4">
          <div className="flex items-center gap-2">
            <Avatar size="2xl">
              <AvatarImage src={data?.profilePicture || "https://github.com/shadcn.png"} />
              <AvatarFallback>{data?.fullName.slice(0, 2).toUpperCase() ?? "CN"}</AvatarFallback>
            </Avatar>

            <div>
              {data && <p className="text-lg font-medium">{data.fullName}</p>}
              {data && <p className="text-sm">{data.phoneNumber}</p>}
            </div>
          </div>

          <form className="flex flex-col gap-2" onSubmit={formik.handleSubmit}>
            <Field data-invalid={amountInvalid}>
              <FieldLabel className={amountInvalid ? "text-red-500" : "text-black"}>Send Amount</FieldLabel>
              <Input
                name="amount"
                value={formik.values.amount}
                onChange={formik.handleChange}
                onBlur={formik.handleBlur}
                placeholder="Rp"
                inputMode="numeric"
                aria-invalid={amountInvalid}
                className="bg-white"
              />
              <FieldDescription>{amountInvalid && formik.errors.amount}</FieldDescription>
            </Field>

            <Field data-invalid={noteInvalid}>
              {/* <FieldLabel className={noteInvalid ? "text-red-500" : "text-black"}></FieldLabel> */}
              <Input
                name="note"
                value={formik.values.note}
                onChange={formik.handleChange}
                onBlur={formik.handleBlur}
                placeholder="Add a note here"
                aria-invalid={noteInvalid}
                className="bg-white"
              />
              <FieldDescription>{noteInvalid && formik.errors.note}</FieldDescription>
            </Field>

            <Button type="submit">Send Now</Button>
          </form>

          {isFetching && <p className="text-sm text-muted-foreground">Loading recipient...</p>}
          {isError && (
            <p role="alert" className="text-sm text-destructive">
              {error.message}
            </p>
          )}
        </div>
      </div>
    </Page>
  );
};

export default TransferPage;
