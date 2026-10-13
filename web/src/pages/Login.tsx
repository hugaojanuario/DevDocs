import { useState, type FormEvent } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { Brand } from "../components/layout/Brand";
import { Input } from "../components/ui/input";
import { Button } from "../components/ui/button";
import { Notice } from "../components/documentation/Notice";
import { login } from "../services/http";
import { safeNext } from "../lib/utils";
export function LoginPage() {
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const client = useQueryClient();
  async function submit(e: FormEvent) {
    e.preventDefault();
    if (pending || !password) return;
    setPending(true);
    setError("");
    try {
      const session = await login(password);
      setPassword("");
      client.setQueryData(["session"], session);
      navigate(safeNext(params.get("next")), { replace: true });
    } catch (e) {
      setError(e instanceof Error ? e.message : "Falha ao entrar.");
    } finally {
      setPending(false);
    }
  }
  return (
    <main id="conteudo" className="dd-login">
      <div className="dd-login-card">
        <Brand />
        <h1>Entrar no DevDocs</h1>
        <p>Acesse sua documentação e consultas salvas.</p>
        <form onSubmit={submit}>
          <label htmlFor="password">Senha</label>
          <Input
            id="password"
            type="password"
            autoComplete="current-password"
            autoFocus
            maxLength={1024}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          {error && (
            <Notice tone="error" title="Não foi possível entrar">
              {error}
            </Notice>
          )}
          <Button
            type="submit"
            variant="primary"
            loading={pending}
            disabled={!password}
          >
            {pending ? "Entrando…" : "Entrar"}
          </Button>
        </form>
        <p className="dd-login-note">
          Seu histórico de pesquisas não é armazenado.
        </p>
      </div>
    </main>
  );
}
