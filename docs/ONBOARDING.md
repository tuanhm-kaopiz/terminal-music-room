# Onboarding — Vibe DevKit for Dev Teams

> Quick start 5 phút. **Hướng dẫn đầy đủ:** [GUIDE.md](GUIDE.md)

## 5-minute start

```bash
# 1. Install into your project
npx @hamanhtuan/vibe-devkit install --preset nextjs .

# 2. Create a feature
vibe new user-profile --title "User Profile"

# 3. Run pipeline (in Cursor chat)
/vibe-clarify docs/vibe/001-user-profile
/vibe-spec docs/vibe/001-user-profile
/vibe-architecture docs/vibe/001-user-profile
/vibe-tasks docs/vibe/001-user-profile
/vibe-code docs/vibe/001-user-profile
/vibe-review docs/vibe/001-user-profile

# 4. Validate before PR
vibe validate --strict
```

## Study the gold standard

Open `docs/vibe/000-example-user-login/` — complete artifacts for every phase.

## Multi-feature projects

```bash
vibe list-features          # shows (active)
vibe set-active 002-billing
```

## Gate override (emergency only)

```
GATE OVERRIDE: P0 production login down
```

Logged to `.vibe/override-log.md`.

## Core skills (mọi install)

| Skill | Khi dùng |
|-------|----------|
| `vibe-*` | Pipeline phases |
| `frontend-design` | UI mới, redesign — architecture & code |
| `find-skills` | Tìm skill khác — qua `/vibe-find-skill` sau spec |

Nguồn: [anthropics/skills](https://github.com/anthropics/skills) — refresh: `npm run vendor:core-skills`

## Team customization

| What | Where |
|------|-------|
| Team rules | `.vibe/rules/90-*.mdc` |
| Domain skills | `.vibe/skills/*/SKILL.md` |
| Constitution append | `.vibe/constitution.patch.md` |
| Test/lint commands | `vibe.config.yaml` |

Sync after edits:

```bash
./install.sh --project-only .
```

## Gate modes

| Mode | When to use |
|------|-------------|
| `strict` | Default — production teams |
| `normal` | Brownfield migration |
| `relaxed` | Spike only (still needs tasks.md) |

## Human approval (optional)

```yaml
require_human_approval: true
```

Spec and architecture must have `**Status:** approved` before code in strict validation.

## CI

Enable `.github/workflows/vibe-validate.yml` — fails PR if pipeline blockers exist.

## Skill discovery (optional, after spec)

Not a pipeline phase. Use when preset + `.vibe/skills` are not enough **after `spec.md` exists**.

```bash
/vibe-find-skill docs/vibe/001-user-profile   # in Cursor, after /vibe-spec
```

Policy:

1. Local skills first (preset → `.vibe/skills/`)
2. Then `npx skills find <keywords>` via vendored `find-skills`
3. Install to `.vibe/skills/` only (`--copy`, no `-g`)
4. Log in `.vibe/skills-installed.log` + `install.sh --project-only .`

Maintainers refresh vendored `find-skills`: `npm run vendor:core-skills`

See `.cursor/rules/03-skill-discovery.mdc` and `SOURCES.md`.

## Help

- **Usage guide:** [GUIDE.md](GUIDE.md)
- Pipeline: `workflow/PIPELINE.md`
- Migration: `docs/MIGRATION.md`
- Changelog: `CHANGELOG.md`
