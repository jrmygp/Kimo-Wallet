"use client";

import type { ComponentProps, Ref } from "react";
import { CountryCodeSelect } from "@/components/country-code-select";
import { Input } from "@/components/ui/input";
import { Field, FieldLabel } from "@/components/ui/field";
import { FieldMessage } from "@/features/auth/components/form-feedback";

type PhoneNumberFieldProps = {
  id: string;
  country: string;
  onCountryChange: (countryCode: string) => void;
  invalid: boolean;
  error?: string;
  inputRef?: Ref<HTMLInputElement>;
  className?: string;
} & Pick<ComponentProps<"input">, "name" | "value" | "onChange" | "onBlur">;

// Country code and number presented as one control with one label and one error line,
// because to the user they are a single value: their mobile number.
export function PhoneNumberField({
  id,
  country,
  onCountryChange,
  invalid,
  error,
  inputRef,
  className,
  ...inputProps
}: PhoneNumberFieldProps) {
  const errorId = `${id}-error`;

  return (
    <Field data-invalid={invalid} className={className}>
      <FieldLabel htmlFor={id} className="text-[13px] text-foreground">
        Mobile number
      </FieldLabel>

      <div
        className={
          invalid
            ? "flex h-13 rounded-xl border border-destructive bg-card ring-3 ring-destructive/15"
            : "flex h-13 rounded-xl border border-input bg-card transition-[border-color,box-shadow] focus-within:border-ring focus-within:ring-3 focus-within:ring-ring/25"
        }
      >
        <CountryCodeSelect
          value={country}
          onChange={onCountryChange}
          placeholder="Country"
          className="h-auto shrink-0 rounded-none rounded-l-xl border-r border-border bg-transparent px-3 text-base text-foreground shadow-none focus-visible:bg-muted focus-visible:ring-0"
        />

        <Input
          {...inputProps}
          ref={inputRef}
          id={id}
          type="tel"
          inputMode="numeric"
          autoComplete="tel-national"
          placeholder="812 3456 7890"
          aria-invalid={invalid}
          aria-describedby={errorId}
          className="h-auto rounded-none rounded-r-xl border-0 bg-transparent px-3 text-base tabular-nums shadow-none focus-visible:ring-0 md:text-base"
        />
      </div>

      <FieldMessage id={errorId} error={invalid ? error : undefined} />
    </Field>
  );
}
