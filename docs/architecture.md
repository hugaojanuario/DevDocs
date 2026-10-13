# Arquitetura

```mermaid
flowchart LR
  CLI[CLI remota Cobra] --> WEB[Nginx :8080]
  UI[React futuro] --> WEB
  WEB --> API[Go / Chi / autenticação]
  API --> SEARCH[Busca: Brave ou índices Sphinx oficiais]
  SEARCH --> FETCH[HTTP com DNS validado / robots / limites]
  FETCH --> ANSWER[Extração + templates + evidências]
  ANSWER --> MEMORY[Cache TTL limitado em RAM]
  API -->|somente salvar explicitamente| PG[(PostgreSQL / sqlc)]
```

Nenhuma consulta entra em logs ou armazenamento automaticamente. PostgreSQL é utilizado
somente por salvos. Sessões e cache vivem em RAM. Reinício invalida sessões e cache.
Token CLI e senha web são segredos distintos, fornecidos pelo ambiente.
