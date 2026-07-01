---
name: vibe-workflow
description: >-
  Vibe DevKit spec-first pipeline with gates. Use when building features,
  vibe coding, pipeline status, active feature, or when user mentions
  idea/clarify/spec/architecture/tasks workflow.
---

# Vibe Workflow Skill

Orchestrate the 7-phase gated pipeline for vibe coding.

## Pipeline

```
Idea → Clarify → Spec → Architecture → Tasks → Code → Review/Test
 G0      G1       G2        G3           G4      G5        G6
```

Full gate definitions: `workflow/PIPELINE.md`

## CLI quick reference

```bash
vibe new <slug> --title "Name"    # Create feature + set active
vibe set-active <folder>          # Switch active feature
vibe validate --strict            # CI / pre-merge check
vibe trace AC-001                 # AC → tasks → review
vibe list-features                # Show (active) marker
```

## Active feature

Set in `vibe.config.yaml`:

```yaml
active_feature: docs/vibe/003-user-auth
```

Hooks and `gate-check` use this — not just the latest folder.

## Gate modes (`vibe.config.yaml`)

| Mode | Behavior |
|------|----------|
| `strict` | All gate checklists + semantic rules must pass |
| `normal` | Blockers on missing files; warnings on partial gates |
| `relaxed` | Code allowed if `tasks.md` exists |

## When to use

- New feature idea → `/vibe-idea` or `vibe new`
- User tries to skip to code → GATE FAIL + suggest correct phase
- Status check → `/vibe-status` or `vibe validate`
- Multi-feature project → `vibe set-active` first

## Slash commands

| Command | Phase | Skill |
|---------|-------|-------|
| `/vibe-idea` | 0 | vibe-idea |
| `/vibe-clarify` | 1 | vibe-clarify |
| `/vibe-spec` | 2 | vibe-spec |
| `/vibe-architecture` | 3 | vibe-architecture |
| `/vibe-tasks` | 4 | vibe-tasks |
| `/vibe-code` | 5 | vibe-code |
| `/vibe-review` | 6 | vibe-review |
| `/vibe-status` | — | vibe-workflow |
| `/vibe-find-skill` | — (post-G2) | vibe-find-skill |

Skill discovery (`vibe-find-skill`) runs **after spec** — not a pipeline phase. See `03-skill-discovery.mdc`.

## Gate enforcement

1. Detect phase via `vibe validate` logic
2. Wrong phase → GATE FAIL format from `AGENTS.md`
3. Override: `GATE OVERRIDE: <reason>` or `QUICK: yes` (audited in `.vibe/override-log.md`)

## Artifact location

`docs/vibe/NNN-{slug}/` — templates in `templates/`

Gold standard: `docs/vibe/000-example-user-login/`

## Senior engineer defaults

- Clarify before assume
- Minimal scope on code
- Run tests before done
- Read `constitution/CONSTITUTION.md`
