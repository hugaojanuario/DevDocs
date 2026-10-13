import type { Answer, SavedQuery } from "../../types/models";
import { ResultHeader } from "./ResultHeader";
import { Notice } from "./Notice";
import { DocSection } from "./DocSection";
import { SourceList } from "./SourceList";
import { CodeBlock } from "../code/CodeBlock";
import { CodeExplanation } from "../code/CodeExplanation";
export function ResultView({
  answer,
  saved,
}: {
  answer: Answer;
  saved?: SavedQuery;
}) {
  return (
    <article className="dd-doc">
      <ResultHeader answer={answer} saved={saved} />
      {answer.mode === "ai_assisted" && (
        <div className="mt-6">
          <Notice tone="ai" title="Síntese por IA">
            Esta resposta foi sintetizada por IA com base nos trechos citados.
            Confira as fontes.
          </Notice>
        </div>
      )}
      {(answer.status === "partial" || answer.quality === "limited") && (
        <div className="mt-6">
          <Notice tone="partial" title="Conteúdo parcial">
            Algumas fontes, descrições ou exemplos não estavam disponíveis.
            Exibimos apenas o que foi fundamentado.
          </Notice>
        </div>
      )}
      {answer.summary && (
        <DocSection title="O que é?" id="o-que-e">
          <p>{answer.summary}</p>
        </DocSection>
      )}
      {answer.description && (
        <DocSection title="Como funciona?" id="como-funciona">
          <p>{answer.description}</p>
        </DocSection>
      )}
      {answer.examples.length > 0 && (
        <DocSection
          title={answer.examples.length === 1 ? "Exemplo" : "Exemplos"}
          id="exemplos"
        >
          <div className="dd-example-stack">
            {answer.examples.map((example, i) => (
              <div key={i}>
                <p className="mb-3 text-sm text-ink-muted">
                  Trecho da documentação. Ajuste arquivos, variáveis e
                  dependências ao seu ambiente.
                </p>
                <CodeBlock
                  code={
                    example.setup
                      ? `${example.setup}\n\n${example.code}`
                      : example.code
                  }
                  language={example.language}
                  title="Código"
                />
                {example.output && (
                  <div className="mt-4">
                    <CodeBlock
                      code={example.output}
                      language="text"
                      title="Saída documentada"
                      showLineNumbers={false}
                    />
                  </div>
                )}
                <CodeExplanation items={example.explanations} />
              </div>
            ))}
          </div>
        </DocSection>
      )}
      {answer.observations.length > 0 && (
        <DocSection title="Observações" id="observacoes">
          <ul className="dd-observations">
            {answer.observations.map((note, i) => (
              <li key={i}>{note}</li>
            ))}
          </ul>
        </DocSection>
      )}
      <DocSection title="Fontes e documentação" id="fontes">
        {answer.references.length ? (
          <SourceList items={answer.references} />
        ) : (
          <Notice tone="partial" title="Nenhuma fonte foi retornada">
            Não há referências verificáveis para esta resposta.
          </Notice>
        )}
      </DocSection>
    </article>
  );
}
