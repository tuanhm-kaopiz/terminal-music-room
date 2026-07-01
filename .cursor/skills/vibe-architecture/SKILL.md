---
name: vibe-architecture
description: >-
  Produces technical design with ADRs and trade-offs (Phase 3). Use when
  running /vibe-architecture or when spec is approved. May read codebase and
  vibe.config.yaml stack hints.
---

# vibe-architecture — Phase 3

## Role

Staff Architect — components, contracts, ADRs with explicit trade-offs.

## Prerequisites

- GATE 2 passed in `spec.md`
- Read `vibe.config.yaml` (`stack`, `context_files`, preset skills)

## Steps

1. Verify GATE 2 in `spec.md`.
2. Read relevant existing code paths (dependencies in spec).
3. Load preset rules (10-*.mdc) and project rules (90-*.mdc).
4. **Skill gap?** If spec domain needs guidance not in preset/`.vibe/skills` → `/vibe-find-skill` (after step 1 only). **UI-heavy features:** use core `frontend-design` (already installed).
5. Write `architecture.md` from `templates/architecture.md`.
6. Include: component table, data/API contracts, **≥1 ADR** with alternatives + trade-offs.
7. Security/permission table if auth or user data involved.
8. Complete Gate G3 checklist.
9. Tell user: `GATE 3 ✅ — Run /vibe-tasks docs/vibe/NNN-{slug}`

## ADR template (required for major decisions)

```markdown
### ADR-001: {title}
**Context:** ...
**Decision:** ...
**Alternatives considered:** ...
**Trade-offs:** ...
**Consequences:** ...
```

## Anti-patterns

- Implementation code in architecture phase
- Decisions without rejected alternatives
- Missing API contract when spec implies API

## Gate fail response

```
**GATE 3 FAIL**: Architecture approved
**Reason:** Missing trade-off for {decision}
**Next step:** Add ADR section, re-run /vibe-architecture
```

## Reference

`docs/vibe/000-example-user-login/architecture.md`
