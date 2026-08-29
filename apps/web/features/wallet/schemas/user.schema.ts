import * as yup from "yup";

/**
 * The User shape as returned by GET /v1/user/{kimoId} (api-gateway's
 * GetUserByID, wrapping user-service's GetUserByID RPC — see
 * packages/contracts/user/v1/user.proto). Deliberately separate from
 * features/auth/schemas/user-response.schema.ts's userResponseSchema:
 * that one is the login/register response contract (User + accessToken),
 * this one is GetUserByID's (User only): different endpoints, different
 * contracts, even though the User fields happen to match today.
 */
export const userSchema = yup.object({
  id: yup.string().required(),
  phoneNumber: yup.string().required(),
  fullName: yup.string().required(),
  createdAt: yup.string().required(),
  // Present but nullable in the response — .nullable().defined() (not
  // .required(), which yup treats null the same as undefined and rejects)
  // is what actually says "this key is always there, and null is a valid
  // value for it."
  profilePicture: yup.string().nullable().defined(),
  // The public-facing identifier — what GET /v1/user/{kimoId} is actually
  // looked up by now (see the endpoint's doc comment above), not `id`.
  kimoId: yup.string().required(),
});

export type WalletUser = yup.InferType<typeof userSchema>;
