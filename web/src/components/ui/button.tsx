import { forwardRef, type ButtonHTMLAttributes } from "react";
import { Slot, Slottable } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { RotateCw } from "lucide-react";
import { cn } from "../../lib/utils";
export const buttonVariants = cva("dd-btn", {
  variants: {
    variant: {
      primary: "dd-btn-primary",
      secondary: "dd-btn-secondary",
      ghost: "dd-btn-ghost",
      danger: "dd-btn-danger",
      "danger-ghost": "dd-btn-danger-ghost",
    },
    size: { md: "", sm: "dd-btn-sm" },
    iconOnly: { true: "dd-btn-icon", false: "" },
  },
  defaultVariants: { variant: "secondary", size: "md", iconOnly: false },
});
type Props = ButtonHTMLAttributes<HTMLButtonElement> &
  VariantProps<typeof buttonVariants> & {
    loading?: boolean;
    asChild?: boolean;
  };
export const Button = forwardRef<HTMLButtonElement, Props>(function Button(
  {
    className,
    variant,
    size,
    iconOnly,
    loading,
    asChild,
    disabled,
    children,
    ...props
  },
  ref,
) {
  const C = asChild ? Slot : "button";
  return (
    <C
      ref={ref}
      className={cn(buttonVariants({ variant, size, iconOnly }), className)}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      {...props}
    >
      {loading && (
        <RotateCw className="dd-icon dd-spin" size={16} aria-hidden />
      )}
      {asChild ? <Slottable>{children}</Slottable> : children}
    </C>
  );
});
