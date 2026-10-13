![DevDocs — Documentação técnica verificável, direta e sem conversa.](docs/assets/devdocs-banner.png)

# DevDocs

Pesquisa de documentação técnica com evidências, exemplos de código e referências verificáveis.
Backend Go, API REST, CLI remota, PostgreSQL e interface React implementada a partir do handoff.
Não há chatbot nem histórico automático de pesquisas.

## Executar

Requisitos: Docker com Compose v2. Para desenvolvimento: Go 1.26.5+, Node 24+ e Python 3.
As imagens compilam com Go 1.27.2 e Node 24. SQL gerado com sqlc 1.31.1.

```sh
python3 scripts/init-env.py    # cria .env 0600; não sobrescreve nem imprime segredos
make up                      # build, migração versionada e health checks
```

Abra **http://localhost:8080**. A senha web está em `WEB_PASSWORD` no `.env` local.
`CLI_TOKEN` é outro segredo, exclusivo da CLI. O portal Maestri também pode abrir essa URL.
Não comite `.env`; o arquivo é ignorado pelo Git e pelo contexto de build Docker.
`make down` preserva o volume PostgreSQL. `make migrate` aplica migrações pendentes sem apagar salvos.

## Busca e limites reais

`SEARCH_PROVIDER=official` funciona sem API paga: consulta índices **Sphinx reais** de Python,
Pandas, SQLAlchemy e Airflow quando acessíveis. Os índices são baixados transitoriamente;
nenhuma pergunta é enviada nesses pedidos de índice. O registro inclui também PostgreSQL, Go,
Docker, Spark, dbt, FastAPI e Kubernetes. Registro de fonte não significa índice público disponível.
Tecnologias sem índice compatível, fontes que bloqueiam acesso e perguntas sem evidência retornam
`503 search_unavailable`, nunca resultados simulados. `read_csv` é priorizado sobre páginas
que apenas mencionam CSV. Use `technology` para restringir a tecnologia sem alterar a pergunta.

Para pesquisa ampla, configure `SEARCH_PROVIDER=brave` e `BRAVE_API_KEY` no `.env` e rode `make up`.
O provedor usa a [API oficial Brave](https://api-dashboard.search.brave.com/api-reference/web/search/get),
sem scraping de páginas de resultados; tenta os índices oficiais se a API falhar.
A política de retenção do provedor externo é responsabilidade desse provedor; DevDocs não cria
logs, analytics ou arquivos de pesquisas. Fontes oficiais são reconhecidas por host e caminho
registrados, não por substring do domínio. Para adicionar fontes/tecnologias, edite
`internal/search/sources.go` e regenere o contrato/tipos quando acrescentar tecnologias selecionáveis.

O resultado padrão traz um resumo curto e um exemplo principal. Sessões Python/IPython e
shell têm prompts removidos e saída separada do código. Imports documentados podem ser
reutilizados das páginas consultadas, com sua própria referência. O botão de copiar inclui
esses imports e os comandos, sem copiar a saída. Os exemplos são trechos documentados;
arquivos, variáveis e dependências podem precisar de ajustes ao ambiente do usuário.

Pedidos básicos de hello world em
Python usam um template curto, com `print` e instrução de execução, sustentado por trechos
coletados ao vivo da documentação oficial; Tkinter e asyncio exigem pedidos explícitos.

Respostas são **determinísticas**: trechos e código de páginas acessíveis. O bloco literal é
preservado em `source_code` e nas evidências; `code`, `setup` e `output` separam sua apresentação.
Não há tradução
inventada, interpretação de cada linha ou preenchimento de campos ausentes. A interface omite
seções vazias e sinaliza conteúdo parcial. Explicações são parágrafos próximos ao exemplo, citados;
exatidão didática perfeita não é garantida. A estrutura distingue resumo, descrição, exemplos,
observações, referências e evidências. Versões são extraídas de metadados, título ou URL quando
identificáveis. Os índices oficiais usam a versão publicada/stable; uma versão solicitada pode
não estar disponível, e a resposta sinaliza quando não consegue confirmá-la.

`AnswerProvider` prepara síntese opcional futura. Nenhum LLM está ativado: `ai_enabled=true`
retorna `503 ai_unavailable`. O frontend mantém o controle de IA desabilitado e sabe rotular
respostas `ai_assisted` preservadas no contrato. A V1 funciona sem IA paga.

## Persistência e exclusão

`POST /search` não utiliza o banco. Resultado e cache existem apenas em RAM; cache padrão de
5 minutos, 32 entradas, até 256 KiB por entrada, máximo configurável de 128 entradas/15 minutos.
Expiração remove entradas da RAM, reiniciar remove todo o cache, e exclusão explícita limpa o cache.
HTML de documentos não é gravado. Nginx não registra acessos nem erros com URLs de consultas;
a API usa `slog` somente para eventos de inicialização, sem corpos, perguntas ou credenciais.

Somente **Salvar consulta**, `POST /saved` ou `devdocs search --save` inserem dados.
Salvar recebe a resposta existente e não refaz pesquisa. Salvos têm UUID, pergunta, título,
tecnologia, resposta JSON e timestamps; podem ser filtrados, paginados, ordenados e renomeados.
Não existe tabela de histórico, telemetria, soft delete, lixeira ou limpeza automática de salvos.
Exclusão executa literalmente `DELETE FROM saved_queries WHERE id = $1`.

**WAL, backups e processos físicos do PostgreSQL podem reter bytes temporariamente.** Exclusão
remove a linha logicamente acessível e fisicamente da tabela por DELETE, não promete apagamento
forense de discos ou backups. Defina manualmente a política dos seus backups.

No frontend só `devdocs.theme` entra no localStorage. Consultas, respostas, preferências de fontes
e cache TanStack vivem em memória. Não há IndexedDB, sessionStorage, service worker ou cache
persistido. Conforme o handoff, links de pesquisa contêm a pergunta na URL: o navegador pode
reter essa URL no seu próprio histórico. Esses links não contêm a resposta ou credenciais.

## API e autenticação

Contrato completo: [`api/openapi.yaml`](api/openapi.yaml), OpenAPI 3.0.3, disponível em `/openapi.yaml`.
O arquivo usa sintaxe JSON, válida em YAML. Prefixo `/api/v1`.

| Método | Rota | Uso |
|---|---|---|
| POST | `/search` | Pesquisa transitória |
| POST | `/saved` | Persistência explícita da resposta atual |
| GET | `/saved` | `page`, `page_size`, `q`, `technology`, `order=recent/oldest/title` |
| GET / PATCH / DELETE | `/saved/{id}` | Abrir, renomear, excluir fisicamente |
| GET | `/technologies`, `/sources` | Registro documental |
| GET | `/health`, `/ready` | Liveness e banco/migração prontos |
| POST / GET / DELETE | `/auth/session` | Login, CSRF da sessão, logout |

API single-user. CLI usa `Authorization: Bearer <CLI_TOKEN>`; nunca inclui token na URL.
Web envia a senha em POST de login e recebe cookie HttpOnly, SameSite=Strict, Secure em origem
HTTPS, válido por 8 horas. Cookie é aleatório, identidade guardada por hash em RAM; logout o
invalida e reinício invalida todas as sessões. CSRF também fica em RAM, obtido via GET da sessão.
Toda mutação por cookie exige `X-CSRF-Token` e `Origin` igual a `WEB_ORIGIN`, inclusive busca.
Login exige a origem configurada. CORS permite exclusivamente essa origem. Sem cadastro público.

Corpos: no máximo 1 MiB; pergunta 3–600 caracteres, título 1–255, page_size 1–100.
JSON com campos desconhecidos e opções inválidas é rejeitado. Salvar valida estrutura, URLs,
referências, marcação oficial e associação de evidências ao código. Erros padronizados incluem
`code`, `message`, `request_id`; dados/segredos de upstream não aparecem nos erros.
Limites globais single-user: 120 requisições autenticadas/minuto, 20 buscas/minuto, 5 logins/minuto.
`429` fornece `Retry-After`. Não há mapas ilimitados por IP nem confiança em X-Forwarded-For.

A obtenção de fontes valida HTTP(S), portas 80/443, todas as respostas DNS e cada redirecionamento;
a conexão usa IP público fixado após resolver DNS e mantém TLS validado. Bloqueia loopback,
privados, link-local, CGNAT, metadados, ranges reservados e mecanismos IPv6 de transição.
Não usa proxy HTTP do ambiente. Respeita robots.txt, crawl-delay, limites de 4 conexões,
redirecionamentos, prazo de 12s por requisição e 2 MiB por documento (índices: 8 MiB).
Robots indisponível/barreira de acesso resulta em fonte indisponível. Não contorna autenticação,
CAPTCHA ou bloqueio. Observe os termos da fonte antes de habilitar uma nova fonte.

## CLI

```sh
make build
./bin/devdocs config set-server http://localhost:8080
./bin/devdocs config set-token            # lê token sem eco; permite stdin seguro
./bin/devdocs ping
./bin/devdocs search "Como ler CSV usando Pandas?"
./bin/devdocs search "Como ler CSV usando Pandas?" --format markdown --no-color
./bin/devdocs search "Como ler CSV usando Pandas?" --format json --save
./bin/devdocs saved list --technology pandas --page 1 --page-size 20
./bin/devdocs saved show <uuid>
./bin/devdocs saved rename <uuid> "Novo título"
./bin/devdocs saved delete <uuid>
./bin/devdocs technologies
```

A CLI consome exclusivamente a API, sem extrair/buscar conteúdo localmente. Formatos text,
Markdown e JSON preservam indentação. Cores respeitam terminal, `NO_COLOR`, `--no-color` e
`--color auto/always/never`; saídas redirecionadas ficam sem escapes de cor. Segredos não entram
em argumentos de configuração; `set-token` usa entrada sem eco/ stdin. Arquivo em
`os.UserConfigDir()/devdocs/config.json`, criado atomicamente com modo 0600 e diretório 0700.
No Windows, restrinja também ACLs do diretório ao seu usuário. Overrides: `DEVDOCS_CONFIG`,
`DEVDOCS_SERVER`, `DEVDOCS_TOKEN`, `DEVDOCS_TRUST_PRIVATE_NETWORK=true`.
HTTPS é obrigatório para servidor remoto, salvo rede privada explicitamente confiável via
`config set-server <url> --trusted-private-network`. Redirecionamentos da API não são seguidos
para evitar enviar token a outro host.

```sh
make release
# bin/release: macOS, Linux e Windows, amd64/arm64
make cli ARGS='search "Como ler CSV usando Pandas?" --format json'
```

## Desenvolvimento e verificação

```sh
make build
make test          # Go, com detector de corridas
make lint          # go vet + gofmt
make generate      # sqlc generate -f db/sqlc.yaml
make integration   # PostgreSQL Docker, banco dedicado devdocs_test
cd web
npm ci
npm run types      # tipos gerados do OpenAPI
npm test
npm run build
npx playwright install chromium
npm run test:e2e   # exige Compose rodando e .env local
```

Integração usa apenas banco `*_test`, aplica migrações e consulta SQL para provar:
contagem inalterada após busca, +1 após salvar, 0 linhas pelo UUID após DELETE. Testes HTTP usam
fixtures declaradas, nunca respostas simuladas em produção. E2E valida sessão real, CSRF,
salvar sem nova busca, renomear, excluir, falha que mantém diálogo/item, responsividade 360px,
indentação, storage e axe nas páginas principais. A busca real e a CLI são validadas separadamente.
Coleção Postman reutilizável: `api/devdocs.postman_collection.json`.
`python3 scripts/postman-smoke.py` lê segredo local com arquivo temporário 0600, executa pesquisa
real, salva explicitamente e exclui sua consulta de teste. Não publica dados no Postman Cloud.

## Deploy Linux e manutenção

Veja [`docs/deploy-linux.md`](docs/deploy-linux.md), [`docs/architecture.md`](docs/architecture.md)
e [`web/README.md`](web/README.md). Só o Nginx publica porta: `127.0.0.1:8080` por padrão.
API e PostgreSQL não publicam portas. Redes internas separam banco/API; a API tem uma rede de
saída para obter documentação. `WEB_BIND`, `WEB_PORT` e `WEB_ORIGIN` configuram acesso externo.
Exija HTTPS via proxy TLS ou uma rede Tailscale confiável; HTTP remoto exige declaração explícita
`TRUST_PRIVATE_NETWORK=true`. Não exponha PostgreSQL à internet.

Limites Docker: rotação 5 MiB × 2 arquivos por serviço, imagens finais enxutas, containers não-root
para API/web, API read-only, health checks, restart policies e WAL proporcional. O volume de dados
é persistente e nunca é removido automaticamente. Disco de 30 GB exige acompanhar dados, WAL,
backups e cache de builds. Faça backup e confirme restauração antes de qualquer manutenção de
volume. Não use `docker compose down -v` nem limpeza de volumes para manutenção rotineira.
