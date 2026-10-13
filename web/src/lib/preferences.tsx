import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { applyTheme, readThemePref, type ThemePref } from "./theme";
const Prefs = createContext<{
  theme: ThemePref;
  setTheme: (v: ThemePref) => void;
  official: boolean;
  setOfficial: (v: boolean) => void;
}>({
  theme: "dark",
  setTheme: () => {},
  official: false,
  setOfficial: () => {},
});
export function Preferences({ children }: { children: ReactNode }) {
  const [theme, setTheme] = useState(readThemePref);
  const [official, setOfficial] = useState(false);
  useEffect(() => {
    applyTheme(theme);
    const m = matchMedia("(prefers-color-scheme: light)");
    const update = () => {
      if (theme === "system") applyTheme(theme);
    };
    m.addEventListener("change", update);
    return () => m.removeEventListener("change", update);
  }, [theme]);
  return (
    <Prefs.Provider value={{ theme, setTheme, official, setOfficial }}>
      {children}
    </Prefs.Provider>
  );
}
export const usePreferences = () => useContext(Prefs);
