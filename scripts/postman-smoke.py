#!/usr/bin/env python3
"""Explicit live search/save/delete contract check. Never export search results or credentials."""
import json
import os
import subprocess
import tempfile
from pathlib import Path
root = Path(__file__).resolve().parent.parent
config = dict(line.split('=', 1) for line in (root / '.env').read_text().splitlines()
              if line and not line.startswith('#'))
with tempfile.TemporaryDirectory(prefix='devdocs-postman-') as directory:
    path = Path(directory) / 'env.json'
    fd = os.open(path, os.O_CREAT | os.O_WRONLY, 0o600)
    with os.fdopen(fd, 'w') as handle:
        json.dump({'name': 'local-private', 'values': [
            {'key': 'cli_token', 'value': config['CLI_TOKEN'], 'enabled': True},
            {'key': 'base_url', 'value': config['WEB_ORIGIN'], 'enabled': True},
        ]}, handle)
    result = subprocess.run(['postman', 'collection', 'run',
                             str(root / 'api/devdocs.postman_collection.json'), '-e', str(path)], cwd=root)
    raise SystemExit(result.returncode)
