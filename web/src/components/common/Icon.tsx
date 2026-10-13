import type { LucideIcon } from "lucide-react";
export function Icon({
  icon: I,
  size = 16,
}: {
  icon: LucideIcon;
  size?: number;
}) {
  return <I size={size} strokeWidth={1.75} className="dd-icon" aria-hidden />;
}
