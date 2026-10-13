import { Skeleton } from "../ui/skeleton";
export function ResultSkeleton() {
  return (
    <div
      className="dd-doc dd-result-skeleton"
      aria-busy="true"
      aria-label="Carregando documentação"
    >
      <div className="flex gap-2">
        <Skeleton className="dd-skel-badge" />
        <Skeleton className="dd-skel-badge" />
        <Skeleton className="dd-skel-badge" />
      </div>
      <Skeleton className="dd-skel-title" />
      <div className="dd-skel-lines">
        <Skeleton />
        <Skeleton className="dd-skel-medium" />
        <Skeleton className="dd-skel-short" />
      </div>
      <Skeleton className="dd-skel-code" />
    </div>
  );
}
