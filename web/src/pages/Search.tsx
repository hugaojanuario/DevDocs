import { useEffect, useState } from "react";
import { Navigate, useSearchParams } from "react-router-dom";
import { ResultView } from "../components/documentation/ResultView";
import { ResultSkeleton } from "../components/documentation/ResultSkeleton";
import { Notice } from "../components/documentation/Notice";
import { EmptyState } from "../components/common/EmptyState";
import { Button } from "../components/ui/button";
import { useSearch } from "../hooks/queries";
import { usePreferences } from "../lib/preferences";
import { ApiError } from "../services/http";
export function SearchPage() {
  const [params] = useSearchParams();
  const q = (params.get("q") ?? "").trim();
  const tech = params.get("tech") ?? "";
  const { official, ai } = usePreferences();
  const result = useSearch(q, tech, official, ai);
  const [slow, setSlow] = useState(false);
  useEffect(() => {
    setSlow(false);
    if (!result.isFetching) return;
    const timer = setTimeout(() => setSlow(true), 8000);
    return () => clearTimeout(timer);
  }, [result.isFetching, q, tech]);
  if (!q) return <Navigate to="/" replace />;
  const retry = (
    <Button size="sm" onClick={() => void result.refetch()}>
      Tentar novamente
    </Button>
  );
  if (result.isPending)
    return (
      <div className="dd-doc">
        {slow && (
          <Notice
            tone="error"
            title="A busca demorou mais que o esperado."
            action={retry}
          >
            A requisição continua em andamento por até 45 segundos.
          </Notice>
        )}
        <ResultSkeleton />
      </div>
    );
  if (result.error) {
    const e = result.error;
    const code = e instanceof ApiError ? e.code : "network_error";
    const status = e instanceof ApiError ? e.status : 0;
    if (status === 404)
      return (
        <EmptyState
          headingLevel={1}
          title="Nenhuma documentação encontrada"
          description={e.message}
          action={
            <Button asChild>
              <a href="/">Nova pesquisa</a>
            </Button>
          }
        />
      );
    return (
      <div className="dd-doc">
        <h1 className="dd-page-title">Resultado da pesquisa</h1>
        <Notice
          tone={code === "network_error" ? "offline" : "error"}
          title={
            status === 503
              ? "As fontes de documentação estão indisponíveis no momento."
              : code === "timeout"
                ? "A busca demorou mais que o esperado."
                : "Não foi possível buscar documentação"
          }
          action={retry}
        >
          {e.message}
        </Notice>
      </div>
    );
  }
  return (
    <ResultView
      key={q + "|" + tech + "|" + official + "|" + ai}
      answer={result.data}
    />
  );
}
