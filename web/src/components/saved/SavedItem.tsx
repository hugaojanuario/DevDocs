import { Link } from "react-router-dom";
import { Trash2 } from "lucide-react";
import type { SavedSummary } from "../../types/models";
import { Badge } from "../common/Badge";
import { Button } from "../ui/button";
import { formatDate } from "../../lib/utils";
export function SavedItem({
  item,
  onDelete,
}: {
  item: SavedSummary;
  onDelete: () => void;
}) {
  return (
    <li className="dd-saved">
      <div className="dd-saved-main">
        <Link to={"/saved/" + item.id}>{item.title}</Link>
        <p className="dd-saved-q" title={item.question}>
          {item.question}
        </p>
      </div>
      <div className="dd-saved-meta">
        {item.technology && <Badge tone="tech">{item.technology}</Badge>}
        <time dateTime={item.created_at}>{formatDate(item.created_at)}</time>
        <Button
          variant="danger-ghost"
          size="sm"
          iconOnly
          aria-label={"Excluir " + item.title}
          onClick={onDelete}
        >
          <Trash2 size={14} aria-hidden />
        </Button>
      </div>
    </li>
  );
}
