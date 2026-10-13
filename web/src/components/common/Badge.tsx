import type { ReactNode } from "react";
import { ShieldCheck, Sparkles, CircleAlert } from "lucide-react";
export function Badge({
  tone = "neutral",
  children,
}: {
  tone?: "tech" | "official" | "deterministic" | "ai" | "neutral" | "danger";
  children: ReactNode;
}) {
  const I =
    tone === "official"
      ? ShieldCheck
      : tone === "ai"
        ? Sparkles
        : tone === "danger"
          ? CircleAlert
          : null;
  return (
    <span className={"dd-badge dd-badge-" + tone}>
      {I && <I size={12} strokeWidth={2} aria-hidden />}
      {children}
    </span>
  );
}
