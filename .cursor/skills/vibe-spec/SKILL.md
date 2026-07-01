---
name: vibe-spec
description: >-
  Writes testable requirements and acceptance criteria (Phase 2). Use when
  running /vibe-spec or when clarify is done. No implementation details or
  stack choices.
---

# vibe-spec — Phase 2

## Role

Senior BA — testable acceptance criteria, traceability REQ↔AC, edge cases.

## Prerequisites

- GATE 1 passed in `clarify.md`
- Still **no codebase reading**

## Steps

1. Verify GATE 1 in `clarify.md` (no blocking questions).
2. Write `spec.md` from `templates/spec.md`.
3. For each requirement: user story + AC with **Given/When/Then** or measurable outcome.
4. Document edge cases & error scenarios table.
5. Add NFRs if security/perf relevant.
6. Complete Gate G2 checklist.
7. If `require_human_approval: true` in `vibe.config.yaml` → set `**Status:** approved` only after user confirms.
8. Tell user: `GATE 2 ✅ — Run /vibe-architecture docs/vibe/NNN-{slug}`

## AC quality bar

Each AC must be verifiable in review:

```markdown
- [ ] AC-001: Given valid credentials, When user submits login, Then redirect to dashboard
```

IDs: `AC-001`, `REQ-001`, `FR-001` — consistent numbering.

## Anti-patterns (GATE 2 FAIL)

- "Use Redis for sessions" in spec (→ architecture)
- AC: "should work correctly"
- Missing error/edge case table for user-facing flows

## Semantic validation (CLI)

`vibe validate --strict` checks AC wording in strict mode.

## Gate fail response

```
**GATE 2 FAIL**: Spec approved
**Reason:** {AC not testable | contains tech details}
**Next step:** Fix spec.md, re-run /vibe-spec
```

## Reference

`docs/vibe/000-example-user-login/spec.md`
