export function Skeleton({ className = "" }: { className?: string }) {
  return <span className={"dd-skel " + className} aria-hidden />;
}
