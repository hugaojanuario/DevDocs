import { http } from "./http";
import type { Technology, Source } from "../types/models";
export const listTechnologies = (signal?: AbortSignal) =>
  http<Technology[]>("/technologies", { signal });
export const listSources = (signal?: AbortSignal) =>
  http<Source[]>("/sources", { signal });
