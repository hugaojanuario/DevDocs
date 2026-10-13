import { useParams, Link } from "react-router-dom";
import { useSaved } from "../hooks/queries";
import { ResultView } from "../components/documentation/ResultView";
import { ResultSkeleton } from "../components/documentation/ResultSkeleton";
import { EmptyState } from "../components/common/EmptyState";
import { Button } from "../components/ui/button";
import { Notice } from "../components/documentation/Notice";
import { ApiError } from "../services/http";
export function SavedDetail() {
  const { id = "" } = useParams();
  const result = useSaved(id);
  if (result.isPending) return <ResultSkeleton />;
  if (result.error) {
    if (result.error instanceof ApiError && result.error.status === 404)
      return (
        <EmptyState
          headingLevel={1}
          title="Consulta não encontrada"
          description="Esta consulta pode ter sido excluída permanentemente."
          action={
            <Button asChild>
              <Link to="/saved">Voltar para consultas salvas</Link>
            </Button>
          }
        />
      );
    return (
      <Notice
        tone="error"
        title="Não foi possível abrir a consulta"
        action={
          <Button onClick={() => void result.refetch()}>
            Tentar novamente
          </Button>
        }
      >
        {result.error.message}
      </Notice>
    );
  }
  return <ResultView answer={result.data.response} saved={result.data} />;
}
