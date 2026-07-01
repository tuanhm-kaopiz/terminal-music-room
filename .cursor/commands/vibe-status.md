# /vibe-status — Pipeline Status

Show current Vibe DevKit pipeline status for a feature or all features.

## Input

Optional: feature path `docs/vibe/NNN-{slug}/`. If omitted, scan all under `docs/vibe/`.

## Output format

For each feature, report:

```
## {NNN-slug}

| Phase        | Artifact          | Gate | Status |
|--------------|-------------------|------|--------|
| Idea         | idea.md           | G0   | ✅/⬜/❌ |
| Clarify      | clarify.md        | G1   | ...    |
| Spec         | spec.md           | G2   | ...    |
| Architecture | architecture.md   | G3   | ...    |
| Tasks        | tasks.md          | G4   | ...    |
| Code         | (tasks checkboxes)| G5   | ...    |
| Review       | review.md         | G6   | ...    |

**Current phase:** {name}
**Next command:** `/vibe-{command} docs/vibe/NNN-{slug}`
**Blockers:** {list or "none"}
```

## Logic

- File missing → ⬜ pending
- File exists, gate checklist incomplete → ⬜ in progress
- Gate checklist complete in file → ✅ pass
- Explicit FAIL marker in file → ❌ fail

Read gate checklists from each artifact footer section.

## No mutations

This command is read-only — do not create or edit files.
