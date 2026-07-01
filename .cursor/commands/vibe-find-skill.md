# /vibe-find-skill — Skill discovery (after spec)

Discover and install agent skills when preset + `.vibe/skills` are not enough.  
**Not a pipeline phase** — requires GATE 2 (`spec.md`) for the active feature.

## Prerequisites

- `spec.md` exists and GATE 2 passed for `docs/vibe/NNN-{slug}/`
- Read `vibe.config.yaml` → `active_feature`

If missing:

```
**GATE 2 required for skill discovery**
**Reason:** spec.md not ready — domain unknown
**Next step:** /vibe-spec docs/vibe/NNN-{slug}
```

## Role

Skill curator — local lookup first, then `npx skills find`, project-scope install only.

## Steps

1. Extract keywords from `spec.md` (domain, NFRs, integrations).
2. Check preset skills + `.vibe/skills/` — if sufficient, report match (no install).
3. Else run discovery per `vibe-find-skill` skill (delegate to `find-skills` for CLI).
4. Present options with install counts + sources.
5. On approval: install to `.vibe/skills/`, log `.vibe/skills-installed.log`, `install.sh --project-only .`
6. Suggest continuing `/vibe-architecture` or `/vibe-tasks` — do not skip to code.

## Constraints

- No `-g` global install without explicit team policy
- No install without user approval
- Does not replace architecture/tasks artifacts
