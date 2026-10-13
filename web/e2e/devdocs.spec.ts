import { test, expect, type Cookie } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";
import { readFileSync } from "node:fs";
import type { Answer } from "../src/types/models";
const env = Object.fromEntries(
  readFileSync("../.env", "utf8")
    .split("\n")
    .filter((l) => l && !l.startsWith("#"))
    .map((l) => {
      const i = l.indexOf("=");
      return [l.slice(0, i), l.slice(i + 1)];
    }),
);
const origin = process.env.DEVDOCS_TEST_URL ?? "http://localhost:8080";
let cookies: Cookie[] = [];
let csrf = "";
const answer: Answer = {
  query: "Python CSV fixture",
  title: "Leitura de CSV com evidências",
  technology: "python",
  version: "",
  summary: "Read CSV using the documented reader.",
  description: "",
  examples: [
    {
      language: "python",
      code: 'if True:\n    read_csv("fixture.csv")\n' + "# ".repeat(80),
      explanations: [],
      reference_id: "source-1",
    },
  ],
  observations: ["Fixture fornecida somente pelo teste de interface."],
  references: [
    {
      id: "source-1",
      title: "Documentação Python",
      url: "https://docs.python.org/3/library/csv.html",
      domain: "docs.python.org",
      official: true,
      technology: "python",
      version: "",
      excerpt: "Read CSV using the documented reader.",
    },
  ],
  evidence: [
    {
      reference_id: "source-1",
      field: "summary",
      quote: "Read CSV using the documented reader.",
    },
    {
      reference_id: "source-1",
      field: "example",
      quote: 'if True:\n    read_csv("fixture.csv")\n' + "# ".repeat(80),
    },
  ],
  mode: "deterministic",
  status: "completed",
  quality: "extracted",
  language: "pt-BR",
};
test.beforeAll(async ({ request }) => {
  const res = await request.post(origin + "/api/v1/auth/session", {
    headers: { Origin: origin },
    data: { password: env.WEB_PASSWORD },
  });
  expect(res.status()).toBe(200);
  csrf = (await res.json()).csrf_token;
  cookies = (await request.storageState()).cookies;
});
test.beforeEach(async ({ context }) => {
  await context.addCookies(cookies);
});
test("home, temas, teclado e privacidade", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "Encontre respostas",
  );
  await expect(page.locator("html")).toHaveAttribute("data-theme", "dark");
  await page.getByRole("button", { name: "Alterar tema" }).click();
  await page.getByRole("menuitem", { name: "Claro", exact: true }).click();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "light");
  await page.reload();
  await expect(page.locator("html")).toHaveAttribute("data-theme", "light");
  const storage = await page.evaluate(() => ({
    local: Object.keys(localStorage),
    session: Object.keys(sessionStorage),
  }));
  expect(storage.local).toEqual(["devdocs.theme"]);
  expect(storage.session).toEqual([]);
  await page.locator("h1").click();
  await page.keyboard.press("/");
  await expect(
    page.getByRole("searchbox", { name: "Pesquisar na documentação" }),
  ).toBeFocused();
  const axe = await new AxeBuilder({ page }).analyze();
  expect(axe.violations).toEqual([]);
});
test("resultado, salvar sem nova busca, renomear e excluir fisicamente", async ({
  page,
  request,
}) => {
  let searches = 0;
  await page.route("**/api/v1/search", (route) => {
    searches++;
    return route.fulfill({ status: 200, json: answer });
  });
  await page.goto("/search?q=Python+CSV+fixture");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    answer.title,
  );
  await expect(
    page.getByRole("heading", { name: "Como funciona?" }),
  ).toHaveCount(0);
  await expect(page.locator("pre code")).toContainText("    read_csv");
  await page
    .getByRole("button", { name: "Salvar consulta", exact: true })
    .click();
  await expect(
    page.getByRole("button", { name: "Salvo", exact: true }),
  ).toBeDisabled();
  expect(searches).toBe(1);
  await page.getByRole("link", { name: "Ver em Consultas salvas" }).click();
  await page.getByRole("button", { name: "Renomear", exact: true }).click();
  await page.getByLabel("Título", { exact: true }).fill("Fixture renomeada");
  await page.getByRole("button", { name: "Salvar título" }).click();
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    "Fixture renomeada",
  );
  const id = page.url().split("/").at(-1)!;
  await page.getByRole("button", { name: "Excluir", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "Cancelar", exact: true }),
  ).toBeFocused();
  await page.getByRole("button", { name: "Excluir permanentemente" }).click();
  await expect(page).toHaveURL(/\/saved$/);
  const res = await request.get(origin + "/api/v1/saved/" + id, {
    headers: { Authorization: "Bearer " + env.CLI_TOKEN },
  });
  expect(res.status()).toBe(404);
  const axe = await new AxeBuilder({ page }).analyze();
  expect(axe.violations).toEqual([]);
});
test("mobile: Sheet e código sem quebra", async ({ page }) => {
  await page.setViewportSize({ width: 360, height: 800 });
  await page.route("**/api/v1/search", (route) =>
    route.fulfill({ json: answer }),
  );
  await page.goto("/search?q=Python+CSV+fixture");
  await expect(page.locator("pre")).toBeVisible();
  expect(
    await page.locator("pre").evaluate((e) => e.scrollWidth > e.clientWidth),
  ).toBe(true);
  expect(
    await page.locator("html").evaluate((e) => e.scrollWidth <= innerWidth),
  ).toBe(true);
  await page.getByRole("button", { name: "Abrir menu" }).click();
  await expect(
    page.getByRole("dialog", { name: "Menu principal" }),
  ).toBeVisible();
  await page
    .getByRole("dialog")
    .getByRole("link", { name: "Tecnologias", exact: true })
    .click();
  await expect(
    page.getByRole("heading", { name: "Tecnologias", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Abrir pesquisa" }).click();
  const searchSheet = page.getByRole("dialog", { name: "Nova pesquisa" });
  await expect(searchSheet).toBeVisible();
  expect((await searchSheet.boundingBox())?.width).toBe(360);
  await page.keyboard.press("Escape");
  await expect(searchSheet).toHaveCount(0);
});
test("erros e parcial não inventam conteúdo", async ({ page }) => {
  await page.route("**/api/v1/search", (route) =>
    route.fulfill({
      json: {
        ...answer,
        status: "partial",
        quality: "limited",
        summary: "",
        description: "",
        examples: [],
        evidence: [],
      },
    }),
  );
  await page.goto("/search?q=Python+CSV+fixture");
  await expect(
    page.getByText("Conteúdo parcial", { exact: true }),
  ).toBeVisible();
  await expect(page.getByRole("heading", { name: "O que é?" })).toHaveCount(0);
  await expect(
    page.getByRole("heading", { name: "Exemplos", exact: true }),
  ).toHaveCount(0);
  await page.unroute("**/api/v1/search");
  await page.route("**/api/v1/search", (route) =>
    route.fulfill({
      status: 503,
      json: {
        error: {
          code: "search_unavailable",
          message: "Fonte indisponível",
          request_id: "test",
        },
      },
    }),
  );
  await page.goto("/search?q=Unavailable+fixture");
  await expect(page.getByRole("alert")).toContainText(
    "As fontes de documentação estão indisponíveis",
  );
});
test("falha ao excluir mantém diálogo e item", async ({ page, request }) => {
  const res = await request.post(origin + "/api/v1/saved", {
    headers: { Authorization: "Bearer " + env.CLI_TOKEN },
    data: { response: answer, title: "Fixture falha de exclusão" },
  });
  expect(res.status()).toBe(201);
  const id = (await res.json()).id;
  try {
    await page.goto("/saved/" + id);
    await expect(page.getByRole("heading", { level: 1 })).toBeVisible();
    await page.route("**/api/v1/saved/" + id, (route) =>
      route.request().method() === "DELETE"
        ? route.fulfill({
            status: 503,
            json: {
              error: {
                code: "database_unavailable",
                message: "Falha temporária",
                request_id: "test",
              },
            },
          })
        : route.continue(),
    );
    await page.getByRole("button", { name: "Excluir", exact: true }).click();
    await page.getByRole("button", { name: "Excluir permanentemente" }).click();
    await expect(page.getByRole("alertdialog")).toBeVisible();
    await expect(
      page.getByRole("alertdialog").getByRole("alert"),
    ).toContainText("Falha temporária");
  } finally {
    await request.delete(origin + "/api/v1/saved/" + id, {
      headers: { Authorization: "Bearer " + env.CLI_TOKEN },
    });
  }
});
test("resultado é acessível e não usa storage de pesquisas", async ({
  page,
}) => {
  await page.route("**/api/v1/search", (route) =>
    route.fulfill({ json: answer }),
  );
  await page.goto("/search?q=Python+CSV+fixture");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText(
    answer.title,
  );
  const axe = await new AxeBuilder({ page }).analyze();
  expect(axe.violations).toEqual([]);
  const keys = await page.evaluate(async () => ({
    local: Object.keys(localStorage),
    session: Object.keys(sessionStorage),
    cache: await caches.keys(),
    databases: await indexedDB.databases(),
  }));
  expect(keys.local.every((k) => k === "devdocs.theme")).toBe(true);
  expect(keys.session).toEqual([]);
  expect(keys.cache).toEqual([]);
  expect(keys.databases).toEqual([]);
});

test("comandos, imports e saída têm apresentação separada", async ({
  page,
  context,
}) => {
  const example = {
    language: "python",
    code: 'pd.read_csv("data.csv")',
    source_code: '>>> pd.read_csv("data.csv")\n   Name Value\n0  foo 1',
    setup: "import pandas as pd",
    output: "   Name Value\n0  foo 1",
    explanations: [],
    reference_id: "source-1",
  };
  await page.route("**/api/v1/search", (route) =>
    route.fulfill({ json: { ...answer, examples: [example] } }),
  );
  await context.grantPermissions(["clipboard-read", "clipboard-write"]);
  await page.goto("/search?q=Como+ler+CSV+com+Pandas");
  await expect(page.locator('pre[aria-label="Código python"]')).toHaveText(
    'import pandas as pd\n\npd.read_csv("data.csv")',
  );
  await expect(page.locator('pre[aria-label="Código text"]')).toHaveText(
    example.output,
  );
  await expect(
    page.locator(".dd-code-title").filter({ hasText: "Saída documentada" }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Copiar", exact: true })
    .first()
    .click();
  expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(
    'import pandas as pd\n\npd.read_csv("data.csv")',
  );
});
