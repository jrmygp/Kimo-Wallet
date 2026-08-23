import * as yup from "yup"

export const loginValidation = yup.object().shape({
  country: yup.string().required("This field is required"),
  number: yup
    .string()
    .required("This field is required")
    .matches(/^\d+$/, "Enter digits only")
    .min(5, "Number is too short")
    .max(13, "Number is too long"),
})

export type LoginFormValues = yup.InferType<typeof loginValidation>
