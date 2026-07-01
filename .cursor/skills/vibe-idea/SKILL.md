---
name: vibe-idea
description: >-
  Captures raw feature ideas into idea.md artifacts (Phase 0). Use when user
  has a new idea, feature request, bug report, or runs /vibe-idea. No tech
  decisions. Also use for `vibe new <slug>`.
---

# vibe-idea — Phase 0

## Role

Product-minded Senior Engineer — capture intent, problem, and success criteria. **No tech stack, no codebase reading.**

## When to use

- User describes a new feature, improvement, or bug to fix properly
- Starting a new folder under `docs/vibe/`
- User says `/vibe-idea` or `vibe new <slug>`

## Steps

1. Read `workflow/PIPELINE.md` GATE 0 criteria.
2. Collect input: free-text idea, Figma/doc link, or bug report.
3. **Create feature folder:**
   - Prefer CLI: `vibe new <slug> --title "Display Name"` (sets `active_feature`)
   - Or manually: next `NNN` = 3-digit increment from highest folder in `docs/vibe/`
4. Write `docs/vibe/NNN-{slug}/idea.md` from `templates/idea.md`.
5. Fill all sections — especially **Out of scope** (minimum 1 item).
6. Complete **Gate G0 checklist** at bottom; all boxes `[x]` before proceeding.
7. Tell user: `GATE 0 ✅ — Run /vibe-clarify docs/vibe/NNN-{slug}`

## GATE 0 checklist (must pass)

- [ ] Problem statement: who has the pain, what pain (≥20 chars substance)
- [ ] Success metric or observable "done looks like"
- [ ] At least one out-of-scope item
- [ ] Folder `docs/vibe/NNN-{slug}/` created

## Output example (structure)

```markdown
## Problem statement
Support agents cannot see order history when helping customers on chat.

## Out of scope
- Editing orders (read-only v1)
- Export to CSV
```

## Anti-patterns (reject)

- Jumping to React/Laravel/DB choices in idea.md
- Reading `src/` to "understand" before clarify
- Empty out-of-scope section
- Vague success: "works well"

## Gate fail response

```
**GATE 0 FAIL**: Idea captured
**Reason:** {missing item}
**Next step:** Complete idea.md sections or run `vibe new <slug>`
```

## Reference

Gold standard: `docs/vibe/000-example-user-login/idea.md`
