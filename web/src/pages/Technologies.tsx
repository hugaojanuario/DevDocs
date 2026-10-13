import { useNavigate } from "react-router-dom";
import { ExternalLink, Search } from "lucide-react";
import { useTechnologies, useSources } from "../hooks/queries";
import { Button } from "../components/ui/button";
import { Icon } from "../components/common/Icon";
import { EmptyState } from "../components/common/EmptyState";
import { Notice } from "../components/documentation/Notice";
import { Skeleton } from "../components/ui/skeleton";
import { safeURL } from "../lib/utils";
export function TechnologiesPage() {
  const technologies = useTechnologies();
  const sources = useSources();
  const navigate = useNavigate();
  return (
    <div className="dd-wide">
      <h1 className="dd-page-title">Tecnologias</h1>
      <p className="dd-page-description">
        Explore as fontes de documentação cadastradas no DevDocs.
      </p>
      {technologies.isPending ? (
        <Skeleton />
      ) : technologies.error ? (
        <Notice
          tone="error"
          title="Não foi possível carregar tecnologias"
          action={
            <Button onClick={() => void technologies.refetch()}>
              Tentar novamente
            </Button>
          }
        >
          {technologies.error.message}
        </Notice>
      ) : technologies.data.length ? (
        <div className="dd-tech-grid">
          {technologies.data.map((t) => {
            const source = sources.data?.find(
              (s) => s.technology === t.id && s.official === true,
            );
            const url = source
              ? safeURL("https://" + source.domain + source.path_prefix)
              : null;
            return (
              <article key={t.id} className="dd-tech">
                <div className="dd-tech-head">
                  <h2>{t.name}</h2>
                </div>
                <div className="dd-tech-foot">
                  {url && (
                    <a
                      className="dd-tech-link"
                      href={url}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      {source?.domain}
                      <Icon icon={ExternalLink} size={12} />
                      <span className="sr-only">(abre em nova aba)</span>
                    </a>
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() =>
                      navigate("/?" + new URLSearchParams({ tech: t.id }))
                    }
                  >
                    <Icon icon={Search} size={14} />
                    Pesquisar em {t.name}
                  </Button>
                </div>
              </article>
            );
          })}
        </div>
      ) : (
        <EmptyState
          title="Nenhuma tecnologia cadastrada"
          description="O backend ainda não retornou tecnologias."
        />
      )}
      {sources.error && (
        <div className="mt-6">
          <Notice tone="partial" title="Links das fontes indisponíveis" />
        </div>
      )}
    </div>
  );
}
