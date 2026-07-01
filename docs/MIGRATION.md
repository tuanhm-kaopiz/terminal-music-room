# Migration Guide

## 0.1.x → 0.2.0

### 1. Update devkit

```bash
npx @hamanhtuan/vibe-devkit@latest install --update-core .
# or
./install.sh --update-core /path/to/project
```

### 2. Add `active_feature` to `vibe.config.yaml`

```yaml
active_feature: docs/vibe/001-your-feature
```

Set via CLI:

```bash
vibe set-active 001-your-feature
```

### 3. Hooks

`hooks.json` now includes `afterFileEdit` and `failClosed: true`. Merge or reinstall:

```bash
./install.sh --update-core .
```

### 4. New CLI commands

| Old workflow | New |
|--------------|-----|
| Manual NNN folder | `vibe new my-feature` |
| Guess latest feature | `vibe set-active` |
| Manual AC trace | `vibe trace AC-001` |

### 5. CI

Copy `.github/workflows/vibe-validate.yml` or add step:

```yaml
- run: python3 scripts/vibe.py validate --strict
```

### 6. `gate_mode`

Default remains `strict`. For legacy projects mid-migration:

```yaml
gate_mode: normal  # or relaxed temporarily
```

### 7. Breaking changes

None for existing artifacts. Validator is stricter in `strict` mode — run `vibe validate --strict` and fix warnings/blockers.
