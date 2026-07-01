# Thêm preset stack mới

> Hướng dẫn mở rộng Vibe DevKit khi team có **codebase sẵn** (brownfield) hoặc **stack mới** (greenfield).

## Mô hình 3 lớp (không đổi)

```
Core (00-09)     →  Preset (10-19)  →  Project .vibe/ (90-99)
vibe-*, design   →  stack rules     →  team overrides
```

**Preset** = `presets/<tên>/` gồm:

```
presets/my-stack/
├── vibe.config.yaml      # commands, stack, context_files, gate_mode
├── rules/10-*.mdc        # conventions khi implement
├── skills/my-stack/SKILL.md
└── README.md             # upstream link, install one-liner
```

## Greenfield vs brownfield

| | Greenfield | Brownfield |
|--|------------|------------|
| Ví dụ | `nextjs`, `laravel` | `vue-element-plus-admin`, `fastapi` |
| `gate_mode` mặc định | `strict` | `normal` (chuyển `strict` sau) |
| `context_files` | Template docs | Entry points repo thật (`main.py`, `src/router`) |
| Skill focus | Scaffold mới | **Mirror module có sẵn** trong repo |

## Các bước thêm preset

### 1. Copy template

```bash
cp -r presets/fastapi presets/my-stack
# Sửa tên trong vibe.config.yaml: preset: my-stack
```

### 2. `vibe.config.yaml`

```yaml
preset: my-stack
gate_mode: normal          # brownfield
commands:
  test: ...                # lệnh THẬT của repo
  lint: ...
context_files:
  - README.md
  - <entry-file>           # file AI đọc ở architecture phase
skills:
  - my-stack
```

### 3. Rule `rules/10-*.mdc`

- `globs` khớp extension stack (`**/*.py`, `src/**/*.vue`)
- Liệt kê **cấu trúc thư mục** và **anti-patterns**
- Không copy rule preset khác (Laravel rule trên Vue repo = sai)

### 4. Skill `skills/<name>/SKILL.md`

- `description` rõ — agent auto-load khi `/vibe-code`
- Checklist module mới
- “Read before coding” — file/path trong repo mẫu

### 5. Đăng ký `presets/registry.yaml`

```yaml
  my-stack:
    description: One line
    skills: [my-stack]
    rules: [10-my-stack]
    stacks: [tag1, tag2]
    kind: brownfield   # optional: brownfield | greenfield
```

### 6. Verify

```bash
./install.sh --preset my-stack /tmp/test-project
cd /tmp/test-project && vibe validate --strict
```

### 7. Document

- `presets/my-stack/README.md` — link upstream template nếu có
- Cập nhật [GUIDE.md](GUIDE.md) bảng preset nếu preset chính thức

## Team preset riêng (không fork devkit)

**Cách A — trong monorepo devkit:**

```yaml
# registry.yaml
team_presets:
  - name: company-erp
    path: presets/company-erp
```

**Cách B — project only:**

```
your-app/.vibe/rules/90-*.mdc
your-app/.vibe/skills/
./install.sh --project-only your-app
```

Preset repo-level khi **nhiều project cùng stack**; `.vibe/` khi **một repo đặc thù**.

## Fullstack = 2 preset?

Admin Vue + API FastAPI tách repo:

```bash
# Repo frontend
install --preset vue-element-plus-admin .

# Repo backend
install --preset fastapi .
```

Cùng `docs/vibe/` convention; feature slug đồng bộ (`001-warehouse`). Spec frontend/backend có thể 2 folder hoặc 1 spec với section FE/BE — team chọn.

Monorepo: một preset chính + `.vibe/rules` cho phần còn lại.

## Checklist preset chất lượng

- [ ] `commands.*` chạy được trên repo mẫu
- [ ] `context_files` tồn tại sau clone template
- [ ] Skill mô tả brownfield “mirror existing module”
- [ ] README có lệnh install copy-paste
- [ ] Không trùng responsibility preset khác

## Preset hiện có

| Preset | Loại |
|--------|------|
| `flutter-supabase` | greenfield |
| `laravel` | greenfield |
| `laravel-vue` | greenfield |
| `nextjs` | greenfield |
| `vue-element-plus-admin` | brownfield |
| `shadcn-vue-admin` | brownfield |
| `fastapi` | brownfield |

Chi tiết: [presets/README.md](../presets/README.md)
