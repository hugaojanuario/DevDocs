import type { ReactNode } from "react";
export function Kbd({ children }: { children: ReactNode }) {
  return (
    <kbd className="dd-kbd" aria-hidden>
      {children}
    </kbd>
  );
}
export const modifier = () =>
  /Mac|iPhone|iPad/.test(navigator.platform) ? "⌘" : "Ctrl";
