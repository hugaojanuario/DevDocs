import { Link } from "react-router-dom";
export function Brand() {
  return (
    <Link className="dd-brand" to="/" aria-label="DevDocs, página inicial">
      devdocs
      <span className="dd-brand-caret" aria-hidden />
    </Link>
  );
}
