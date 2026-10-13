import { http } from "./http";
import type { Answer, SavedQuery, SavedPage } from "../types/models";
export const listSaved = (params: URLSearchParams, signal?: AbortSignal) =>
  http<SavedPage>("/saved?" + params, { signal });
export const getSaved = (id: string, signal?: AbortSignal) =>
  http<SavedQuery>("/saved/" + encodeURIComponent(id), { signal });
export const createSaved = (response: Answer) =>
  http<SavedQuery>("/saved", { method: "POST", body: { response } });
export const renameSaved = (id: string, title: string) =>
  http<SavedQuery>("/saved/" + encodeURIComponent(id), {
    method: "PATCH",
    body: { title },
  });
export const deleteSaved = (id: string) =>
  http<void>("/saved/" + encodeURIComponent(id), { method: "DELETE" });
