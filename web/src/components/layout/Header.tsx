import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { Menu, Search, Moon, Sun, Monitor, LogOut } from "lucide-react";
import { Brand } from "./Brand";
import { MobileNav } from "./MobileNav";
import { SearchField } from "../search/SearchField";
import { Button } from "../ui/button";
import { Icon } from "../common/Icon";
import { DropdownMenu } from "../ui/dropdown-menu";
import { usePreferences } from "../../lib/preferences";
import { logout } from "../../services/http";
import { Sheet } from "../ui/sheet";
export function Header() {
  const location = useLocation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { theme, setTheme } = usePreferences();
  const [menu, setMenu] = useState(false);
  const [search, setSearch] = useState(false);
  const [signingOut, setSigningOut] = useState(false);
  const [error, setError] = useState("");
  const q = new URLSearchParams(location.search).get("q") ?? "";
  return (
    <header className="dd-header">
      <div className="flex items-center gap-3">
        <Button
          className="dd-mobile-control"
          variant="ghost"
          iconOnly
          aria-label="Abrir menu"
          onClick={() => setMenu(true)}
        >
          <Icon icon={Menu} />
        </Button>
        <Brand />
      </div>
      {location.pathname !== "/" && (
        <>
          <div className="dd-header-search">
            <SearchField compact initial={q} />
          </div>
          <Button
            className="dd-mobile-control"
            variant="ghost"
            iconOnly
            aria-label="Abrir pesquisa"
            onClick={() => setSearch(true)}
          >
            <Icon icon={Search} />
          </Button>
        </>
      )}
      <div className="dd-header-actions">
        <DropdownMenu
          trigger={
            <Button variant="ghost" iconOnly aria-label="Alterar tema">
              <Icon
                icon={
                  theme === "light" ? Sun : theme === "system" ? Monitor : Moon
                }
              />
            </Button>
          }
          items={[
            { label: "Escuro", action: () => setTheme("dark") },
            { label: "Claro", action: () => setTheme("light") },
            { label: "Sistema", action: () => setTheme("system") },
          ]}
        />
        <Button
          variant="ghost"
          size="sm"
          loading={signingOut}
          onClick={async () => {
            setSigningOut(true);
            try {
              await logout();
              await queryClient.cancelQueries();
              queryClient.clear();
              navigate("/login", { replace: true });
            } catch {
              setError("Falha ao sair. Tente novamente.");
            } finally {
              setSigningOut(false);
            }
          }}
        >
          <Icon icon={LogOut} size={14} />
          Sair
        </Button>
        {error && (
          <span role="alert" className="text-danger text-label">
            {error}
          </span>
        )}
      </div>
      <MobileNav open={menu} onOpenChange={setMenu} />
      <Sheet fullWidth title="Nova pesquisa" open={search} onOpenChange={setSearch}>
        <div className="p-4">
          <SearchField
            autoFocus
            initial={q}
            onSubmit={(value) => {
              setSearch(false);
              navigate("/search?" + new URLSearchParams({ q: value }));
            }}
          />
        </div>
      </Sheet>
    </header>
  );
}
