# Integração da interface

Frontend React/TypeScript/Vite será entregue por handoff. Não há interface React nesta etapa.
Contrato estável: `../api/openapi.yaml` (JSON é YAML válido). Base relativa `/api/v1`.

Login: POST `/auth/session` com `{ "password": "..." }`. Cookie HttpOnly fica no navegador;
o JSON retorna `csrf_token`. GET `/auth/session` recupera o token após reload.
Nas mutações enviar `X-CSRF-Token` e usar `credentials: 'same-origin'`. Logout: DELETE da sessão.
Nunca armazenar senha/token CLI no bundle ou localStorage.

Busca: POST `/search`. Exibir campos vazios como indisponíveis; renderizar código como texto,
nunca HTML. POST `/saved` recebe `{ "response": <resposta já obtida>, "title": "opcional" }`.
Paginação `/saved?page=1&page_size=20&q=...&technology=...`. DELETE remove fisicamente.
Qualidade e evidências fazem parte do contrato; textos extraídos preservam o idioma original.
Nginx serve `web/dist` na porta 8080 e encaminha API privadamente.
