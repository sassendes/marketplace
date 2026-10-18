import { useState } from "react";
import { BookOpen, Code2, Gamepad2, Package, Palette, ShieldCheck, Zap, type LucideIcon } from "lucide-react";

const icons: Record<string, LucideIcon> = {
  "developer-tools": Code2,
  productivity: Zap,
  security: ShieldCheck,
  design: Palette,
  games: Gamepad2,
  guides: BookOpen,
};

const tones: Record<string, string> = {
  "developer-tools": "bg-[#e6edf2] text-[#3c5a73]",
  productivity: "bg-[#f3eedf] text-[#7a5d12]",
  security: "bg-[#e5efed] text-[#2f6159]",
  design: "bg-[#f1e8ec] text-[#7a3d57]",
  games: "bg-[#e9eee6] text-[#476036]",
  guides: "bg-[#efe9e1] text-[#6b4a33]",
};

export function ProductImage({ src, category, alt, size = "md" }: { src?: string; category?: string; alt: string; size?: "sm" | "md" | "lg" }) {
  const [failed, setFailed] = useState(false);
  const iconSize = size === "lg" ? 120 : size === "md" ? 64 : 28;
  if (src && !failed) {
    return (
      <img
        src={src}
        alt={alt}
        loading="lazy"
        onError={() => setFailed(true)}
        className="h-full w-full object-contain"
      />
    );
  }
  const Icon = icons[category ?? ""] ?? Package;
  return (
    <div role="img" aria-label={alt} className={`flex h-full w-full items-center justify-center ${tones[category ?? ""] ?? "bg-[#eceef0] text-muted"}`}>
      <Icon size={iconSize} strokeWidth={1.25} aria-hidden />
    </div>
  );
}
