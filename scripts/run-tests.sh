#!/usr/bin/env bash
# Run Vibe DevKit tests (creates .venv if pytest missing)
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

python3 scripts/vibe.py validate 000-example --strict

if python3 -m pytest --version &>/dev/null; then
  python3 -m pytest tests/ -q
elif [[ -x "$ROOT/.venv/bin/pytest" ]]; then
  "$ROOT/.venv/bin/pytest" tests/ -q
else
  echo "Creating .venv for tests..."
  python3 -m venv .venv
  .venv/bin/pip install -q -r requirements-dev.txt
  .venv/bin/pytest tests/ -q
fi

echo "✅ All tests passed"
