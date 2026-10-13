export type ThemePref = "dark" | "light" | "system";
const KEY = "devdocs.theme";
export function readThemePref(): ThemePref {
  try {
    const v = localStorage.getItem(KEY);
    return v === "light" || v === "system" ? v : "dark";
  } catch {
    return "dark";
  }
}
export function applyTheme(pref: ThemePref) {
  const sys = window.matchMedia("(prefers-color-scheme: light)").matches
    ? "light"
    : "dark";
  document.documentElement.dataset.theme = pref === "system" ? sys : pref;
  try {
    localStorage.setItem(KEY, pref);
  } catch {}
}
