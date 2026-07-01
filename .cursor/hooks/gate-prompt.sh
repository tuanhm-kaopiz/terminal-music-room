#!/usr/bin/env bash
# Vibe DevKit hook — block "just code it" without pipeline artifacts
# Event: beforeSubmitPrompt
set -euo pipefail

ROOT="$(pwd)"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEVKIT_SCRIPT=""

# Resolve vibe.py: project scripts/ or devkit bundled path
for candidate in \
  "$ROOT/scripts/vibe.py" \
  "$SCRIPT_DIR/../../scripts/vibe.py" \
  "$SCRIPT_DIR/../../../scripts/vibe.py"; do
  if [[ -f "$candidate" ]]; then
    DEVKIT_SCRIPT="$candidate"
    break
  fi
done

input=$(cat)

# Extract prompt — prefer python3 (no jq dependency)
extract_prompt() {
  python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
except Exception:
    print('')
    sys.exit(0)
for key in ('prompt', 'text', 'message', 'content', 'user_message'):
    v = data.get(key)
    if isinstance(v, str) and v.strip():
        print(v)
        break
else:
    # fallback: dump all string values
    def walk(obj):
        if isinstance(obj, str) and len(obj) > 10:
            yield obj
        elif isinstance(obj, dict):
            for x in obj.values():
                yield from walk(x)
        elif isinstance(obj, list):
            for x in obj:
                yield from walk(x)
    for s in walk(data):
        print(s)
        break
" <<< "$input" 2>/dev/null || echo ""
}

prompt=$(extract_prompt)

if [[ -z "$prompt" ]]; then
  echo '{"permission":"allow"}'
  exit 0
fi

if [[ -z "$DEVKIT_SCRIPT" ]] || ! command -v python3 &>/dev/null; then
  # Fail open if tooling missing
  echo '{"permission":"allow"}'
  exit 0
fi

result=$(python3 "$DEVKIT_SCRIPT" --project "$ROOT" gate-check "$prompt" 2>/dev/null || echo '{"permission":"allow"}')

permission=$(python3 -c "import json,sys; print(json.loads(sys.argv[1]).get('permission','allow'))" "$result" 2>/dev/null || echo "allow")

if [[ "$permission" == "deny" ]]; then
  python3 -c "
import json, sys
r = json.loads(sys.argv[1])
out = {
    'permission': 'deny',
    'user_message': r.get('user_message', 'Vibe DevKit gate blocked this prompt.'),
    'agent_message': r.get('agent_message', 'Complete the vibe pipeline or use GATE OVERRIDE.')
}
print(json.dumps(out))
" "$result"
  exit 2
fi

echo '{"permission":"allow"}'
exit 0
