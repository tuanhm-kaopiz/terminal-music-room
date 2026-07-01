# /vibe-clarify — Phase 1: Clarify

Resolve ambiguity before writing spec. Output: `clarify.md`

## Required input

Feature path: `docs/vibe/NNN-{slug}/` or `idea.md` path.

If missing:

```
**GATE 1 FAIL**: Clarify complete
**Reason:** Missing feature path or idea.md.
**Next step:** `/vibe-clarify docs/vibe/001-my-feature` or run `/vibe-idea` first.
```

## Pre-read (in order)

1. `constitution/CONSTITUTION.md`
2. `docs/vibe/NNN-{slug}/idea.md`
3. `workflow/PIPELINE.md` (G1 criteria)

## Role

You are a **Business Analyst + Senior Engineer** — find gaps, ask blocking questions, do not assume.

## Steps

1. Load `idea.md`. List all ambiguities, missing actors, unclear scope.
2. **Ask user blocking questions** — use AskQuestion tool if available, else numbered list in chat.
3. If blocking questions remain unanswered → **STOP**. Do not write final clarify.md.
4. When resolved, create/update `clarify.md` from `templates/clarify.md`.
5. Fill: resolved Q&A, scope in/out, actors, assumptions, risks.
6. Validate **GATE 1** checklist.

## GATE 1 — blocking rules

**BLOCK** if any open question is marked blocking.

- [ ] No blocking open questions
- [ ] Scope bounded
- [ ] Actors identified
- [ ] Assumptions explicit

If pass → `GATE 1 ✅ — Run /vibe-spec docs/vibe/NNN-{slug}`

## Constraints

- Do NOT write spec.md in this phase
- Do NOT read codebase (requirements only)
- Do NOT guess answers to blocking questions
