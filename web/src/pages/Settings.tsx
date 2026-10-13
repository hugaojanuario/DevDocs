import { Moon, Sun, Monitor } from "lucide-react";
import { usePreferences } from "../lib/preferences";
import { useHealth } from "../hooks/queries";
import { Button } from "../components/ui/button";
import { Icon } from "../components/common/Icon";
import { Notice } from "../components/documentation/Notice";
export function SettingsPage() {
  const { theme, setTheme, official, setOfficial } = usePreferences();
  const system = useHealth();
  return (
    <div className="dd-doc">
      <h1 className="dd-page-title">Configurações</h1>
      <section className="dd-settings-section">
        <h2>Aparência</h2>
        <p>Escolha o tema da interface.</p>
        <div className="dd-segmented" role="group" aria-label="Tema">
          {(
            [
              { value: "dark", label: "Escuro", icon: Moon },
              { value: "light", label: "Claro", icon: Sun },
              { value: "system", label: "Sistema", icon: Monitor },
            ] as const
          ).map((item) => (
            <Button
              key={item.value}
              variant="ghost"
              aria-pressed={theme === item.value}
              onClick={() => setTheme(item.value)}
            >
              <Icon icon={item.icon} />
              {item.label}
            </Button>
          ))}
        </div>
      </section>
      <section className="dd-settings-section">
        <h2>Fontes</h2>
        <label className="dd-toggle">
          <input
            type="checkbox"
            checked={official}
            onChange={(e) => setOfficial(e.target.checked)}
          />
          Priorizar fontes oficiais
        </label>
        <p>A preferência fica apenas em memória nesta sessão.</p>
      </section>
      <section className="dd-settings-section">
        <h2>Inteligência artificial</h2>
        <label className="dd-toggle">
          <input type="checkbox" disabled />
          Permitir síntese por IA
        </label>
        <Notice tone="info" title="IA opcional">
          Disponível quando o backend habilitar síntese por IA.
        </Notice>
      </section>
      <section className="dd-settings-section">
        <h2>Sistema</h2>
        <div className="dd-system-status">
          <span
            className={
              "dd-status-dot " + (system.isSuccess ? "bg-accent" : "bg-danger")
            }
            aria-hidden
          />
          {system.isPending
            ? "Verificando…"
            : system.isSuccess
              ? "Conectado"
              : "Indisponível"}
        </div>
        <p>
          API v1.0.0 ·{" "}
          <a
            className="dd-link"
            href="/openapi.yaml"
            target="_blank"
            rel="noopener noreferrer"
          >
            Contrato OpenAPI<span className="sr-only"> (abre em nova aba)</span>
          </a>
        </p>
        <p>
          Pesquisas não são gravadas automaticamente. Apenas consultas que você
          salvar são persistidas.
        </p>
      </section>
    </div>
  );
}
