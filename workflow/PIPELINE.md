# Vibe Coding Pipeline

```
Idea → Clarify → Spec → Architecture → Tasks → Code → Review/Test
  G0      G1       G2        G3           G4      G5        G6
```

## Quick start

```bash
# Trong project đã cài devkit:
/vibe-idea Mô tả ý tưởng của bạn
/vibe-clarify
/vibe-spec
/vibe-architecture
/vibe-tasks
/vibe-code
/vibe-review
```

Kiểm tra tiến độ: `/vibe-status`

---

## Phase 0 — Idea (`/vibe-idea`)

**Mục tiêu:** Capture ý tưởng thô, chưa quyết định tech.

**Input:** Mô tả tự do, link Figma/doc, hoặc bug report.

**Output:** `docs/vibe/{NNN-slug}/idea.md`

### GATE 0 — Idea captured

| # | Criteria | Required |
|---|----------|----------|
| 0.1 | Problem statement rõ (ai, pain gì) | ✅ |
| 0.2 | Success metric hoặc "done looks like" | ✅ |
| 0.3 | Out of scope (ít nhất 1 item) | ✅ |
| 0.4 | Feature slug + folder created | ✅ |

---

## Phase 1 — Clarify (`/vibe-clarify`)

**Mục tiêu:** Loại bỏ ambiguity trước khi spec.

**Input:** `idea.md`

**Output:** `clarify.md` với Q&A đã resolve

### GATE 1 — Clarify complete

| # | Criteria | Required |
|---|----------|----------|
| 1.1 | Không còn open questions blocking spec | ✅ |
| 1.2 | Scope bounded (in/out rõ) | ✅ |
| 1.3 | Actors/users identified | ✅ |
| 1.4 | Assumptions listed explicitly | ✅ |
| 1.5 | Risk flags (nếu có) documented | ⚪ |

**BLOCK nếu:** Còn >0 blocking questions → AI phải hỏi user, không tự assume.

---

## Phase 2 — Spec (`/vibe-spec`)

**Mục tiêu:** Requirements testable, không nói tech stack.

**Input:** `idea.md` + `clarify.md`

**Output:** `spec.md`

### GATE 2 — Spec approved

| # | Criteria | Required |
|---|----------|----------|
| 2.1 | User stories hoặc functional requirements | ✅ |
| 2.2 | Acceptance criteria testable | ✅ |
| 2.3 | Edge cases + error scenarios | ✅ |
| 2.4 | Non-functional reqs (perf, security) nếu relevant | ⚪ |
| 2.5 | Traceability: mỗi AC link về requirement | ✅ |

**BLOCK nếu:** AC không testable hoặc spec chứa implementation details.

---

## Phase 3 — Architecture (`/vibe-architecture`)

**Mục tiêu:** Quyết định "how" — stack, components, data flow.

**Input:** `spec.md` + project context (`AGENTS.md`, codebase)

**Output:** `architecture.md`

### GATE 3 — Architecture approved

| # | Criteria | Required |
|---|----------|----------|
| 3.1 | Component/module breakdown | ✅ |
| 3.2 | Data model hoặc API contracts (nếu có) | ✅ |
| 3.3 | Key decisions + trade-offs (ADR style) | ✅ |
| 3.4 | Dependencies on existing code identified | ✅ |
| 3.5 | Security/permission model (nếu relevant) | ⚪ |

**BLOCK nếu:** Thiếu trade-off cho quyết định quan trọng.

---

## Phase 4 — Tasks (`/vibe-tasks`)

**Mục tiêu:** Breakdown actionable, ordered, estimable.

**Input:** `spec.md` + `architecture.md`

**Output:** `tasks.md`

### GATE 4 — Tasks ready

| # | Criteria | Required |
|---|----------|----------|
| 4.1 | Tasks ordered by dependency | ✅ |
| 4.2 | Mỗi task có clear deliverable | ✅ |
| 4.3 | Mỗi task map về AC/requirement | ✅ |
| 4.4 | Validation step per task (test/lint/check) | ✅ |
| 4.5 | Không có task mơ hồ ("fix stuff") | ✅ |

---

## Phase 5 — Code (`/vibe-code`)

**Mục tiêu:** Implement theo tasks, minimal diff.

**Input:** `tasks.md` + artifacts trước đó

**Output:** Code changes + updated `tasks.md` checkboxes

### GATE 5 — Code complete

| # | Criteria | Required |
|---|----------|----------|
| 5.1 | All tasks marked done or N/A with reason | ✅ |
| 5.2 | Lint/typecheck pass (project commands) | ✅ |
| 5.3 | Tests pass (existing + new if applicable) | ✅ |
| 5.4 | No unrelated changes | ✅ |
| 5.5 | No secrets committed | ✅ |

---

## Phase 6 — Review/Test (`/vibe-review`)

**Mục tiêu:** Senior review + test evidence.

**Input:** All artifacts + code diff

**Output:** `review.md`

### GATE 6 — Ship ready

| # | Criteria | Required |
|---|----------|----------|
| 6.1 | AC checklist: all pass or waived with reason | ✅ |
| 6.2 | Security review (auth, input validation) | ⚪ |
| 6.3 | Performance concerns addressed | ⚪ |
| 6.4 | Review findings resolved or tracked | ✅ |
| 6.5 | Test evidence attached (command + output) | ✅ |

---

## Fast paths

| Scenario | Skip to | Condition |
|----------|---------|-----------|
| Bug fix (P0) | `/vibe-code` | Ghi `GATE OVERRIDE: P0 hotfix` + mô tả bug |
| Tiny change (<10 LOC) | `/vibe-tasks` | User confirm: `QUICK: yes` |
| Spike/POC | `/vibe-architecture` | Ghi rõ throwaway code; `gate_mode: relaxed` |

## CLI helpers

```bash
vibe new <slug>              # Phase 0 scaffold + active_feature
vibe set-active <folder>     # Multi-feature switch
vibe validate --strict       # Pre-PR / CI
vibe trace AC-001            # Traceability check
```

See `docs/ONBOARDING.md` and `docs/vibe/000-example-user-login/`.
