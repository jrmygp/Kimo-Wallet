import * as yup from "yup"

export const loginValidation = yup.object().shape({
  country: yup.string().required("This field is required"),
  number: yup.string().required("This field is required")
})