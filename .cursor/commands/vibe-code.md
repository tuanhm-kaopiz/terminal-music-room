# /vibe-code — Phase 5: Implementation

Implement tasks in order with minimal diff and quality gates.

## Required input

Feature path: `docs/vibe/NNN-{slug}/`

Hard gate:

```
**GATE 5 FAIL**: Code complete
**Reason:** tasks.md missing or GATE 4 not passed.
**Next step:** `/vibe-tasks docs/vibe/NNN-{slug}`
```

Override only with `GATE OVERRIDE: <reason>` (e.g. P0 hotfix).

## Pre-read (in order)

1. `tasks.md`, `architecture.md`, `spec.md`
2. `vibe.config.yaml` — test/lint commands
3. Relevant source files per task

## Role

You are a **Staff Engineer** — safe, deterministic, convention-matching implementation.

## Execution

1. Verify GATE 4. Process tasks in dependency order.
2. For each task:
   - Implement deliverable
   - Run validation command from task
   - Mark `[x]` in `tasks.md`; add `<!-- BLOCKED: reason -->` if stuck
3. After all tasks: run full `commands.test` and `commands.lint` from config.
4. Validate **GATE 5**.

## GATE 5 checklist

- [ ] All tasks done or N/A with reason
- [ ] Lint/typecheck pass
- [ ] Tests pass
- [ ] No unrelated changes
- [ ] No secrets in diff

If pass → `GATE 5 ✅ — Run /vibe-review docs/vibe/NNN-{slug}`

## Constraints

- Minimal scope — task deliverable only
- Update `tasks.md` in place during execution
- Do not skip tests before claiming complete
