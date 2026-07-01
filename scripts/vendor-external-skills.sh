#!/usr/bin/env bash
# Vendor external skills into presets/laravel-vue/skills/
# Equivalent to:
#   npx skills add <repo> --skill <name> -y --copy
# but pins into devkit for offline install + npm package.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEST="$ROOT/presets/laravel-vue/skills"
LOCK="$ROOT/presets/laravel-vue/external-skills.lock.yaml"
TMP="${TMPDIR:-/tmp}/vibe-vendor-skills-$$"

mkdir -p "$DEST"
trap 'rm -rf "$TMP"' EXIT

vendor_one() {
  local name="$1" repo="$2" path="$3"
  local clone_dir="$TMP/$name"

  echo "==> $name  ($repo → $path)"

  rm -rf "$clone_dir"
  git clone --depth 1 --filter=blob:none --sparse "$repo" "$clone_dir" 2>/dev/null
  (
    cd "$clone_dir"
    git sparse-checkout set "$path" 2>/dev/null
  )

  local src="$clone_dir/$path"
  if [[ ! -f "$src/SKILL.md" ]]; then
    echo "ERROR: $src/SKILL.md not found"
    exit 1
  fi

  rm -rf "$DEST/$name"
  mkdir -p "$DEST/$name"
  # Copy skill folder contents (SKILL.md + references)
  if command -v rsync &>/dev/null; then
    rsync -a "$src/" "$DEST/$name/"
  else
    cp -R "$src/." "$DEST/$name/"
  fi

  # Attribution stub
  cat > "$DEST/$name/.vibe-source" <<EOF
repo=$repo
path=$path
vendored=$(date -u +%Y-%m-%dT%H:%M:%SZ)
update: ./scripts/vendor-external-skills.sh
EOF
  echo "    ✓ $DEST/$name"
}

echo "Vibe DevKit — vendor external skills → $DEST"
echo ""

# Parse lock file (simple yaml — no external deps)
current_name="" current_repo="" current_path=""
while IFS= read -r line; do
  line="${line%%#*}"
  line="$(echo "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
  [[ -z "$line" ]] && continue

  if [[ "$line" == "- name:"* ]]; then
    [[ -n "$current_name" && -n "$current_repo" && -n "$current_path" ]] && \
      vendor_one "$current_name" "$current_repo" "$current_path"
    current_name="$(echo "$line" | sed 's/- name: //')"
    current_repo="" current_path=""
  elif [[ "$line" == repo:* ]]; then
    current_repo="$(echo "$line" | sed 's/repo: //')"
  elif [[ "$line" == path:* ]]; then
    current_path="$(echo "$line" | sed 's/path: //')"
  fi
done < "$LOCK"

[[ -n "$current_name" && -n "$current_repo" && -n "$current_path" ]] && \
  vendor_one "$current_name" "$current_repo" "$current_path"

echo ""
echo "✅ Done. Skills in: presets/laravel-vue/skills/"
echo "   Update vibe.config.yaml skills: list if needed."
echo "   Re-install project: ./install.sh --preset laravel-vue <project>"
