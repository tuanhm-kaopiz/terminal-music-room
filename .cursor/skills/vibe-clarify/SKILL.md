---
name: vibe-clarify
description: >-
  Resolves ambiguity before spec (Phase 1). Use when running /vibe-clarify,
  when idea.md exists but scope is unclear, or blocking questions remain.
  STOP and ask user — never assume blocking answers.
---

# vibe-clarify — Phase 1

## Role

BA + Senior Engineer — surface gaps, ask blocking questions, document scope and actors.

## Prerequisites

- `idea.md` exists and GATE 0 passed
- Do **not** read codebase yet

## Steps

1. Read `idea.md`. List every ambiguity (actors, permissions, edge cases, integrations).
2. For each **blocking** question → ask user explicitly. **STOP** until answered.
3. Write `clarify.md` from `templates/clarify.md`.
4. Fill: Resolved Q&A, Scope in/out, Actors table, Assumptions, Risks.
5. **Open questions (blocking)** table: header only if none remain (no placeholder rows).
6. Complete Gate G1 checklist — all `[x]`.
7. Tell user: `GATE 1 ✅ — Run /vibe-spec docs/vibe/NNN-{slug}`

## GATE 1 blockers

- Any row in "Open questions (blocking)" with real content
- Missing actors
- Scope not bounded (in AND out lists)

## Question quality

Ask specific questions:

- ❌ "Any preferences?"
- ✅ "Should guest users see prices without login? (yes/no)"

## Anti-patterns

- Assuming default auth model
- Writing spec-level AC in clarify
- Leaving "(TBD)" in blocking table without asking user

## Gate fail response

```
**GATE 1 FAIL**: Clarify complete
**Reason:** Blocking question unanswered: {question}
**Next step:** Answer the question above, then re-run /vibe-clarify
```

## Reference

`docs/vibe/000-example-user-login/clarify.md`
