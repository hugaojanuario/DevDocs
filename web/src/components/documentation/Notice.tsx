import type { ReactNode } from "react";
import { CircleAlert, ShieldCheck, Sparkles, WifiOff } from "lucide-react";
import { Icon } from "../common/Icon";
export function Notice({
  tone = "partial",
  title,
  children,
  action,
}: {
  tone?: "info" | "partial" | "ai" | "error" | "offline";
  title: string;
  children?: ReactNode;
  action?: ReactNode;
}) {
  return (
    <div
      className={"dd-notice dd-notice-" + (tone === "offline" ? "error" : tone)}
      role={tone === "error" || tone === "offline" ? "alert" : "note"}
    >
      <Icon
        icon={
          tone === "ai"
            ? Sparkles
            : tone === "info"
              ? ShieldCheck
              : tone === "offline"
                ? WifiOff
                : CircleAlert
        }
      />
      <div>
        <span className="dd-notice-title">{title}</span>
        {children && <div className="dd-notice-body">{children}</div>}
        {action && <div className="dd-notice-action">{action}</div>}
      </div>
    </div>
  );
}
