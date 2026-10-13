import type { ReactNode } from "react";
import { FileSearch, type LucideIcon } from "lucide-react";
import { Icon } from "./Icon";
export function EmptyState({
  title,
  description,
  action,
  icon = FileSearch,
  headingLevel = 2,
}: {
  title: string;
  description: string;
  action?: ReactNode;
  icon?: LucideIcon;
  headingLevel?: 1 | 2;
}) {
  const H = headingLevel === 1 ? "h1" : "h2";
  return (
    <div className="dd-empty">
      <div className="dd-empty-ic">
        <Icon icon={icon} size={20} />
      </div>
      <H>{title}</H>
      <p>{description}</p>
      {action}
    </div>
  );
}
