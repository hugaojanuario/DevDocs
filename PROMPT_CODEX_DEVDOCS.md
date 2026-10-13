# DevDocs — Prompt de implementação para Codex

## Papel e missão
Atue como arquiteto de software e engenheiro Go sênior. Implemente um produto real chamado **DevDocs**: mecanismo de pesquisa de documentação técnica na internet com respostas objetivas, exemplos de código indentados, explicação por trecho e referências verificáveis. Não é um chatbot. Entregue backend, API REST, CLI, banco e infraestrutura, prontos para integrar a interface React desenvolvida em paralelo. Não crie o frontend nesta etapa.

## Stack obrigatória
- Go (versão estável), Chi, `slog`, `goquery` para extração HTML.
- PostgreSQL em Docker; `pgx` + `sqlc` (SQL explícito; PostgreSQL não é ORM); `golang-migrate`.
- CLI Go com Cobra para macOS, Linux e Windows, consumindo exclusivamente a API remota.
- Docker Compose, multi-stage builds, health checks, Makefile e OpenAPI 3.
- Web prevista: React + TypeScript + Vite; Nginx/servidor estático na porta **8080**, com proxy de `/api/v1` para API privada.

## Funcionalidade principal: busca confiável
Perguntas de exemplo: “Como ler .env em Python?”, “Como ler CSV com Pandas?”, “Como usar CTE com três JOINs no PostgreSQL?”, “Como criar um DAG no Airflow?”, “Como usar volumes no Docker?”.

Pipeline: pergunta → normalização → identificação de tecnologias/versões → termos de busca → consulta a provedor real → classificação de fontes → obtenção segura do conteúdo → extração dos exemplos → estruturação didática → resposta JSON para web/CLI.

1. Crie interface `SearchProvider` intercambiável, implementando pelo menos um provedor real configurável. Não faça scraping de páginas de resultado de buscadores como solução principal. Permita fallback para busca em índices/endpoints públicos oficiais *quando disponível*; se indisponível, retorne erro explícito, sem falsificar resultados.
2. Mantenha cadastro extensível das fontes oficiais: `docs.python.org`, `pandas.pydata.org`, `postgresql.org/docs`, `go.dev/doc`, `pkg.go.dev`, `docs.docker.com`, `airflow.apache.org/docs`, `spark.apache.org/docs`, `docs.getdbt.com`, `fastapi.tiangolo.com`, `kubernetes.io/docs`, `docs.sqlalchemy.org`. Priorize documentação oficial no ranking, mas aceite complementares identificadas.
3. Retorne título, URL exata, domínio, fonte oficial/complementar, tecnologia, versão quando detectável, trecho relevante e blocos de código.
4. Preserve indentação; descarte HTML irrelevante; aplique timeout, tamanho máximo, concorrência limitada e tratamento de erros. Proteja contra SSRF inclusive DNS, redirecionamentos e destinos privados/loopback/link-local/metadados. Respeite robots/termos e não contorne barreiras de acesso.
5. Estruture a resposta em título, tecnologia, “O que é?”, “Como funciona?”, exemplos de código com linguagem, explicações de trechos, observações e fontes. Campos não fundamentados devem ficar vazios/indisponíveis, nunca inventados.
6. **Modo determinístico obrigatório:** extração + templates + evidências citadas, sem alucinações. **IA opcional posteriormente:** interface `AnswerProvider` para Ollama/API compatível, síntese limitada aos trechos coletados, resposta claramente rotulada como assistida por IA, referências preservadas. A V1 funciona sem IA paga. Reconheça que a explicação automática perfeita não é garantida no modo determinístico.

## Persistência seletiva — regra essencial
- **Nenhuma pesquisa é gravada automaticamente** em PostgreSQL, arquivos, logs de consulta, analytics ou histórico. O resultado da busca pode existir apenas em memória transitória, com cache TTL limitado e não persistente.
- O usuário decide clicar em **Salvar consulta** ou usar `--save` no CLI. Só então persista pergunta, título, tecnologia, resposta JSON estruturada, referências, UUID e timestamps.
- Consultas salvas: listar, abrir, filtrar, renomear e excluir.
- **Exclusão física real:** `DELETE FROM saved_queries WHERE id = $1`; nunca soft delete, `deleted_at`, lixeira ou retenção lógica. Remova dependências/caches relacionados. Explique no README que WAL, backups e processos físicos do PostgreSQL podem reter bytes temporariamente; não prometa apagamento forense.
- Nunca apague automaticamente conteúdo salvo. Não crie tabelas de histórico ou telemetria.

### Migração inicial
```sql
CREATE TABLE saved_queries (
  id UUID PRIMARY KEY,
  question TEXT NOT NULL,
  title VARCHAR(255) NOT NULL,
  technology VARCHAR(100),
  response JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
Use migrations versionadas, índices proporcionais, queries SQL tipadas por `sqlc`. Valide tamanho dos payloads salvos.

## API REST — contrato a publicar em `api/openapi.yaml`
Prefixo `/api/v1`:

| Método | Endpoint | Descrição |
|---|---|---|
| POST | `/search` | Pesquisa sem persistência |
| POST | `/saved` | Salva resultado existente explicitamente |
| GET | `/saved` | Lista com paginação e filtros |
| GET | `/saved/{id}` | Abre consulta salva |
| PATCH | `/saved/{id}` | Renomeia |
| DELETE | `/saved/{id}` | Exclusão definitiva |
| GET | `/technologies` | Tecnologias registradas |
| GET | `/sources` | Fontes documentais |
| GET | `/health` | Liveness |
| GET | `/ready` | Readiness |

Exemplo de request de busca:
```json
{"query":"Como ler CSV usando Pandas?","language":"pt-BR","preferred_sources":["official"],"ai_enabled":false}
```

Exemplo **ilustrativo do contrato**, não resultado simulado de produção:
```json
{
  "query":"Como ler CSV usando Pandas?",
  "title":"Leitura de CSV com Pandas",
  "technology":"Python / Pandas",
  "summary":"Utilize read_csv para carregar CSVs.",
  "description":"A função carrega dados em um DataFrame.",
  "examples":[{"language":"python","code":"import pandas as pd\ndf = pd.read_csv('clientes.csv')","explanations":[{"snippet":"pd.read_csv","description":"Realiza a leitura de arquivos CSV."}]}],
  "references":[{"title":"pandas.read_csv","url":"https://pandas.pydata.org/docs/reference/api/pandas.read_csv.html","official":true}],
  "mode":"deterministic","status":"completed"
}
```
Defina no OpenAPI o schema **completo**, inclusive erros, paginação, URLs, IDs, indicação de qualidade/evidência, status e autenticação. O frontend seguirá **esse contrato**, e não o exemplo ilustrativo. Use respostas HTTP corretas, erros JSON padronizados, limites e validação de entrada. `POST /saved` deve aceitar a resposta já consultada, sem disparar outra pesquisa.

## CLI Go + Cobra
Compilar binário `devdocs`, que não realiza busca web local e se conecta à API:
```bash
devdocs config set-server http://localhost:8080
devdocs config set-token
devdocs search "Como utilizar CTE no PostgreSQL?"
devdocs search "Como ler CSV?" --format markdown
devdocs search "Como usar Pandas?" --format json
devdocs search "Como usar JOIN?" --save
devdocs saved list
devdocs saved show <uuid>
devdocs saved delete <uuid>
devdocs technologies
devdocs ping
```
Ofereça formato texto/Markdown/JSON, links, código com identação, cores configuráveis e `--no-color`; respeite saída redirecionada. No modo `--save`, efetue persistência explícita. Arquivo de configuração local com permissões restritas; evite expor tokens nos argumentos e logs.

## Autenticação e segurança
Single-user na V1. Implemente autenticação real, simples e documentada, distinguindo token de CLI e sessão segura web (cookie HttpOnly, Secure quando HTTPS, SameSite e proteção contra CSRF quando aplicável). Nunca incorpore segredos no JavaScript, repositório ou logs. Rate limiting, CORS restritivo, proteção SSRF, validação e saneamento de inputs. Para acesso remoto, exigir HTTPS ou rede privada confiável (como Tailscale); não expor PostgreSQL. Não implemente cadastro multiusuário.

## Docker e restrições de disco
Servidor: **100 GB RAM, 100 vCPUs, 30 GB de disco**. Serviços `devdocs-web`, `devdocs-api`, `devdocs-postgres`. Porta web host configurável com padrão `127.0.0.1:8080:8080`; API e PostgreSQL somente em redes internas; PostgreSQL em volume persistente. Rotação/limites de logs Docker, imagens enxutas, builds multi-stage, health checks, restart policies, cache TTL **em memória** com tamanho máximo, sem armazenamento automático de HTML bruto. Não executar limpeza automática de volumes PostgreSQL. Documentar manutenção segura.

## Organização sugerida
```text
devdocs/
├── cmd/api/main.go
├── cmd/devdocs/main.go
├── internal/api/{handlers,middleware,router,dto}/
├── internal/{config,database,search,extraction,formatter,documentation,bookmarks,security,service}/
├── internal/search/{providers,ranking,sources}/
├── cli/{commands,client,output,config}/
├── db/{migrations,queries,sqlc.yaml}
├── api/openapi.yaml
├── web/README.md
├── deployments/{Dockerfile.api,Dockerfile.web,docker-compose.yml,nginx.conf}
├── tests/
├── scripts/
├── .env.example
├── Makefile
├── go.mod
└── README.md
```
É permitido refinar a estrutura sem misturar responsabilidades.

## Testes e critérios de aceite
- Testes Go unitários e de integração de API/PostgreSQL. HTTP mockado em testes, nunca resultado inventado em produção.
- Cobrir busca e provedores falhando, fontes oficiais, extração, indentação, explicações baseadas em evidência, formatos, segurança SSRF, autenticação, paginação, CLI.
- Demonstrar que `POST /search` **não produz INSERT** e que somente `POST /saved` grava.
- Demonstrar com consulta SQL que `DELETE /saved/{id}` remove fisicamente a linha.
- Provar comunicação da CLI com API remota e execução Docker saudável.
- Documentar limitações de extração sem LLM, variantes de versão e fontes indisponíveis.

## Entregáveis
Código funcional, migrations, `sqlc`, Dockerfiles, Compose, `.env.example`, Makefile (`make build`, `test`, `lint`, `up`, `down`, `migrate`, `cli`), README, diagrama de arquitetura, OpenAPI, exemplos HTTP, guia de build da CLI e deploy Linux. **Prioridade 1: publique `api/openapi.yaml` e esquemas de autenticação estáveis antes de implementar para permitir frontend paralelo.** Desenvolva incrementalmente: contrato → banco/API → pesquisa → salvos → CLI → Docker/testes/docs. Não criar interface React agora.
