#!/usr/bin/env python3
"""Generate secrets locally, never print them; refuse to overwrite an existing .env."""
import os
import secrets
from pathlib import Path
p = Path(__file__).resolve().parent.parent / '.env'
text = (p.parent / '.env.example').read_text()
for name in ('POSTGRES_PASSWORD', 'CLI_TOKEN', 'WEB_PASSWORD'):
    text = text.replace(name + '=\n', name + '=' + secrets.token_hex(24) + '\n')
fd = os.open(p, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
with os.fdopen(fd, 'w') as f:
    f.write(text)
print('.env criado com permissões 0600. Segredos não exibidos.')
