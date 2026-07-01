# /vibe-tasks — Phase 4: Task Breakdown

Break spec + architecture into ordered, actionable tasks.

## Required input

Feature path: `docs/vibe/NNN-{slug}/`

Hard gate:

```
**GATE 4 FAIL**: Tasks ready
**Reason:** architecture.md missing or GATE 3 not passed.
**Next step:** `/vibe-architecture docs/vibe/NNN-{slug}`
```

Exception: user says `QUICK: yes` for tiny changes (<10 LOC) — create minimal tasks.md only.

## Pre-read

1. `spec.md`, `architecture.md`
2. `templates/tasks.md`
3. `vibe.config.yaml` (validation commands)

## Role

You are a **Tech Lead** — small tasks, clear deliverables, dependency order.

## Steps

1. Verify GATE 3 passed.
2. Create `tasks.md` with phases: Foundation → Core → Integration.
3. Each task: ID, deliverable, maps-to (AC/FR), validation command, target files.
4. Order by dependency. No vague tasks ("fix things", "cleanup").
5. Validate **GATE 4**.

## GATE 4 checklist

- [ ] Ordered by dependency
- [ ] Deliverable + validation per task
- [ ] Maps to AC/requirement
- [ ] No vague tasks

If pass → `GATE 4 ✅ — Run /vibe-code docs/vibe/NNN-{slug}`

## Constraints

- Do NOT implement code in this phase
