import * as Menu from "@radix-ui/react-dropdown-menu";
import type { ReactNode } from "react";
export function DropdownMenu({
  trigger,
  items,
}: {
  trigger: ReactNode;
  items: { label: string; action: () => void }[];
}) {
  return (
    <Menu.Root>
      <Menu.Trigger asChild>{trigger}</Menu.Trigger>
      <Menu.Portal>
        <Menu.Content className="dd-popover" sideOffset={8}>
          {items.map((item) => (
            <Menu.Item
              key={item.label}
              className="dd-menu-item"
              onSelect={item.action}
            >
              {item.label}
            </Menu.Item>
          ))}
        </Menu.Content>
      </Menu.Portal>
    </Menu.Root>
  );
}
