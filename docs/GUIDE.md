# Hướng dẫn sử dụng Vibe DevKit

> Tài liệu dành cho **dev trong team** — từ cài đặt đến ship feature với pipeline có gate.  
> Đọc file này trước khi dùng `/vibe-*` lần đầu.

**Liên quan:** [README](../README.md) · [Onboarding nhanh](ONBOARDING.md) · [Pipeline gates](../workflow/PIPELINE.md) · [Ví dụ mẫu](vibe/000-example-user-login/)

---

## 1. DevKit làm gì?

Vibe DevKit **không** thay thế IDE hay AI — nó thêm **quy trình có gate** để AI không nhảy thẳng vào code:

```
Ý tưởng → Làm rõ → Spec → Kiến trúc → Tasks → Code → Review
```

Mỗi bước tạo file trong `docs/vibe/NNN-ten-feature/`. AI bị chặn (hook + rules) nếu bạn bảo "implement luôn" mà chưa có `tasks.md`.

**Bạn vẫn quyết định** — có thể override khẩn cấp bằng `GATE OVERRIDE: ...`.

---

## 2. Cài đặt lần đầu

### 2.1 Cài vào project

```bash
cd /path/to/your-app
npx @hamanhtuan/vibe-devkit install --preset nextjs .
# hoặc: laravel | laravel-vue | flutter-supabase
```

Sau cài đặt, project có:

| Path | Ý nghĩa |
|------|---------|
| `AGENTS.md` / `CLAUDE.md` | AI đọc trước mọi task |
| `.cursor/commands/` | Slash `/vibe-*` trên Cursor |
| `.cursor/skills/` | Skills pipeline + `frontend-design` |
| `docs/vibe/` | Nơi lưu feature artifacts |
| `vibe.config.yaml` | Lệnh test/lint, active feature, gate mode |

### 2.2 Kiểm tra

```bash
vibe validate --strict
# Lần đầu có thể chỉ thấy example 000-example-user-login — OK
```

### 2.3 Cấu hình lệnh test (quan trọng)

Sửa `vibe.config.yaml` cho đúng stack project:

```yaml
commands:
  test: npm test          # hoặc: php artisan test, flutter test, ...
  lint: npm run lint
  typecheck: npm run typecheck
```

Phase **Code** và **Review** dùng các lệnh này — sai config = gate fail giả.

---

## 3. Làm feature mới (workflow đầy đủ)

### Bước 0 — Tạo feature

**Cách A — CLI (khuyến nghị):**

```bash
vibe new thanh-toan --title "Thanh toán đơn hàng"
# → docs/vibe/001-thanh-toan/idea.md
# → set active_feature trong vibe.config.yaml
```

**Cách B — Trong Cursor chat:**

```
/vibe-idea Thêm flow thanh toán cho user đã login, hỗ trợ thẻ và ví nội bộ
```

AI tạo folder `docs/vibe/NNN-slug/` và `idea.md`.

---

### Bước 1 — Clarify (`/vibe-clarify`)

**Khi nào:** Sau `idea.md`, trước khi viết spec.

**Trong Cursor:**

```
/vibe-clarify docs/vibe/001-thanh-toan
```

**Bạn làm gì:**

- Trả lời câu hỏi AI đưa ra (actor, scope, edge case).
- Không để câu hỏi blocking mở — gate G1 fail nếu còn.

**Kết quả:** `clarify.md` — scope in/out, actors, assumptions.

---

### Bước 2 — Spec (`/vibe-spec`)

```
/vibe-spec docs/vibe/001-thanh-toan
```

**Kết quả:** `spec.md` với REQ, AC dạng Given/When/Then.

**Lưu ý:** Spec **không** chọn tech (Redis, React…) — để phase Architecture.

**Kiểm tra:**

```bash
vibe validate 001-thanh-toan
```

---

### Bước 3 — Architecture (`/vibe-architecture`)

```
/vibe-architecture docs/vibe/001-thanh-toan
```

AI đọc codebase + preset skills, viết `architecture.md`: components, API, **ít nhất 1 ADR** có trade-off.

**UI-heavy:** AI tự load skill `frontend-design` (có sẵn trong core).

**Thiếu skill domain?** (sau khi có spec):

```
/vibe-find-skill docs/vibe/001-thanh-toan
```

---

### Bước 4 — Tasks (`/vibe-tasks`)

```
/vibe-tasks docs/vibe/001-thanh-toan
```

**Kết quả:** `tasks.md` — T-001, T-002… có deliverable, lệnh validation, map AC.

**Quan trọng:** Hook Cursor chặn "code ngay" nếu **GATE 4** chưa pass (checklist trong `tasks.md`).

---

### Bước 5 — Code (`/vibe-code`)

```
/vibe-code docs/vibe/001-thanh-toan
```

AI implement từng task, tick `[x]` trong `tasks.md`, chạy test/lint.

**Bạn nên:**

- Review diff từng task.
- Không merge khi task chưa tick hết.

---

### Bước 6 — Review (`/vibe-review`)

```
/vibe-review docs/vibe/001-thanh-toan
```

**Kết quả:** `review.md` — bảng AC pass/fail, output test, quyết định SHIP/HOLD.

---

### Bước 7 — Trước PR

```bash
vibe validate 001-thanh-toan --strict
```

Exit code `0` = pipeline OK cho feature đó.

---

## 4. Xem tiến độ

```
/vibe-status
```

hoặc:

```bash
vibe validate
vibe list-features    # feature (active) đang làm
```

---

## 5. Các tình huống thường gặp

### 5.1 Hotfix production

Trong prompt Cursor, **ghi rõ**:

```
GATE OVERRIDE: P0 login down trên production
Fix session cookie không set khi HTTPS proxy...
```

Được ghi vào `.vibe/override-log.md`. Vẫn nên tạo artifact sau nếu fix lớn.

### 5.2 Sửa nhỏ (<10 dòng)

```
QUICK: yes
Sửa typo label nút Submit trên LoginForm
```

### 5.3 Nhiều feature song song

```bash
vibe set-active 002-billing
/vibe-code docs/vibe/002-billing
```

Hook và gate-check dùng `active_feature` trong `vibe.config.yaml`.

### 5.4 Trace AC về task

```bash
vibe trace AC-003 --feature 001-thanh-toan
```

### 5.5 Brownfield — clone template / codebase sẵn

Dùng **preset brownfield** (không chọn preset sai stack):

| Repo của bạn | Preset |
|--------------|--------|
| [vue-element-plus-admin](https://github.com/kailong321200875/vue-element-plus-admin) | `vue-element-plus-admin` |
| [shadcn-vue-admin](https://github.com/Whbbit1999/shadcn-vue-admin) | `shadcn-vue-admin` |
| FastAPI Python (`app/`, `pytest`) | `fastapi` |

```bash
# shadcn-vue-admin (Tailwind + shadcn — KHÔNG dùng preset element-plus)
git clone https://github.com/Whbbit1999/shadcn-vue-admin.git warehouse-admin
cd warehouse-admin && pnpm install
npx @hamanhtuan/vibe-devkit install --preset shadcn-vue-admin .
vibe new warehouse --title "Quản lý kho"
```

Xem `presets/shadcn-vue-admin/README.md` — merge `AGENTS.md` với template.

**FE + BE tách repo:** cài preset khác nhau mỗi repo; đồng bộ slug feature (`001-warehouse`).

**Thêm stack mới (Go, Django, …):** [PRESETS.md](PRESETS.md)

### 5.6 Brownfield — chưa có preset (legacy)

1. `npx @hamanhtuan/vibe-devkit install .` (core only)
2. `gate_mode: normal` trong `vibe.config.yaml`
3. `.vibe/rules/90-*.mdc` mô tả cấu trúc repo
4. Hoặc đóng góp preset mới theo [PRESETS.md](PRESETS.md)

---

## 6. Khi bị chặn (hook / gate fail)

### Hook: "Vibe DevKit: Chưa có tasks.md..."

**Nguyên nhân:** Bạn nhắn kiểu implement/fix mà GATE 4 chưa pass.

**Cách xử lý:**

1. Chạy đúng phase: `/vibe-tasks` → `/vibe-code`
2. Hoặc dùng `/vibe-code docs/vibe/...` trong prompt (whitelist)
3. Khẩn cấp: `GATE OVERRIDE: <lý do>`

### AI báo GATE FAIL trong chat

Format:

```
**GATE 4 FAIL**: Tasks ready
**Reason:** ...
**Next step:** /vibe-tasks docs/vibe/...
```

Làm đúng **Next step** — không tranh cãi với AI để skip.

### `vibe validate --strict` fail

Đọc **Blockers** trong output — thường là checklist chưa `[x]` ở cuối artifact.

---

## 7. Tùy biến cho team

| Nhu cầu | Cách làm |
|---------|----------|
| Rule riêng (naming, PR…) | `.vibe/rules/90-team.mdc` → `install.sh --project-only .` |
| Skill domain | `.vibe/skills/my-domain/SKILL.md` → `--project-only` |
| Bổ sung constitution | `.vibe/constitution.patch.md` |
| Bật duyệt spec/architecture | `require_human_approval: true` + `status: approved` |

---

## 8. CI & Pull Request

1. Copy `.github/workflows/vibe-validate.yml` vào repo app (nếu chưa có).
2. PR template có checklist Vibe — điền path feature + paste `vibe validate`.

```bash
vibe validate 001-thanh-toan --strict
```

---

## 9. Cursor vs Claude Code

| | Cursor | Claude Code |
|--|--------|-------------|
| Slash commands | `.cursor/commands/vibe-*.md` | `.claude/commands/vibe-*.md` |
| Hook chặn prompt | Có (`beforeSubmitPrompt`) | Không — tự discipline + `AGENTS.md` |
| Skills | `.cursor/skills/` | Cài tương tự qua `install.sh` |

Claude: luôn nhắc `AGENTS.md` hoặc chạy `/vibe-status` đầu session.

---

## 10. FAQ

**Có bắt buộc đủ 7 phase cho mọi thay đổi?**  
Không — hotfix (`GATE OVERRIDE`) và sửa nhỏ (`QUICK: yes`). Production feature nên đủ pipeline.

**Example `000-example-user-login` có implement code không?**  
Không — chỉ artifacts mẫu. Học cấu trúc, không copy sang production.

**`frontend-design` vs `/vibe-find-skill`?**  
`frontend-design` có sẵn cho UI. `find-skills` chỉ khi cần skill khác **sau spec**.

**Cập nhật devkit?**

```bash
npx @hamanhtuan/vibe-devkit install --update-core .
```

**Refresh skill vendored (maintainer):**

```bash
npm run vendor:core-skills
```

---

## 11. Checklist dev mới (ngày 1)

- [ ] Cài preset đúng stack
- [ ] Sửa `commands.test` / `lint` trong `vibe.config.yaml`
- [ ] Đọc `docs/vibe/000-example-user-login/`
- [ ] Chạy thử `vibe new hello-vibe` + `/vibe-clarify` … `/vibe-spec` (có thể dừng sớm để làm quen)
- [ ] Thử `vibe validate --strict`
- [ ] Bookmark file này + [PIPELINE.md](../workflow/PIPELINE.md)

---

## 12. Bản đồ tài liệu

| File | Đọc khi |
|------|---------|
| **docs/GUIDE.md** (file này) | Học cách dùng hàng ngày |
| [ONBOARDING.md](ONBOARDING.md) | Quick start 5 phút |
| [README.md](../README.md) | Tổng quan, CLI, cấu trúc repo |
| [PIPELINE.md](../workflow/PIPELINE.md) | Tiêu chí từng gate (G0–G6) |
| [PRESETS.md](PRESETS.md) | Thêm preset stack mới |
| [MIGRATION.md](MIGRATION.md) | Upgrade 0.1 → 0.2 |
| [000-example-user-login/](vibe/000-example-user-login/) | Mẫu output chuẩn |
| [`presets/vue-element-plus-admin/README.md`](../presets/vue-element-plus-admin/README.md) | Admin Vue template |
| [`presets/fastapi/README.md`](../presets/fastapi/README.md) | FastAPI brownfield |
| [AGENTS.md](../AGENTS.md) | Cho AI agent |
