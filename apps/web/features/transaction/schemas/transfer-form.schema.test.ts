import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { ValidationError } from "yup";
import { createTransferFormSchema, MIN_TRANSFER_AMOUNT } from "@/features/transaction/schemas/transfer-form.schema";

const BALANCE = 100_000;

async function amountError(amount: string, balance: number | null = BALANCE): Promise<string | null> {
  try {
    await createTransferFormSchema({ balance }).validateAt("amount", { amount, note: "" });
    return null;
  } catch (error) {
    assert.ok(error instanceof ValidationError, `expected a ValidationError, got ${String(error)}`);
    return error.message;
  }
}

describe("transfer form schema — amount", () => {
  it("accepts an amount within the balance", async () => {
    assert.equal(await amountError("50000"), null);
  });

  it("accepts exactly the balance", async () => {
    assert.equal(await amountError(String(BALANCE)), null);
  });

  it("accepts exactly the minimum", async () => {
    assert.equal(await amountError(String(MIN_TRANSFER_AMOUNT)), null);
  });

  it("rejects one rupiah below the minimum", async () => {
    assert.match((await amountError(String(MIN_TRANSFER_AMOUNT - 1))) ?? "", /minimum/i);
  });

  it("accepts surrounding whitespace", async () => {
    assert.equal(await amountError(" 50000 "), null);
  });

  it("rejects empty with a readable message, not yup's NaN message", async () => {
    assert.equal(await amountError(""), "Enter an amount");
  });

  it("rejects zero", async () => {
    assert.match((await amountError("0")) ?? "", /minimum/i);
  });

  it("rejects negative amounts", async () => {
    assert.match((await amountError("-5000")) ?? "", /digits only/i);
  });

  for (const input of ["1.5", "1e3", "50.000", "50,000", "abc"]) {
    it(`rejects non-integer / non-numeric ${JSON.stringify(input)}`, async () => {
      assert.match((await amountError(input)) ?? "", /digits only/i);
    });
  }

  it("rejects an amount above the balance", async () => {
    assert.match((await amountError(String(BALANCE + 1))) ?? "", /balance/i);
  });

  it("rejects an amount above the safe-integer maximum", async () => {
    assert.match((await amountError("9007199254740993", null)) ?? "", /too large/i);
  });

  it("skips the balance check when the balance is unknown (the server decides)", async () => {
    assert.equal(await amountError("5000000", null), null);
  });
});

describe("transfer form schema — note", () => {
  it("rejects a note over 100 characters", async () => {
    await assert.rejects(
      createTransferFormSchema({ balance: BALANCE }).validateAt("note", {
        amount: String(MIN_TRANSFER_AMOUNT),
        note: "x".repeat(101),
      }),
    );
  });
});
