import type { Example } from "../../types/models";
export function CodeExplanation({ items }: { items: Example["explanations"] }) {
  if (!items.length) return null;
  return (
    <div className="dd-explain">
      <table>
        <caption>Entendendo o código</caption>
        <thead>
          <tr>
            <th scope="col">Trecho</th>
            <th scope="col">Explicação</th>
          </tr>
        </thead>
        <tbody>
          {items.map((item, i) => (
            <tr key={i}>
              <td>
                <code>{item.snippet}</code>
              </td>
              <td>
                {item.description}
                <a href={"#" + item.reference_id} className="dd-evidence-link">
                  [{item.reference_id}]
                </a>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
