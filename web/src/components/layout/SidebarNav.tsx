import { NavLink } from "react-router-dom";
import {
  Search,
  Bookmark,
  Layers2,
  Settings,
  PanelLeftClose,
  PanelLeftOpen,
} from "lucide-react";
import { Icon } from "../common/Icon";
import { Tooltip } from "../ui/tooltip";
import { Button } from "../ui/button";
export function SidebarNav({
  collapsed = false,
  onToggle,
  onNavigate,
  total,
}: {
  collapsed?: boolean;
  onToggle?: () => void;
  onNavigate?: () => void;
  total?: number;
}) {
  const items = [
    { to: "/", name: "Pesquisa", icon: Search },
    { to: "/saved", name: "Consultas salvas", icon: Bookmark },
    { to: "/technologies", name: "Tecnologias", icon: Layers2 },
    { to: "/settings", name: "Configurações", icon: Settings },
  ];
  return (
    <nav className="dd-nav" aria-label="Principal" data-collapsed={collapsed}>
      {items.map((item) => {
        const link = (
          <NavLink
            to={item.to}
            end={item.to === "/" || item.to === "/saved"}
            onClick={onNavigate}
            aria-label={item.name}
          >
            <Icon icon={item.icon} />
            <span className="dd-nav-label">{item.name}</span>
            {item.to === "/saved" && total !== undefined && (
              <span className="dd-nav-count">{total}</span>
            )}
          </NavLink>
        );
        return collapsed ? (
          <Tooltip key={item.to} label={item.name}>
            {link}
          </Tooltip>
        ) : (
          <div key={item.to}>{link}</div>
        );
      })}
      {onToggle && (
        <div className="dd-nav-bottom">
          <Tooltip label={collapsed ? "Expandir menu" : "Recolher menu"}>
            <Button
              variant="ghost"
              iconOnly
              aria-label={collapsed ? "Expandir menu" : "Recolher menu"}
              onClick={onToggle}
            >
              <Icon icon={collapsed ? PanelLeftOpen : PanelLeftClose} />
            </Button>
          </Tooltip>
        </div>
      )}
    </nav>
  );
}
