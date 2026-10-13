import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Bookmark } from "lucide-react";
import { useSavedList, useTechnologies } from "../hooks/queries";
import { SavedList } from "../components/saved/SavedList";
import { EmptyState } from "../components/common/EmptyState";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Notice } from "../components/documentation/Notice";
import { Skeleton } from "../components/ui/skeleton";
export function SavedPage() {
  const [q, setQ] = useState("");
  const [tech, setTech] = useState("");
  const [page, setPage] = useState(1);
  const [order, setOrder] = useState("recent");
  const technologies = useTechnologies();
  const params = new URLSearchParams({
    page: String(page),
    page_size: "20",
    q,
    technology: tech,
    order,
  });
  const result = useSavedList(params);
  const navigate = useNavigate();
  const filtered = !!q || !!tech;
  return (
    <div className="dd-wide">
      <div className="dd-page-heading">
        <h1 className="dd-page-title">Consultas salvas</h1>
        {result.data && (
          <span className="text-ink-subtle text-body-sm">
            {result.data.total} consultas
          </span>
        )}
      </div>
      <div className="dd-saved-toolbar">
        <label className="sr-only" htmlFor="filter-saved">
          Filtrar consultas salvas
        </label>
        <Input
          id="filter-saved"
          type="search"
          autoComplete="off"
          placeholder="Filtrar por título ou pergunta"
          value={q}
          onChange={(e) => {
            setQ(e.target.value);
            setPage(1);
          }}
          maxLength={600}
        />
        <label className="sr-only" htmlFor="filter-technology">
          Filtrar tecnologia
        </label>
        <select
          id="filter-technology"
          className="dd-input"
          value={tech}
          onChange={(e) => {
            setTech(e.target.value);
            setPage(1);
          }}
        >
          <option value="">Todas as tecnologias</option>
          {technologies.data?.map((t) => (
            <option key={t.id} value={t.id}>
              {t.name}
            </option>
          ))}
        </select>
        <label className="sr-only" htmlFor="sort-saved">
          Ordenação
        </label>
        <select
          id="sort-saved"
          className="dd-input"
          value={order}
          onChange={(e) => {
            setOrder(e.target.value);
            setPage(1);
          }}
        >
          <option value="recent">Mais recentes</option>
          <option value="oldest">Mais antigas</option>
          <option value="title">Título A–Z</option>
        </select>
      </div>
      {result.isPending ? (
        <div
          className="dd-saved-loading"
          aria-busy="true"
          aria-label="Carregando consultas salvas"
        >
          {Array.from({ length: 5 }, (_, i) => (
            <Skeleton key={i} />
          ))}
        </div>
      ) : result.error ? (
        <Notice
          tone="error"
          title="Não foi possível carregar consultas salvas"
          action={
            <Button onClick={() => void result.refetch()}>
              Tentar novamente
            </Button>
          }
        >
          {result.error.message}
        </Notice>
      ) : result.data.items.length ? (
        <>
          <SavedList items={result.data.items} />
          <div className="dd-pagination">
            <Button disabled={page === 1} onClick={() => setPage((v) => v - 1)}>
              Anterior
            </Button>
            <span>
              Página {page} de {Math.max(1, Math.ceil(result.data.total / 20))}
            </span>
            <Button
              disabled={page * 20 >= result.data.total}
              onClick={() => setPage((v) => v + 1)}
            >
              Próxima
            </Button>
          </div>
        </>
      ) : (
        <EmptyState
          icon={Bookmark}
          title={
            filtered
              ? "Nenhuma consulta corresponde ao filtro"
              : "Você ainda não salvou consultas"
          }
          description={
            filtered
              ? "Ajuste o texto ou a tecnologia selecionada."
              : "Ao encontrar uma resposta útil, use Salvar consulta para guardá-la aqui. O DevDocs não guarda seu histórico de pesquisas."
          }
          action={
            <Button
              variant={filtered ? "secondary" : "primary"}
              onClick={() => {
                if (filtered) {
                  setQ("");
                  setTech("");
                  setPage(1);
                } else navigate("/");
              }}
            >
              {filtered ? "Limpar filtros" : "Fazer uma pesquisa"}
            </Button>
          }
        />
      )}
    </div>
  );
}
