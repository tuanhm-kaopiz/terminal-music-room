---
name: vibe-code
description: >-
  Implements tasks with minimal diff and quality gates (Phase 5). Use when
  running /vibe-code or implementing from tasks.md. Runs tests before done.
  Respects active_feature in vibe.config.yaml.
---

# vibe-code — Phase 5

## Role

Staff Engineer — convention-matching, minimal scope, prove it works.

## Prerequisites

- GATE 4 passed — `tasks.md` exists with checklist complete
- `active_feature` in `vibe.config.yaml` points to this feature (or pass explicit path)
- Read: `tasks.md`, `architecture.md`, `spec.md`, `vibe.config.yaml`

## Steps

1. **Verify GATE 4** — run `vibe validate <slug> --strict` or check tasks gate checklist.
2. Execute tasks **T-001 → T-N** in dependency order.
3. Per task:
   - Implement deliverable only
   - Run task validation command
   - Mark `[x]` in `tasks.md`
   - If blocked: `<!-- BLOCKED: reason -->` and STOP
4. After all tasks: run full `commands.test`, `commands.lint`, `commands.typecheck` from config.
5. Complete GATE 5 checklist in workflow sense (all tasks done, tests green).
6. Tell user: `GATE 5 ✅ — Run /vibe-review docs/vibe/NNN-{slug}`

## Constraints

- **Minimal diff** — no drive-by refactors
- Match patterns in touched directories
- No secrets in committed files
- Load preset skills (e.g. `flutter-mvvm`, `laravel-api`) for stack patterns
- **UI / new screens:** load core `frontend-design` (anthropics) — distinctive layout, typography, motion; not generic AI template UI

## Override paths

- `GATE OVERRIDE: P0 hotfix` — logged to `.vibe/override-log.md`
- `QUICK: yes` — changes under `quick_change_loc_threshold` LOC

## Anti-patterns

- Implementing before tasks.md exists
- Skipping validation commands
- Marking tasks done without running tests

## Gate fail response

```
**GATE 5 FAIL**: Code complete
**Reason:** Task T-002 incomplete / tests failing
**Next step:** Complete task or fix tests before /vibe-review
```

## Hooks

`afterFileEdit` may remind you to update `tasks.md` when code files change.
