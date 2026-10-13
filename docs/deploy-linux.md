# Deploy Linux

1. Instale Docker Engine/Compose v2, Git e Python 3. Clone este repositório em um diretório privado.
2. Execute `python3 scripts/init-env.py`. Guarde `.env` e credenciais em local protegido; modo 0600.
3. Para loopback: `make up`. Para HTTPS: mantenha `WEB_BIND=127.0.0.1`, exponha um proxy TLS
   (Nginx/Caddy) apontando para `http://127.0.0.1:8080` e configure `WEB_ORIGIN=https://docs.seudominio`.
   O proxy deve preservar Origin e ter prazo de resposta de pelo menos 65s. Headers não substituem
   TLS; o cookie Secure é determinado pela origem configurada.
4. Alternativa privada: defina WEB_BIND com IP Tailscale, WEB_ORIGIN com a mesma origem/porta e
   `TRUST_PRIVATE_NETWORK=true` somente nessa rede confiável. Proteja a porta por firewall/ACL.
   CLI com HTTP nessa rede precisa `--trusted-private-network` ao configurar o servidor.
5. Configure provedor Brave para buscas amplas se necessário. Não inclua chaves em VITE_*.
6. Rode `make up`, confira `docker compose --env-file .env -f deployments/docker-compose.yml ps`
   e faça login web/ping da CLI. Não há portas publicadas de API/PostgreSQL.

## Atualizar

Faça backup, revise migrações, obtenha o commit desejado e execute `make up`.
Migrações rodam como job separado antes da API. Uma migração que falha impede inicialização;
não existe fallback silencioso. Nunca rode migração down em produção sem avaliar perda de dados.
`make down` preserva o volume. Reiniciar invalida sessões/cache em RAM, mas mantém consultas salvas.

## Backup manual e disco de 30 GB

```sh
# Execute de um terminal privado; backup contém todas as consultas que você salvou.
docker compose --env-file .env -f deployments/docker-compose.yml exec -T devdocs-postgres \
  pg_dump -U devdocs -d devdocs -Fc > /caminho/protegido/devdocs.dump
chmod 600 /caminho/protegido/devdocs.dump
```

Prefira guardar backups cifrados fora do disco do servidor. Teste restauração em banco separado.
Monitore `df -h`, `docker system df` e `pg_database_size('devdocs')` manualmente. PostgreSQL usa
max_wal_size de 512 MiB; isso é meta operacional, não um limite absoluto de espaço.
Logs giram com 5 MiB/2 arquivos. Imagens de build podem ocupar bem mais que as imagens finais.
Remova apenas imagens/cache de build comprovadamente não usados após verificar espaço e backups.
Não faça limpeza automática; nunca remova volumes PostgreSQL para recuperar espaço.

Excluir uma consulta usa DELETE real. WAL, backups e MVCC/VACUUM podem manter bytes históricos;
escolha políticas de backup/VACUUM com consciência desse limite. DevDocs não promete apagamento
forense nem exclui consultas salvas automaticamente.
