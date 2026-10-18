const money = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" });

export function formatMoney(cents: number): string {
  return money.format(cents / 100);
}

export function splitPrice(cents: number): { dollars: string; cents: string } {
  const whole = Math.floor(cents / 100);
  return {
    dollars: new Intl.NumberFormat("en-US").format(whole),
    cents: String(cents % 100).padStart(2, "0"),
  };
}

export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" });
}

export function describeAttributes(attrs: Record<string, string> | null | undefined): string {
  if (!attrs) return "";
  return Object.entries(attrs)
    .map(([k, v]) => `${k[0].toUpperCase()}${k.slice(1)}: ${v}`)
    .join(", ");
}
