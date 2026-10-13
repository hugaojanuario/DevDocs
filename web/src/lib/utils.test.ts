import { describe, it, expect } from "vitest";
import { safeNext, safeURL } from "./utils";
describe("URLs do frontend", () => {
  it("bloqueia redirecionamento externo e esquemas ativos", () => {
    for (const url of ["//evil.test", "https://evil.test", "/\\evil.test"])
      expect(safeNext(url)).toBe("/");
    expect(safeNext("/saved")).toBe("/saved");
    expect(safeURL("javascript:alert(1)")).toBeNull();
    expect(safeURL("https://docs.python.org/3/")).toBe(
      "https://docs.python.org/3/",
    );
  });
});
