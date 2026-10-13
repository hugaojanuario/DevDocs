import { useState } from "react";
import type { SavedSummary } from "../../types/models";
import { SavedItem } from "./SavedItem";
import { ConfirmDeleteDialog } from "./ConfirmDeleteDialog";
export function SavedList({ items }: { items: SavedSummary[] }) {
  const [deleteId, setDeleteId] = useState("");
  return (
    <>
      <ul className="dd-saved-list">
        {items.map((item) => (
          <SavedItem
            key={item.id}
            item={item}
            onDelete={() => setDeleteId(item.id)}
          />
        ))}
      </ul>
      {deleteId && (
        <ConfirmDeleteDialog
          id={deleteId}
          open={!!deleteId}
          onOpenChange={(open) => {
            if (!open) setDeleteId("");
          }}
        />
      )}
    </>
  );
}
