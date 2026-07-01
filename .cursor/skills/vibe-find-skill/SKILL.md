---
name: vibe-find-skill
description: >-
  Discovers and installs agent skills after spec domain is clear (GATE 2+).
  Use during /vibe-architecture when preset/.vibe skills are insufficient,
  or when user asks for a skill for a known domain. Not for skipping the
  vibe pipeline. Delegates to find-skills + npx skills CLI after local lookup.
---

# vibe-find-skill — Skill discovery layer

> **Not core workflow.** Pipeline stays: Idea → Clarify → Spec → Architecture → Tasks → Code → Review.  
> This skill is **npm search for agent skills** — only after the domain is known from `spec.md`.

## When to use

| ✅ Use | ❌ Do not use |
|--------|----------------|
| GATE 2 passed; architecture needs a domain skill not in preset | Before spec (domain unknown) |
| Preset + `.vibe/skills` checked, gap remains | Instead of `/vibe-spec` or `/vibe-tasks` |
| User asks "skill for X" **and** active feature has `spec.md` | Random `npx skills add -g` without approval |
| Architecture phase: ADR needs specialized guidance (e2e, a11y, etc.) | To bypass gates |

## Decision flow

```
1. Verify GATE 2 — spec.md exists for active feature
2. Extract domain keywords from spec.md (NFRs, stack, integrations)
3. Local lookup (in order):
   a. **Core:** `frontend-design` (UI), preset skills, `find-skills` (discovery only)
   b. presets/registry.yaml + installed preset skills (.cursor/skills/)
   c. .vibe/skills/
   d. vibe.config.yaml skills: list
4. If match found → recommend existing skill; NO external install
5. If gap → load find-skills skill + run:
   npx skills find <keywords>
   (check https://skills.sh/ leaderboard first)
6. Present options: name, source, install count, install command
7. On user approval only → install PROJECT scope:
   npx skills add <owner/repo@skill> -y --copy -a cursor
   Target: .vibe/skills/<name>/ then ./install.sh --project-only .
8. Log install in .vibe/skills-installed.log
```

## Install command (team-safe)

```bash
# Project scope only — never -g without team policy
npx skills add vercel-labs/agent-skills@react-best-practices -y --copy -a cursor
mkdir -p .vibe/skills/react-best-practices
# move/copy into .vibe/skills/ per skills CLI output, then:
./install.sh --project-only .
```

Append log:

```markdown
## 2026-06-30 — react-best-practices
- Feature: docs/vibe/003-checkout
- Reason: AC-005 requires performance audit patterns
- Source: vercel-labs/agent-skills
- Installed by: @dev
```

## Quality bar (from find-skills)

Before recommending:

1. Prefer 1K+ installs on [skills.sh](https://skills.sh/)
2. Prefer official sources: `vercel-labs`, `anthropics`, known maintainers
3. Skeptical if repo &lt;100 stars or unknown author
4. Maintainer must review before merge if PR adds new skill

## Integration with architecture

When writing `architecture.md`, if a gap skill is installed:

- Reference it in **Implementation notes**: "Follow `.cursor/skills/<name>`"
- Do not duplicate skill content in architecture.md

## Gate interaction

- Installing a skill **does not** replace GATE 3/4/5
- Still run `/vibe-tasks` and `/vibe-code` after architecture

## Delegation

For search/install mechanics, follow `.cursor/skills/find-skills/SKILL.md` after local lookup fails.

## Anti-patterns

- Installing 5 skills "just in case"
- Global install (`-g`) on shared team repos
- Discovery before clarify/spec
- Using skills to implement without `tasks.md`

## Reference

- Policy: `.cursor/rules/03-skill-discovery.mdc`
- Upstream: `SOURCES.md` → `find-skills`
- Pipeline: `vibe-workflow` skill
