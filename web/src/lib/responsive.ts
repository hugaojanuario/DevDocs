// Breakpoints from section 7 of the design handoff.
export const defaultCollapsed = () => window.innerWidth < 1024;
export const desktopMedia = () => window.matchMedia('(min-width: 1024px)');
