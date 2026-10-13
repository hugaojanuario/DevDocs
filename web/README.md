# DevDocs Web

React + TypeScript + Vite, Tailwind 3.4.17, primitivos Radix no padrão shadcn/ui, TanStack Query,
Lucide, Sonner e Shiki lazy. Design: `../docs/DEVDOCS_DESIGN_HANDOFF.md`; dados: `../api/openapi.yaml`.
Tokens dark/light copiados literalmente em `src/styles/tokens.css`; fontes são locais, sem pedidos
à Google Fonts. Há layout de documentação, sem chat, com navegação responsiva e Sheet mobile.

```sh
# Primeiro suba a API/proxy com make up na raiz.
npm ci                 # npm i também funciona; lockfile mantido no repo
npm run types          # OpenAPI -> src/types/api.ts
npm run dev            # http://localhost:5173
npm run build
npm run preview        # http://localhost:4173
npm test
npx playwright install chromium
npm run test:e2e        # usa .env da raiz, somente para testes locais
```

O proxy local de desenvolvimento encaminha `/api` ao `http://localhost:8080` e ajusta Origin
apenas para as origens locais de desenvolvimento autorizadas. `DEVDOCS_PROXY_TARGET` pode
alterar o alvo do proxy (variável do processo Vite, sem segredo no cliente).
Em produção, Docker faz build multi-stage e Nginx serve a interface na porta 8080, proxy para API
privada. Após mudar o contrato, execute `npm run types`; não edite os tipos gerados à mão.

Rotas: `/`, `/search?q=...&tech=...`, `/saved`, `/saved/:id`, `/technologies`, `/settings`, `/login`.
Busca envia `technology` opcional, nunca campo fora do contrato. Modo IA é `ai_assisted` e notas
são `observations`. Fontes oficiais são exibidas somente com `official === true` da API.
Salvos usam filtros e ordenação globais no banco, com paginação. Buscar novamente não sobrescreve
consulta salva. Falha ao excluir mantém o diálogo e o item; cancelar recebe foco inicial.

Login: POST `/api/v1/auth/session` com senha. Cookie HttpOnly permanece com o navegador; JSON
retorna CSRF, mantido em RAM. GET recupera CSRF após reload. Mutações enviam `X-CSRF-Token`,
Origin e cookie. Logout invalida sessão e esvazia cache de interface. Nunca coloque CLI_TOKEN,
WEB_PASSWORD, BRAVE_API_KEY em VITE_*, bundle, localStorage ou URL.

Não há fixtures no frontend de produção, nem flag `VITE_USE_FIXTURES` que substitua a busca.
Fixtures são exclusivas de Playwright, injetadas por interceptação de HTTP nos testes. Nenhum
resultado sintético é usado em requisições de produção. Somente `devdocs.theme` é persistido;
preferência de fontes, pesquisa, resposta e TanStack Query ficam em RAM, sem persistQueryClient.
Links de pesquisa contêm a pergunta; o histórico nativo do navegador pode guardar essa URL.

`npm audit --omit=dev` verifica dependências do bundle. O Tailwind **3.4.17 exigido pelo handoff**
ainda depende de `braces@3.0.3`, marcado por GHSA-vfj7-8cjw-p6xm (recursão em padrões de glob).
A auditoria completa aponta cinco entradas de severidade alta nessa cadeia de ferramentas de
build, sem correção disponível mantendo Tailwind 3. `braces` não entra no bundle nem processa
perguntas. Vitest foi atualizado e o parser PostCSS corrigido por override. Não há vulnerabilidades
conhecidas nas dependências de produção nesta validação; atualizar Tailwind para 4 requer revisão
do handoff/configuração. Esta limitação está registrada, não ocultada pela auditoria.
