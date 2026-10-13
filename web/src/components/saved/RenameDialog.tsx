import { useState, useEffect, type FormEvent } from "react";
import { toast } from "sonner";
import * as Dialog from "../ui/dialog";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { useRenameSaved } from "../../hooks/queries";
export function RenameDialog({
  id,
  title,
  open,
  onOpenChange,
}: {
  id: string;
  title: string;
  open: boolean;
  onOpenChange: (value: boolean) => void;
}) {
  const [value, setValue] = useState(title);
  const mutation = useRenameSaved();
  useEffect(() => {
    if (open) {
      setValue(title);
      mutation.reset();
    }
  }, [open, title]);
  function submit(e: FormEvent) {
    e.preventDefault();
    if (!value.trim() || mutation.isPending) return;
    mutation.mutate(
      { id, title: value.trim() },
      {
        onSuccess: () => {
          onOpenChange(false);
          toast.success("Título atualizado");
        },
        onError: () =>
          toast.error("Falha ao renomear consulta", { duration: Infinity }),
      },
    );
  }
  return (
    <Dialog.Root
      open={open}
      onOpenChange={(v) => {
        if (!mutation.isPending) onOpenChange(v);
      }}
    >
      <Dialog.Content
        onEscapeKeyDown={(e) => {
          if (mutation.isPending) e.preventDefault();
        }}
      >
        <Dialog.Title>Renomear consulta</Dialog.Title>
        <Dialog.Description>
          Atualize o título da consulta salva.
        </Dialog.Description>
        <form onSubmit={submit} className="dd-rename-form">
          <label htmlFor="saved-title">Título</label>
          <Input
            id="saved-title"
            maxLength={255}
            value={value}
            onChange={(e) => setValue(e.target.value)}
            autoFocus
          />
          {mutation.error && (
            <p className="dd-dialog-err" role="alert">
              {mutation.error.message}
            </p>
          )}
          <div className="dd-dialog-actions">
            <Dialog.Close asChild>
              <Button disabled={mutation.isPending}>Cancelar</Button>
            </Dialog.Close>
            <Button
              type="submit"
              variant="primary"
              loading={mutation.isPending}
              disabled={!value.trim()}
            >
              Salvar título
            </Button>
          </div>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
