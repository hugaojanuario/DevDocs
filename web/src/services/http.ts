import type { Session } from "../types/models";
let csrf = "";
export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string,
  ) {
    super(message);
  }
}
export function setSession(session: Session | null) {
  csrf = session?.csrf_token ?? "";
}
export async function http<T>(
  path: string,
  options: {
    method?: string;
    body?: unknown;
    signal?: AbortSignal;
    timeout?: number;
  } = {},
): Promise<T> {
  const controller = new AbortController();
  const abort = () => controller.abort(options.signal?.reason);
  options.signal?.addEventListener("abort", abort, { once: true });
  if (options.signal?.aborted) abort();
  const timer = setTimeout(
    () => controller.abort("timeout"),
    options.timeout ?? 30_000,
  );
  const method = options.method ?? "GET";
  const headers: Record<string, string> = {};
  if (options.body !== undefined) headers["Content-Type"] = "application/json";
  if (method !== "GET" && csrf) headers["X-CSRF-Token"] = csrf;
  try {
    const res = await fetch("/api/v1" + path, {
      method,
      headers,
      credentials: "include",
      body:
        options.body !== undefined ? JSON.stringify(options.body) : undefined,
      signal: controller.signal,
      cache: "no-store",
    });
    if (res.status === 204) return undefined as T;
    const value: unknown = await res.json();
    if (!res.ok) {
      const envelope = value as { error?: { code?: string; message?: string } };
      if (res.status === 401 && path !== "/auth/session")
        window.dispatchEvent(new Event("devdocs:unauthorized"));
      throw new ApiError(
        res.status,
        envelope.error?.code ?? "http_error",
        envelope.error?.message ?? "Falha na API.",
      );
    }
    return value as T;
  } catch (error) {
    if (error instanceof ApiError) throw error;
    if (controller.signal.aborted)
      throw new ApiError(
        0,
        controller.signal.reason === "timeout" ? "timeout" : "aborted",
        "A busca demorou mais que o esperado.",
      );
    throw new ApiError(
      0,
      "network_error",
      "Não foi possível conectar ao DevDocs.",
    );
  } finally {
    clearTimeout(timer);
    options.signal?.removeEventListener("abort", abort);
  }
}
export async function session(signal?: AbortSignal) {
  const value = await http<Session>("/auth/session", { signal });
  setSession(value);
  return value;
}
export async function login(password: string) {
  const value = await http<Session>("/auth/session", {
    method: "POST",
    body: { password },
  });
  setSession(value);
  return value;
}
export async function logout() {
  await http<void>("/auth/session", { method: "DELETE" });
  setSession(null);
}
