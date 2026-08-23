import * as yup from "yup";

/**
 * Every response from the API Gateway is {statusCode, message, data}.
 * lib/axios.ts's response interceptor already unwraps axios's own
 * `response.data` wrapper, so a resolved request returns this envelope
 * directly — but axios's TypeScript types don't reflect that interceptor,
 * so the resolved value is typed `AxiosResponse<T>` while at runtime it's
 * actually this envelope. Treat it as `unknown` and validate with this
 * schema rather than trusting a cast.
 */
export function apiEnvelopeSchema<T extends yup.AnyObject>(dataSchema: yup.ObjectSchema<T>) {
  return yup.object({
    statusCode: yup.number().required(),
    message: yup.string().required(),
    data: dataSchema.required(),
  });
}
