import { Sheet } from "../ui/sheet";
import { SidebarNav } from "./SidebarNav";
export function MobileNav({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (v: boolean) => void;
}) {
  return (
    <Sheet title="Menu principal" open={open} onOpenChange={onOpenChange}>
      <SidebarNav onNavigate={() => onOpenChange(false)} />
    </Sheet>
  );
}
