import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { ShieldCheck } from "lucide-react";
import { SearchField } from "../components/search/SearchField";
import { TechChipGroup } from "../components/search/TechChipGroup";
import { Suggestions } from "../components/search/Suggestions";
import { Notice } from "../components/documentation/Notice";
import { useTechnologies } from "../hooks/queries";
export function Home() {
  const navigate = useNavigate();
  const [params, setParams] = useSearchParams();
  const [tech, setTech] = useState(params.get("tech") ?? "");
  const technologies = useTechnologies();
  useEffect(() => {
    setTech(params.get("tech") ?? "");
  }, [params]);
  return (
    <div className="dd-home">
      <h1>
        Encontre respostas
        <br className="dd-home-break" /> na documentação
      </h1>
      <p className="dd-home-description">
        Respostas a partir de documentação oficial, com código explicado e links
        para as fontes.
      </p>
      <div className="dd-home-search">
        <SearchField autoFocus technology={tech} />
        <div className="mt-4">
          {technologies.data && (
            <TechChipGroup
              items={technologies.data}
              value={tech}
              onChange={(value) => {
                setTech(value);
                setParams(value ? { tech: value } : {}, { replace: true });
              }}
            />
          )}
          {technologies.error && (
            <Notice
              tone="error"
              title="Não foi possível carregar as tecnologias"
            />
          )}
        </div>
      </div>
      <section className="dd-home-suggestions" aria-labelledby="experimente">
        <h2 id="experimente">Experimente</h2>
        <Suggestions
          onSelect={(q) =>
            navigate(
              "/search?" +
                new URLSearchParams({ q, ...(tech ? { tech } : {}) }),
            )
          }
        />
      </section>
      <p className="dd-home-note">
        <ShieldCheck size={16} strokeWidth={1.75} aria-hidden />
        Priorizamos fontes oficiais. Conteúdo gerado por IA é sempre sinalizado.
      </p>
    </div>
  );
}
