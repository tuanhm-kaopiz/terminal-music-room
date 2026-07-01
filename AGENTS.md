# Vibe DevKit — AI Agent Instructions

> Bộ công cụ + rules + workflow biến ý tưởng thành code có kiểm soát.
> Đọc file này trước mọi task. Tuân thủ constitution và gates.

## Constitution

Đọc và tuân theo: `constitution/CONSTITUTION.md`

## Workflow bắt buộc

```
Idea → Clarify → Spec → Architecture → Tasks → Code → Review/Test
```

Chi tiết gates: `workflow/PIPELINE.md`

**Quy tắc vàng:** Không nhảy phase. Mỗi phase tạo artifact trong `docs/vibe/{NNN-slug}/`.

## Slash commands (Cursor / Claude)

| Command | Phase | Skill |
|---------|-------|-------|
| `/vibe-idea` | Capture ý tưởng | `vibe-idea` |
| `/vibe-clarify` | Resolve ambiguity | `vibe-clarify` |
| `/vibe-spec` | Requirements | `vibe-spec` |
| `/vibe-architecture` | Tech design | `vibe-architecture` |
| `/vibe-tasks` | Task breakdown | `vibe-tasks` |
| `/vibe-code` | Implementation | `vibe-code` |
| `/vibe-review` | Review & test | `vibe-review` |
| `/vibe-status` | Pipeline status | `vibe-workflow` |
| `/vibe-find-skill` | Skill discovery (post-G2) | `vibe-find-skill` |

**Commands** = user trigger (`/vibe-*`). **Skills** = agent auto-load theo description khi relevant.

`/vibe-find-skill` = discovery layer (như `npm search` cho agent skills) — **sau GATE 2**, không thay pipeline.

## Customization & scale

3 lớp ưu tiên (thấp → cao):

```
Core (devkit)  →  Preset (stack)  →  Project (.vibe/)
   00-09 rules       10-19 rules        90-99 rules
   vibe-* skills     flutter-mvvm       your-domain skills
```

- **Core**: `./install.sh` — cập nhật bằng `--update-core`
- **Preset**: `./install.sh --preset flutter-supabase`
- **Project**: sửa `.vibe/rules/`, `.vibe/skills/` → `./install.sh --project-only .`

Chi tiết: `.vibe/README.md`

## Senior engineer behavior (always)

1. **Clarify before assume** — hỏi khi thiếu info quan trọng
2. **Minimal scope** — chỉ sửa đúng phần cần thiết
3. **Read before write** — đọc code/convention xung quanh trước
4. **Prove it works** — chạy test/lint trước khi báo xong
5. **Gate enforcement** — nếu gate fail, STOP và báo user bước tiếp theo

## Gate override

Chỉ khi user ghi rõ: `GATE OVERRIDE: <lý do>`

Ví dụ hợp lệ:
- `GATE OVERRIDE: P0 hotfix production down`
- `QUICK: yes` (cho thay đổi <10 LOC)

## Artifact numbering

- Folder: `docs/vibe/NNN-{kebab-slug}/`
- `NNN` = 3-digit auto-increment từ folder cao nhất hiện có
- Templates: `templates/*.md` trong devkit root

## Khi bị reject (gate fail)

Trả lời theo format:

```
**GATE {N} FAIL**: {tên gate}
**Reason:** {lý do cụ thể}
**Next step:** Chạy `/vibe-{command}` hoặc cung cấp {input cần thiết}
```

## CLI validate

```bash
./bin/vibe new <slug> --title "Name"   # create feature + set active
./bin/vibe set-active <folder>         # switch active feature
./bin/vibe validate --strict           # CI exit 1 if fail
./bin/vibe trace AC-001                # AC → tasks → review
./bin/vibe list-features               # list (active) marker
python3 scripts/vibe.py --project . gate-check "implement now"
```

Config: `active_feature`, `gate_mode` (strict|normal|relaxed), `require_human_approval` in `vibe.config.yaml`.

Onboarding: `docs/ONBOARDING.md` | Example: `docs/vibe/000-example-user-login/`

## Project-specific overrides

Tạo file `vibe.config.yaml` ở project root để override defaults (stack, test commands, language).
