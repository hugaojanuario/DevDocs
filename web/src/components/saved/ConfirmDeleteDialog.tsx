import { useRef } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { Trash2, CircleAlert } from "lucide-react";
import { toast } from "sonner";
import * as Alert from "../ui/alert-dialog";
import { Button } from "../ui/button";
import { useDeleteSaved } from "../../hooks/queries";
export function ConfirmDeleteDialog({
  id,
  open,
  onOpenChange,
}: {
  id: string;
  open: boolean;
  onOpenChange: (value: boolean) => void;
}) {
  const cancel = useRef<HTMLButtonElement>(null);
  const mutation = useDeleteSaved();
  const navigate = useNavigate();
  const location = useLocation();
  return (
    <Alert.Root
      open={open}
      onOpenChange={(v) => {
        if (!mutation.isPending) {
          onOpenChange(v);
          mutation.reset();
        }
      }}
    >
      <Alert.Content
        onOpenAutoFocus={(e) => {
          e.preventDefault();
          cancel.current?.focus();
        }}
        onEscapeKeyDown={(e) => {
          if (mutation.isPending) e.preventDefault();
        }}
      >
        <Alert.Title>Excluir consulta permanentemente?</Alert.Title>
        <Alert.Description>
          Esta ação não pode ser desfeita. A consulta e todo o seu conteúdo
          serão removidos do DevDocs.
        </Alert.Description>
        {mutation.error && (
          <p className="dd-dialog-err" role="alert">
            <CircleAlert size={16} aria-hidden />
            {mutation.error.message}
          </p>
        )}
        <div className="dd-dialog-actions">
          <Alert.Cancel asChild>
            <Button ref={cancel} disabled={mutation.isPending}>
              Cancelar
            </Button>
          </Alert.Cancel>
          <Button
            variant="danger"
            loading={mutation.isPending}
            onClick={() =>
              mutation.mutate(id, {
                onSuccess: () => {
                  onOpenChange(false);
                  toast.success("Consulta excluída");
                  if (location.pathname === "/saved/" + id)
                    navigate("/saved", { replace: true });
                },
                onError: () =>
                  toast.error("Falha ao excluir consulta", {
                    duration: Infinity,
                  }),
              })
            }
          >
            {!mutation.isPending && <Trash2 size={16} aria-hidden />}
            {mutation.isPending ? "Excluindo…" : "Excluir permanentemente"}
          </Button>
        </div>
      </Alert.Content>
    </Alert.Root>
  );
}
