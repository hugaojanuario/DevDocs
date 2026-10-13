import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import * as saved from "../services/saved";
import * as tech from "../services/technologies";
import { postSearch } from "../services/search";
import { health, ready } from "../services/system";
import { ApiError } from "../services/http";
import type { Answer } from "../types/models";
export const useSearch = (q: string, technology: string, official: boolean) =>
  useQuery({
    queryKey: ["search", q, technology, official],
    queryFn: ({ signal }) =>
      postSearch(
        {
          query: q,
          language: "pt-BR",
          preferred_sources: official
            ? ["official"]
            : ["official", "complementary"],
          ai_enabled: false,
          ...(technology
            ? {
                technology: technology as NonNullable<
                  import("../types/models").SearchRequest["technology"]
                >,
              }
            : {}),
        },
        signal,
      ),
    enabled: !!q,
    staleTime: 0,
    gcTime: 5 * 60_000,
    retry: (count, error) =>
      count < 1 &&
      !(error instanceof ApiError && error.status >= 400 && error.status < 500),
  });
export const useTechnologies = () =>
  useQuery({
    queryKey: ["technologies"],
    queryFn: ({ signal }) => tech.listTechnologies(signal),
    staleTime: 300_000,
  });
export const useSources = () =>
  useQuery({
    queryKey: ["sources"],
    queryFn: ({ signal }) => tech.listSources(signal),
    staleTime: 300_000,
  });
export const useSavedList = (params: URLSearchParams) =>
  useQuery({
    queryKey: ["saved", "list", params.toString()],
    queryFn: ({ signal }) => saved.listSaved(params, signal),
  });
export const useSaved = (id: string) =>
  useQuery({
    queryKey: ["saved", id],
    queryFn: ({ signal }) => saved.getSaved(id, signal),
    enabled: !!id,
    retry: false,
  });
export function useCreateSaved() {
  const c = useQueryClient();
  return useMutation({
    mutationFn: (a: Answer) => saved.createSaved(a),
    onSuccess: () => c.invalidateQueries({ queryKey: ["saved"] }),
  });
}
export function useRenameSaved() {
  const c = useQueryClient();
  return useMutation({
    mutationFn: ({ id, title }: { id: string; title: string }) =>
      saved.renameSaved(id, title),
    onSuccess: (value) => {
      c.setQueryData(["saved", value.id], value);
      void c.invalidateQueries({ queryKey: ["saved", "list"] });
    },
  });
}
export function useDeleteSaved() {
  const c = useQueryClient();
  return useMutation({
    mutationFn: saved.deleteSaved,
    onSuccess: (_, id) => {
      c.removeQueries({ queryKey: ["saved", id], exact: true });
      c.removeQueries({ queryKey: ["search"] });
      void c.invalidateQueries({ queryKey: ["saved"] });
    },
  });
}
export const useHealth = () =>
  useQuery({
    queryKey: ["system"],
    queryFn: async ({ signal }) => {
      await health(signal);
      return ready(signal);
    },
    refetchInterval: 30_000,
    retry: false,
  });
