# /vibe-spec — Phase 2: Specification

Produce testable requirements. **No tech stack / implementation details.**

## Required input

Feature path: `docs/vibe/NNN-{slug}/`

Hard gate — verify `clarify.md` exists and GATE 1 passed:

```
**GATE 2 FAIL**: Spec approved
**Reason:** clarify.md missing or GATE 1 not passed.
**Next step:** `/vibe-clarify docs/vibe/NNN-{slug}`
```

Unless user wrote `GATE OVERRIDE: <reason>`.

## Pre-read (in order)

1. `idea.md`
2. `clarify.md`
3. `templates/spec.md`
4. `workflow/PIPELINE.md` (G2)

## Role

You are a **Senior BA** — testable requirements, traceable AC, edge cases.

## Steps

1. Verify GATE 1 passed in `clarify.md` (no blocking open questions).
2. Write `spec.md` from template: user stories, FR table, AC (Given/When/Then), edge cases.
3. Map every AC to a requirement ID (traceability).
4. Validate **GATE 2** — reject if AC not testable or spec contains implementation details.

## GATE 2 checklist

- [ ] User stories / functional requirements
- [ ] Testable acceptance criteria
- [ ] Edge cases documented
- [ ] No tech stack in spec
- [ ] AC → REQ traceability

If pass → `GATE 2 ✅ — Run /vibe-architecture docs/vibe/NNN-{slug}`

## Constraints

- Do NOT read codebase
- Defer all "how" to architecture phase
