import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { intentFor, type TransferIntent } from "@/features/transaction/transfer-intent";

const RECEIVER = "33333333-3333-4333-8333-333333333333";

function keyMinter() {
  let count = 0;
  return () => `key-${++count}`;
}

describe("intentFor", () => {
  it("mints a key for the first intent", () => {
    const intent = intentFor(null, { receiverUserId: RECEIVER, amount: 50000, currency: "IDR" }, keyMinter());
    assert.equal(intent.idempotencyKey, "key-1");
  });

  it("reuses the same key when the same transfer is reviewed again (retry / back-and-forth)", () => {
    const mint = keyMinter();
    const first = intentFor(null, { receiverUserId: RECEIVER, amount: 50000, currency: "IDR" }, mint);
    const again = intentFor(first, { receiverUserId: RECEIVER, amount: 50000, currency: "IDR" }, mint);
    assert.equal(again, first);
    assert.equal(again.idempotencyKey, "key-1");
  });

  it("mints a new key when the amount changes", () => {
    const mint = keyMinter();
    const first: TransferIntent = intentFor(null, { receiverUserId: RECEIVER, amount: 50000, currency: "IDR" }, mint);
    const changed = intentFor(first, { receiverUserId: RECEIVER, amount: 60000, currency: "IDR" }, mint);
    assert.equal(changed.idempotencyKey, "key-2");
  });

  it("mints a new key when the recipient changes", () => {
    const mint = keyMinter();
    const first = intentFor(null, { receiverUserId: RECEIVER, amount: 50000, currency: "IDR" }, mint);
    const changed = intentFor(
      first,
      { receiverUserId: "55555555-5555-4555-8555-555555555555", amount: 50000, currency: "IDR" },
      mint,
    );
    assert.equal(changed.idempotencyKey, "key-2");
  });

  it("mints a new key after the previous intent was completed (cleared to null)", () => {
    const mint = keyMinter();
    intentFor(null, { receiverUserId: RECEIVER, amount: 50000, currency: "IDR" }, mint);
    const next = intentFor(null, { receiverUserId: RECEIVER, amount: 50000, currency: "IDR" }, mint);
    assert.equal(next.idempotencyKey, "key-2");
  });
});
