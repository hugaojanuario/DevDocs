import { FileSearch } from "lucide-react";
import { Icon } from "../common/Icon";
const examples = [
  "Como ler um CSV com Pandas?",
  "Como escrever uma CTE no PostgreSQL?",
  "Como criar uma DAG no Airflow?",
  "Como carregar um .env no Python?",
  "Como subir serviços com Docker Compose?",
];
export function Suggestions({ onSelect }: { onSelect: (q: string) => void }) {
  return (
    <div className="dd-suggest">
      {examples.map((text) => (
        <button key={text} type="button" onClick={() => onSelect(text)}>
          <Icon icon={FileSearch} />
          {text}
        </button>
      ))}
    </div>
  );
}
