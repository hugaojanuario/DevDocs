import { http } from "./http";
export const health = (signal?: AbortSignal) =>
  http<{ status: "ok" }>("/health", { signal });
export const ready = (signal?: AbortSignal) =>
  http<{ status: "ready" }>("/ready", { signal });
