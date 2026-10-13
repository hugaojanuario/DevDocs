import { useState, useEffect } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { Header } from "./Header";
import { SidebarNav } from "./SidebarNav";
import { SkipLink } from "./SkipLink";
import { defaultCollapsed, desktopMedia } from "../../lib/responsive";
export function AppShell() {
  const [collapsed, setCollapsed] = useState(defaultCollapsed);
  const location = useLocation();
  useEffect(() => {
    const media = desktopMedia();
    const update = () => setCollapsed(!media.matches);
    media.addEventListener("change", update);
    return () => media.removeEventListener("change", update);
  }, []);
  useEffect(() => {
    document.getElementById("conteudo")?.scrollTo(0, 0);
  }, [location.pathname, location.search]);
  return (
    <>
      <SkipLink />
      <Header />
      <div className="dd-app-body">
        <div className="dd-desktop-nav">
          <SidebarNav
            collapsed={collapsed}
            onToggle={() => setCollapsed((v) => !v)}
          />
        </div>
        <main id="conteudo" className="dd-main" tabIndex={-1}>
          <Outlet />
        </main>
      </div>
    </>
  );
}
