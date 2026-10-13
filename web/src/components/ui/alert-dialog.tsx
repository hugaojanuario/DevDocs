import * as Alert from "@radix-ui/react-alert-dialog";
import type { ComponentProps } from "react";
export const Root = Alert.Root;
export const Title = Alert.Title;
export const Description = Alert.Description;
export const Cancel = Alert.Cancel;
export const Action = Alert.Action;
export function Content({
  children,
  ...props
}: ComponentProps<typeof Alert.Content>) {
  return (
    <Alert.Portal>
      <Alert.Overlay className="dd-overlay" />
      <Alert.Content className="dd-dialog dd-dialog-fixed" {...props}>
        {children}
      </Alert.Content>
    </Alert.Portal>
  );
}
