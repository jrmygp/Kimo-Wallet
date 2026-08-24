import * as yup from "yup";

export const registerValidation = yup.object().shape({
  country: yup.string().required("This field is required"),
  number: yup
    .string()
    .required("This field is required")
    .matches(/^\d+$/, "Enter digits only")
    .min(5, "Number is too short")
    .max(13, "Number is too long"),
  fullName: yup
    .string()
    .required("This field is required")
    .min(10, "Minimum 10 characters and maximum 50 characters")
    .max(50, "Minimum 10 characters and maximum 50 characters"),
});
