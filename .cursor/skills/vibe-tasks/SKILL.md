---
name: vibe-tasks
description: >-
  Breaks spec and architecture into ordered actionable tasks (Phase 4). Use
  when running /vibe-tasks or when architecture is approved. No code yet.
---

# vibe-tasks — Phase 4

## Role

Tech Lead — small tasks (~1 session each), clear deliverables, validation commands, AC traceability.

## Prerequisites

- GATE 3 passed in `architecture.md`
- Read `spec.md` + `architecture.md`

## Steps

1. Verify GATE 3 in `architecture.md`.
2. Write `tasks.md` from `templates/tasks.md`.
3. For each task **T-00N**:
   - **Deliverable:** concrete output
   - **Maps to:** AC-xxx, FR-xxx
   - **Validation:** exact command from `vibe.config.yaml` or task-specific test
   - **Files:** paths to create/edit
4. Order by dependency (foundation → core → polish).
5. Complete Gate G4 checklist.
6. Tell user: `GATE 4 ✅ — Run /vibe-code docs/vibe/NNN-{slug}`

## Task size rule

Each task completable in one focused session. Split if >5 files or multiple unrelated concerns.

## Anti-patterns

- "Fix stuff", "Implement feature"
- Missing validation command
- No AC mapping

## Gate fail response

```
**GATE 4 FAIL**: Tasks ready
**Reason:** Task T-003 has no deliverable / validation
**Next step:** Fix tasks.md, re-run /vibe-tasks
```

## Reference

`docs/vibe/000-example-user-login/tasks.md`
