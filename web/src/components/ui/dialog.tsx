import * as Dialog from "@radix-ui/react-dialog";
import type { ComponentProps } from "react";
export const Root = Dialog.Root;
export const Title = Dialog.Title;
export const Description = Dialog.Description;
export const Close = Dialog.Close;
export const Trigger = Dialog.Trigger;
export function Content({
  children,
  ...props
}: ComponentProps<typeof Dialog.Content>) {
  return (
    <Dialog.Portal>
      <Dialog.Overlay className="dd-overlay" />
      <Dialog.Content className="dd-dialog dd-dialog-fixed" {...props}>
        {children}
      </Dialog.Content>
    </Dialog.Portal>
  );
}
