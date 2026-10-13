import type { Technology } from "../../types/models";
import { TechChip } from "./TechChip";
export function TechChipGroup({
  items,
  value,
  onChange,
}: {
  items: Technology[];
  value: string;
  onChange: (id: string) => void;
}) {
  return (
    <div className="dd-chips" role="group" aria-label="Tecnologias">
      {items.map((t) => (
        <TechChip
          key={t.id}
          id={t.id}
          selected={value === t.id}
          onClick={() => onChange(value === t.id ? "" : t.id)}
        />
      ))}
    </div>
  );
}
