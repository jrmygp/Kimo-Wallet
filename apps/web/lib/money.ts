const CURRENCY_SYMBOLS: Record<string, string> = {
  IDR: "Rp",
};

const numberFormatter = new Intl.NumberFormat("id-ID", { maximumFractionDigits: 0 });

export function currencySymbol(currency: string): string {
  return CURRENCY_SYMBOLS[currency] ?? currency;
}

/**
 * Display-only formatting of an amount the server already computed — e.g. 150000 → "150.000".
 * Never do arithmetic on the result or feed it back into a request (docs/CLAUDE.md §3.3).
 */
export function formatAmount(amount: number): string {
  return numberFormatter.format(amount);
}

export function formatMoney(amount: number, currency: string): string {
  return `${currencySymbol(currency)} ${formatAmount(amount)}`;
}
