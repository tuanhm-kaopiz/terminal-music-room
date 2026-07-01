#!/usr/bin/env bash
# Vibe DevKit hook — remind to update tasks.md when editing implementation files
# Event: afterFileEdit
set -euo pipefail

ROOT="$(pwd)"
input=$(cat)

extract_paths() {
  python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
except Exception:
    sys.exit(0)
paths = []
for key in ('file', 'path', 'filePath', 'uri'):
    v = data.get(key)
    if isinstance(v, str):
        paths.append(v)
edits = data.get('edits') or data.get('files') or []
if isinstance(edits, list):
    for e in edits:
        if isinstance(e, dict):
            for k in ('file', 'path', 'filePath'):
                if k in e and isinstance(e[k], str):
                    paths.append(e[k])
for p in paths:
    p = p.replace('file://', '')
    if p and not p.endswith('.md'):
        print(p)
" <<< "$input" 2>/dev/null || true
}

mapfile -t edited < <(extract_paths)
if [[ ${#edited[@]} -eq 0 ]]; then
  exit 0
fi

# Skip if only docs/vibe artifacts edited
code_edited=false
for f in "${edited[@]}"; do
  case "$f" in
    *docs/vibe/*|*templates/*|*.md) continue ;;
    *) code_edited=true; break ;;
  esac
done
$code_edited || exit 0

tasks_file=""
if [[ -f "$ROOT/vibe.config.yaml" ]]; then
  active=$(python3 -c "
import re, pathlib
p = pathlib.Path('$ROOT/vibe.config.yaml')
t = p.read_text(encoding='utf-8', errors='replace')
m = re.search(r'^active_feature:\s*(.+)$', t, re.M)
print(m.group(1).strip().strip('\"') if m else '')
" 2>/dev/null || echo "")
  if [[ -n "$active" ]]; then
    tasks_file="$ROOT/$active/tasks.md"
    [[ -f "$tasks_file" ]] || tasks_file="$ROOT/$active/tasks.md"
  fi
fi

if [[ -z "$tasks_file" || ! -f "$tasks_file" ]]; then
  latest=$(find "$ROOT/docs/vibe" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort | tail -1)
  [[ -n "$latest" && -f "$latest/tasks.md" ]] && tasks_file="$latest/tasks.md"
fi

[[ -f "$tasks_file" ]] || exit 0

python3 -c "
import json, sys
msg = (
    'Vibe DevKit: Code file edited. Update tasks.md checkboxes for completed work '
    f'({sys.argv[1]}). Run validation commands before marking done.'
)
print(json.dumps({'user_message': msg}))
" "$tasks_file" 2>/dev/null || true

exit 0
