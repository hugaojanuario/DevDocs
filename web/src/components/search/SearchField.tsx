import { useEffect, useRef, useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { Search } from "lucide-react";
import { Button } from "../ui/button";
import { Icon } from "../common/Icon";
import { Kbd, modifier } from "../common/Kbd";
export function SearchField({
  compact = false,
  initial = "",
  technology = "",
  autoFocus = false,
  onSubmit,
}: {
  compact?: boolean;
  initial?: string;
  technology?: string;
  autoFocus?: boolean;
  onSubmit?: (q: string) => void;
}) {
  const [q, setQ] = useState(initial);
  const ref = useRef<HTMLInputElement>(null);
  const navigate = useNavigate();
  useEffect(() => setQ(initial), [initial]);
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const editing =
        ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName) ||
        target.isContentEditable;
      if (
        (e.key === "/" && !editing) ||
        ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k")
      ) {
        e.preventDefault();
        ref.current?.focus();
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, []);
  function submit(e: FormEvent) {
    e.preventDefault();
    const value = q.trim();
    if (value.length < 3) return;
    if (onSubmit) {
      onSubmit(value);
      return;
    }
    navigate(
      "/search?" +
        new URLSearchParams({
          q: value,
          ...(technology ? { tech: technology } : {}),
        }),
    );
  }
  return (
    <form
      role="search"
      onSubmit={submit}
      className={"dd-search " + (compact ? "dd-search-compact" : "")}
    >
      <span className="dd-search-ic">
        <Icon icon={Search} size={compact ? 16 : 18} />
      </span>
      <input
        ref={ref}
        type="search"
        aria-label="Pesquisar na documentação"
        placeholder="O que você quer aprender hoje?"
        value={q}
        onChange={(e) => setQ(e.target.value)}
        autoComplete="off"
        spellCheck={false}
        maxLength={600}
        autoFocus={autoFocus}
      />
      {!q && (
        <span className="dd-search-shortcuts">
          <Kbd>/</Kbd>
          {!compact && <Kbd>{modifier()}K</Kbd>}
        </span>
      )}
      {!compact && (
        <Button type="submit" variant="primary" disabled={q.trim().length < 3}>
          Buscar
        </Button>
      )}
    </form>
  );
}
