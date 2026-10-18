import { splitPrice } from "../format";

export function Price({ cents, size = "md", prefix }: { cents: number; size?: "sm" | "md" | "lg"; prefix?: string }) {
  const { dollars, cents: c } = splitPrice(cents);
  const big = size === "lg" ? "text-[28px]" : size === "md" ? "text-[21px]" : "text-[17px]";
  const small = size === "lg" ? "text-[13px] mt-[5px]" : "text-[11px] mt-[3px]";
  return (
    <span className="tnum inline-flex items-start leading-none text-ink" aria-label={`${prefix ?? ""}$${dollars}.${c}`}>
      {prefix && <span className="mr-1 mt-[3px] text-xs text-muted">{prefix}</span>}
      <span className={`${small} font-medium`}>$</span>
      <span className={`${big} font-semibold tracking-tight`}>{dollars}</span>
      <span className={`${small} font-medium`}>{c}</span>
    </span>
  );
}
