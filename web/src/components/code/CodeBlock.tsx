import { useEffect, useState } from "react";
import { Check, Copy } from "lucide-react";
import type { ThemedToken } from "shiki";
import { Button } from "../ui/button";
import { Icon } from "../common/Icon";
export function CodeBlock({
  code,
  language,
  title,
  showLineNumbers = true,
  highlightLines = [],
}: {
  code: string;
  language: string;
  title?: string;
  showLineNumbers?: boolean;
  highlightLines?: number[];
}) {
  const display = code.replace(/\n$/, "");
  const [tokens, setTokens] = useState<ThemedToken[][] | null>(null);
  const [copy, setCopy] = useState<"idle" | "done" | "error">("idle");
  useEffect(() => {
    let active = true;
    setTokens(null);
    void import("./shiki")
      .then((m) => m.highlight(display, language))
      .then((t) => {
        if (active) setTokens(t);
      })
      .catch(() => {});
    return () => {
      active = false;
    };
  }, [display, language]);
  useEffect(() => {
    if (copy === "idle") return;
    const t = setTimeout(() => setCopy("idle"), 1800);
    return () => clearTimeout(t);
  }, [copy]);
  const lines = display.split("\n");
  return (
    <figure className="dd-code" data-numbers={showLineNumbers}>
      <figcaption className="dd-code-head">
        <span className="dd-code-lang">
          {language || "text"}
          {title && <span className="dd-code-title"> · {title}</span>}
        </span>
        <Button
          variant="ghost"
          size="sm"
          data-state={copy}
          onClick={async () => {
            try {
              await navigator.clipboard.writeText(code);
              setCopy("done");
            } catch {
              setCopy("error");
            }
          }}
        >
          <Icon icon={copy === "done" ? Check : Copy} size={14} />
          {copy === "done" ? "Copiado" : copy === "error" ? "Falhou" : "Copiar"}
        </Button>
        <span className="sr-only" aria-live="polite">
          {copy === "done"
            ? "Código copiado"
            : copy === "error"
              ? "Falha ao copiar código"
              : ""}
        </span>
      </figcaption>
      <pre tabIndex={0} aria-label={"Código " + language}>
        <code>
          {lines.map((line, i) => (
            <span
              key={i}
              className="dd-ln"
              data-n={i + 1}
              data-hl={highlightLines.includes(i + 1)}
            >
              {tokens?.[i]
                ? tokens[i].map((token, j) => (
                    <span key={j} style={{ color: token.color }}>
                      {token.content}
                    </span>
                  ))
                : line}
              {i < lines.length - 1 ? "\n" : ""}
            </span>
          ))}
        </code>
      </pre>
    </figure>
  );
}
