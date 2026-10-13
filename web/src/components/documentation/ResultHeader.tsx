import { useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import {
  Bookmark,
  BookmarkCheck,
  Link,
  Search,
  Pencil,
  Trash2,
  RotateCw,
  CircleAlert,
} from "lucide-react";
import { toast } from "sonner";
import type { Answer, SavedQuery } from "../../types/models";
import { Badge } from "../common/Badge";
import { Icon } from "../common/Icon";
import { Button } from "../ui/button";
import { Tooltip } from "../ui/tooltip";
import { ConfirmDeleteDialog } from "../saved/ConfirmDeleteDialog";
import { RenameDialog } from "../saved/RenameDialog";
import { useCreateSaved } from "../../hooks/queries";
import { formatDate } from "../../lib/utils";
export function ResultHeader({
  answer,
  saved,
}: {
  answer: Answer;
  saved?: SavedQuery;
}) {
  const navigate = useNavigate();
  const create = useCreateSaved();
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [renameOpen, setRenameOpen] = useState(false);
  const [savedId, setSavedId] = useState("");
  const isAI = answer.mode === "ai_assisted";
  function save() {
    create.mutate(answer, {
      onSuccess: (v) => {
        setSavedId(v.id);
        toast.success("Consulta salva", {
          action: {
            label: "Ver consulta",
            onClick: () => navigate("/saved/" + v.id),
          },
        });
      },
      onError: () =>
        toast.error("Falha ao salvar consulta", { duration: Infinity }),
    });
  }
  async function copyLink() {
    try {
      await navigator.clipboard.writeText(window.location.href);
      toast.success("Link copiado");
    } catch {
      toast.error("Falha ao copiar link", { duration: Infinity });
    }
  }
  return (
    <header className="dd-rh">
      <div className="dd-rh-meta">
        {answer.technology && <Badge tone="tech">{answer.technology}</Badge>}
        {answer.references.some((r) => r.official === true) && (
          <Badge tone="official">Fonte oficial</Badge>
        )}
        <Badge tone={isAI ? "ai" : "deterministic"}>
          {isAI ? "Síntese por IA" : "Determinístico"}
        </Badge>
        {saved && <Badge>Salvo em {formatDate(saved.created_at)}</Badge>}
      </div>
      <h1>{saved?.title ?? answer.title}</h1>
      <p className="dd-rh-q">Pergunta: {answer.query}</p>
      <div className="dd-rh-actions">
        {saved ? (
          <>
            <Button
              onClick={() =>
                navigate(
                  "/search?" +
                    new URLSearchParams({
                      q: answer.query,
                      ...(answer.technology ? { tech: answer.technology } : {}),
                    }),
                )
              }
            >
              <Icon icon={RotateCw} />
              Pesquisar novamente
            </Button>
            <Button variant="ghost" onClick={() => setRenameOpen(true)}>
              <Icon icon={Pencil} />
              Renomear
            </Button>
            <Button variant="danger-ghost" onClick={() => setDeleteOpen(true)}>
              <Icon icon={Trash2} />
              Excluir
            </Button>
          </>
        ) : (
          <>
            <Button
              variant={savedId ? "secondary" : "primary"}
              disabled={!!savedId}
              loading={create.isPending}
              onClick={save}
            >
              <Icon
                icon={
                  savedId
                    ? BookmarkCheck
                    : create.error
                      ? CircleAlert
                      : Bookmark
                }
              />
              {savedId
                ? "Salvo"
                : create.isPending
                  ? "Salvando…"
                  : create.error
                    ? "Tentar salvar de novo"
                    : "Salvar consulta"}
            </Button>
            <Tooltip label="O link contém só a pergunta, não a resposta salva.">
              <Button variant="ghost" onClick={copyLink}>
                <Icon icon={Link} />
                Copiar link
              </Button>
            </Tooltip>
            <Button variant="ghost" onClick={() => navigate("/")}>
              <Icon icon={Search} />
              Nova pesquisa
            </Button>
            {savedId && (
              <RouterLink className="dd-link" to={"/saved/" + savedId}>
                Ver em Consultas salvas
              </RouterLink>
            )}
          </>
        )}
      </div>
      <span className="sr-only" aria-live="polite">
        {create.isPending
          ? "Salvando consulta"
          : savedId
            ? "Consulta salva"
            : ""}
      </span>
      {saved && (
        <>
          <ConfirmDeleteDialog
            id={saved.id}
            open={deleteOpen}
            onOpenChange={setDeleteOpen}
          />
          <RenameDialog
            id={saved.id}
            title={saved.title}
            open={renameOpen}
            onOpenChange={setRenameOpen}
          />
        </>
      )}
    </header>
  );
}
