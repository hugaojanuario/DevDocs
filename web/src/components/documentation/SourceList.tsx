import { ExternalLink } from "lucide-react";
import type { Reference } from "../../types/models";
import { Badge } from "../common/Badge";
import { Icon } from "../common/Icon";
import { safeURL } from "../../lib/utils";
export function SourceList({ items }: { items: Reference[] }) {
  const sorted = [
    ...items.filter((x) => x.official === true),
    ...items.filter((x) => x.official !== true),
  ];
  return (
    <ul className="dd-src">
      {sorted.map((source) => {
        const url = safeURL(source.url);
        if (!url) return null;
        return (
          <li key={source.id} id={source.id}>
            <a href={url} target="_blank" rel="noopener noreferrer">
              <div className="dd-src-body">
                <span className="dd-src-title">{source.title}</span>
                <div className="dd-src-meta">
                  <span className="dd-src-domain">{source.domain}</span>
                  <Badge
                    tone={source.official === true ? "official" : "neutral"}
                  >
                    {source.official === true ? "Oficial" : "Complementar"}
                  </Badge>
                  {source.version && <span>v{source.version}</span>}
                </div>
              </div>
              <Icon icon={ExternalLink} />
              <span className="sr-only">(abre em nova aba)</span>
            </a>
          </li>
        );
      })}
    </ul>
  );
}
