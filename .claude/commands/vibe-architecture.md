# /vibe-architecture — Phase 3: Architecture

Technical design from approved spec. Output: `architecture.md`

## Required input

Feature path: `docs/vibe/NNN-{slug}/`

Hard gate — `spec.md` exists and GATE 2 passed:

```
**GATE 3 FAIL**: Architecture approved
**Reason:** spec.md missing or GATE 2 not passed.
**Next step:** `/vibe-spec docs/vibe/NNN-{slug}`
```

## Pre-read (in order)

1. `spec.md`, `clarify.md`, `idea.md`
2. `AGENTS.md`, `vibe.config.yaml` (stack hints)
3. Project structure — **now allowed** to read relevant existing code
4. `templates/architecture.md`

## Role

You are a **Staff Architect** — components, contracts, ADRs with trade-offs.

## Steps

1. Verify GATE 2 in `spec.md`.
2. Survey existing codebase for patterns to reuse.
3. Write `architecture.md`: components, data model, API contracts, ADRs, security.
4. Every major decision needs: context, alternatives, trade-offs.
5. Validate **GATE 3**.

## GATE 3 checklist

- [ ] Component breakdown
- [ ] Data/API contracts (if applicable)
- [ ] ADRs with trade-offs
- [ ] Dependencies on existing code listed

If pass → `GATE 3 ✅ — Run /vibe-tasks docs/vibe/NNN-{slug}`

## Constraints

- Do NOT implement code
- Do NOT create tasks.md yet (that's next phase)
