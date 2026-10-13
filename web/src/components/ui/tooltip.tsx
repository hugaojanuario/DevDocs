import * as Tip from "@radix-ui/react-tooltip";
import type { ReactNode } from "react";
export const TooltipProvider = Tip.Provider;
export function Tooltip({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) {
  return (
    <Tip.Root>
      <Tip.Trigger asChild>{children}</Tip.Trigger>
      <Tip.Portal>
        <Tip.Content className="dd-popover" sideOffset={8}>
          {label}
        </Tip.Content>
      </Tip.Portal>
    </Tip.Root>
  );
}
