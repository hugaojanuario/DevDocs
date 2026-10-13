import type { Config } from "tailwindcss";

const v = (name: string) => `var(--${name})`;

export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  darkMode: ["selector", '[data-theme="dark"]'],
  theme: {
    colors: {
      transparent: "transparent",
      current: "currentColor",
      surface: {
        DEFAULT: v("surface"),
        raised: v("surface-raised"),
        sunken: v("surface-sunken"),
        hover: v("surface-hover"),
      },
      line: { DEFAULT: v("line"), control: v("line-control") },
      ink: {
        DEFAULT: v("ink"),
        muted: v("ink-muted"),
        subtle: v("ink-subtle"),
      },
      accent: {
        DEFAULT: v("accent"),
        soft: v("accent-soft"),
        on: v("on-accent"),
      },
      ai: { DEFAULT: v("ai"), soft: v("ai-soft") },
      danger: {
        DEFAULT: v("danger"),
        soft: v("danger-soft"),
        on: v("on-danger"),
      },
      link: v("link"),
      focus: v("focus"),
      syn: {
        keyword: v("syn-keyword"),
        string: v("syn-string"),
        function: v("syn-function"),
        comment: v("syn-comment"),
        number: v("syn-number"),
        type: v("syn-type"),
        plain: v("syn-plain"),
        punct: v("syn-punct"),
      },
    },
    fontFamily: { sans: v("font-sans"), mono: v("font-mono") },
    fontSize: {
      display: [
        "40px",
        { lineHeight: "44px", letterSpacing: "-0.02em", fontWeight: "600" },
      ],
      title: [
        "28px",
        { lineHeight: "34px", letterSpacing: "-0.015em", fontWeight: "600" },
      ],
      heading: [
        "20px",
        { lineHeight: "28px", letterSpacing: "-0.01em", fontWeight: "600" },
      ],
      subheading: ["15px", { lineHeight: "22px", fontWeight: "600" }],
      body: ["15px", { lineHeight: "24px" }],
      "body-sm": ["13px", { lineHeight: "20px" }],
      label: [
        "12px",
        { lineHeight: "16px", letterSpacing: "0.01em", fontWeight: "500" },
      ],
      code: ["13px", { lineHeight: "21px" }],
      "code-sm": ["12px", { lineHeight: "18px", fontWeight: "500" }],
    },
    spacing: {
      0: "0",
      px: "1px",
      0.5: "2px",
      1: v("space-1"),
      2: v("space-2"),
      3: v("space-3"),
      4: v("space-4"),
      6: v("space-6"),
      8: v("space-8"),
      12: v("space-12"),
      16: v("space-16"),
      control: v("control-height"),
      header: v("header-height"),
      sidebar: v("sidebar-width"),
    },
    borderRadius: {
      none: "0",
      sm: v("radius-sm"),
      md: v("radius-md"),
      lg: v("radius-lg"),
      full: v("radius-full"),
    },
    boxShadow: {
      none: "none",
      popover: v("shadow-popover"),
      dialog: v("shadow-dialog"),
      focus: `0 0 0 2px ${v("surface")}, 0 0 0 4px ${v("focus")}`,
    },
    extend: {
      maxWidth: { content: v("content-width"), wide: v("wide-width") },
      width: { sidebar: v("sidebar-width") },
      transitionDuration: { DEFAULT: "120ms" },
    },
  },
  plugins: [],
} satisfies Config;
