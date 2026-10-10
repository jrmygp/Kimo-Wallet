import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { parseAmountInput } from "@/lib/money";

describe("parseAmountInput", () => {
  it("parses whole rupiah", () => {
    assert.equal(parseAmountInput("50000"), 50000);
    assert.equal(parseAmountInput("1"), 1);
  });

  it("trims surrounding whitespace", () => {
    assert.equal(parseAmountInput(" 50000 "), 50000);
  });

  for (const input of ["", "   ", "0.5", "1.5", "-5000", "+5000", "1e3", "50.000", "50,000", "abc", "Rp5000"]) {
    it(`rejects ${JSON.stringify(input)}`, () => {
      assert.equal(parseAmountInput(input), null);
    });
  }

  it("rejects amounts beyond the safe-integer range instead of rounding them", () => {
    assert.equal(parseAmountInput("9007199254740993"), null);
  });
});
