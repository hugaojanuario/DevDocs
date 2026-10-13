import { Check } from "lucide-react";
export function TechChip({
  id,
  selected,
  onClick,
}: {
  id: string;
  selected: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      className="dd-chip"
      aria-pressed={selected}
      onClick={onClick}
    >
      {selected && <Check size={12} aria-hidden />}
      {id}
    </button>
  );
}
