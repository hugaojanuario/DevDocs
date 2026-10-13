import * as Dialog from "@radix-ui/react-dialog";
import type { ReactNode } from "react";
export function Sheet({
  open,
  onOpenChange,
  title,
  children,
  fullWidth = false,
}: {
  open: boolean;
  onOpenChange: (value: boolean) => void;
  title: string;
  children: ReactNode;
  fullWidth?: boolean;
}) {
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="dd-overlay" />
        <Dialog.Content className={fullWidth ? "dd-sheet dd-sheet-search" : "dd-sheet"}>
          <Dialog.Title className="sr-only">{title}</Dialog.Title>
          <Dialog.Description className="sr-only">
            Navegue pelas páginas do DevDocs.
          </Dialog.Description>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
