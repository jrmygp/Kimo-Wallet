import * as yup from "yup";

export const userResponseSchema = yup.object({
  user: yup
    .object({
      id: yup.string().required(),
      phoneNumber: yup.string().required(),
      fullName: yup.string().required(),
      createdAt: yup.string().required(),
    })
    .required(),
  accessToken: yup.string().required(),
});

export type UserResponseData = yup.InferType<typeof userResponseSchema>;
