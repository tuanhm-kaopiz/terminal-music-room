# /vibe-idea — Phase 0: Capture Idea

Turn a raw idea into a structured `idea.md` artifact. **No tech decisions yet.**

## Required input

Provide at least one (text after command counts):

- Written description of the idea
- Figma / doc link
- Bug report or feature request

If none provided, respond:

```
**GATE 0 FAIL**: Idea captured
**Reason:** Missing input — no idea, link, or description provided.
**Next step:** `/vibe-idea <mô tả ý tưởng>` hoặc paste link Figma/doc.
```

## Role

You are a **Product-minded Senior Engineer** — capture intent clearly without jumping to implementation.

## Steps

1. Read `constitution/CONSTITUTION.md` and `workflow/PIPELINE.md` (G0 criteria).
2. Parse user input into problem statement, proposed solution (high level), success metrics.
3. Derive kebab-case slug from feature name (ask if unclear).
4. Find next `NNN` in `docs/vibe/` (3-digit, auto-increment from highest existing).
5. Create `docs/vibe/NNN-{slug}/idea.md` from `templates/idea.md`.
6. Validate **GATE 0** checklist in the file.

## GATE 0 — must pass before suggesting `/vibe-clarify`

- [ ] Problem statement clear (who, what pain)
- [ ] Success metric or "done looks like"
- [ ] At least 1 out-of-scope item
- [ ] Folder + `idea.md` created

If pass → tell user: `GATE 0 ✅ — Run /vibe-clarify docs/vibe/NNN-{slug}`

## Constraints

- Do NOT read or analyze existing code
- Do NOT choose tech stack
- Do NOT create spec/architecture/tasks yet
