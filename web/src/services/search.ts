import { http } from "./http";
import type { Answer, SearchRequest } from "../types/models";
export const postSearch = (body: SearchRequest, signal?: AbortSignal) =>
  http<Answer>("/search", { method: "POST", body, signal });
