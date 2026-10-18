import { AlertTriangle, CheckCircle2 } from "lucide-react";
import type { ReactNode } from "react";

export function Alert({ tone = "bad", children }: { tone?: "bad" | "ok"; children: ReactNode }) {
  const Icon = tone === "bad" ? AlertTriangle : CheckCircle2;
  const color = tone === "bad" ? "border-bad/40 bg-[#fdf3f2] text-bad" : "border-ok/40 bg-[#f1f8f3] text-ok";
  return (
    <div role={tone === "bad" ? "alert" : "status"} className={`flex items-start gap-2 rounded-[var(--radius-panel)] border px-3 py-2 text-sm ${color}`}>
      <Icon size={18} className="mt-px shrink-0" aria-hidden />
      <div className="text-ink">{children}</div>
    </div>
  );
}

export function Loading({ label = "Loading" }: { label?: string }) {
  return (
    <div className="flex items-center gap-3 py-12 text-sm text-muted" role="status">
      <span className="h-4 w-4 animate-spin rounded-full border-2 border-line border-t-teal" aria-hidden />
      {label}
    </div>
  );
}

export function SkeletonGrid({ count = 8 }: { count?: number }) {
  return (
    <ul className="grid grid-cols-2 gap-px overflow-hidden rounded-[var(--radius-panel)] border border-line bg-line sm:grid-cols-3 lg:grid-cols-4" aria-hidden>
      {Array.from({ length: count }).map((_, i) => (
        <li key={i} className="bg-white p-3">
          <div className="aspect-square animate-pulse rounded-[3px] bg-page" />
          <div className="mt-3 h-4 w-4/5 animate-pulse rounded bg-page" />
          <div className="mt-2 h-4 w-1/3 animate-pulse rounded bg-page" />
        </li>
      ))}
    </ul>
  );
}
