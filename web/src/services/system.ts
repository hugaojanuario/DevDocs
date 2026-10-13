import { http } from "./http";
export const health = (signal?: AbortSignal) =>
  http<{ status: "ok" }>("/health", { signal });
export const ready = (signal?: AbortSignal) =>
  http<{ status: "ready" }>("/ready", { signal });

export const capabilities = (signal?: AbortSignal) =>
  http<{ ai_available: boolean; ai_provider: string; ai_model: string }>(
    "/capabilities",
    { signal },
  );
