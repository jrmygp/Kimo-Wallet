import * as yup from "yup";

export const userResponseSchema = yup.object({
  user: yup
    .object({
      id: yup.string().required(),
      phoneNumber: yup.string().required(),
      fullName: yup.string().required(),
      createdAt: yup.string().required(),
      // Present but nullable — .nullable().defined() (not .required(),
      // which yup treats null the same as undefined and rejects) says
      // "this key is always there, and null is a valid value for it."
      profilePicture: yup.string().nullable().defined(),
      kimoId: yup.string().required(),
    })
    .required(),
  accessToken: yup.string().required(),
});

export type UserResponseData = yup.InferType<typeof userResponseSchema>;
