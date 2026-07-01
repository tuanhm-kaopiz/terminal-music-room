# Project-specific Vibe extensions
#
# Files here are YOUR project customizations.
# install.sh syncs them to .cursor/ — devkit updates never overwrite .vibe/
#
# Structure:
#   .vibe/rules/     → project rules (installed as 90-*.mdc)
#   .vibe/skills/    → project skills (installed to .cursor/skills/)
#   constitution.patch.md → append to constitution (optional)
#   skills-installed.log  → audit log for external skill installs (optional)

## Quick start

1. Copy example files:
   cp rules/90-project.example.mdc rules/90-my-project.mdc
   cp skills/project-domain/SKILL.example.md skills/project-domain/SKILL.md

2. Edit with your team conventions

3. Re-sync:
   ./install.sh --project-only /path/to/this/project
   # or from devkit repo:
   ~/vibe-devkit/install.sh /path/to/this/project

## Rule numbering

| Range | Owner | Updated by install? |
|-------|-------|---------------------|
| 00-09 | Vibe DevKit core | Yes (with --update-core) |
| 10-19 | Preset (stack) | Yes when --preset changes |
| 90-99 | Your project (.vibe/) | Never — you own these |

## Skills layering

```
~/.cursor/skills/          # Personal (all projects)
.cursor/skills/vibe-*      # DevKit core (from install)
.cursor/skills/flutter-*   # Preset (from --preset)
.cursor/skills/my-*        # Project (from .vibe/skills/)
```

Agent picks skills by description match — write specific descriptions.
