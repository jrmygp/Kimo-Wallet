import * as yup from "yup";
import { walletSchema } from "@/features/wallet/schemas/wallet.schema";

const userSchema = yup
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
  .required();

// login (via POST /v1/auth/login) is the only response that carries an
// access token and a wallet — see registerResponseSchema below for why
// register's shape is deliberately different.
export const userResponseSchema = yup.object({
  user: userSchema,
  wallet: walletSchema,
  accessToken: yup.string().required(),
});

export type UserResponseData = yup.InferType<typeof userResponseSchema>;

// register (via POST /v1/auth/register) deliberately returns only the
// created user — no access token, no wallet. Registering creates the
// account but does not start a session; the client must call Login
// separately with the same phone number (see apps/api-gateway's
// Register handler doc comment for the full reasoning, including why
// this also sidesteps the wallet-provisioning race).
export const registerResponseSchema = yup.object({
  user: userSchema,
});

export type RegisterResponseData = yup.InferType<typeof registerResponseSchema>;
