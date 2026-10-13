import { Toaster as Sonner } from "sonner";
import { Check, CircleAlert } from "lucide-react";
export const Toasts = () => (
  <Sonner
    position="bottom-right"
    duration={4000}
    closeButton
    icons={{
      success: <Check className="text-accent" size={16} />,
      error: <CircleAlert className="text-danger" size={16} />,
    }}
    toastOptions={{
      classNames: {
        toast: "dd-toast",
        description: "text-ink-muted",
        closeButton: "dd-toast-close",
      },
      unstyled: true,
    }}
  />
);
