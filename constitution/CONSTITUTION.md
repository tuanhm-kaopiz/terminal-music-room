# Vibe DevKit Constitution

> Nguyên tắc bất biến — mọi phase trong workflow phải tuân theo.
> Cập nhật file này khi team/project thay đổi chuẩn cốt lõi.

## 1. Spec trước, code sau

- Không implement khi chưa có spec được approve (trừ hotfix P0 có ghi rõ lý do).
- Mọi thay đổi scope phải quay lại phase Clarify hoặc Spec.

## 2. Senior engineer behavior

- **Clarify trước khi assume** — hỏi khi thiếu thông tin quan trọng, không đoán mò.
- **Minimal scope** — chỉ sửa đúng phần cần thiết; không refactor lan man.
- **Match conventions** — đọc code xung quanh trước khi viết mới.
- **Prove it works** — chạy test/lint liên quan trước khi báo hoàn thành.
- **No secrets in repo** — không commit `.env`, keys, credentials.

## 3. Artifact-driven workflow

Mỗi feature lưu artifacts tại `docs/vibe/{NNN-feature-slug}/`:

| File | Phase | Mô tả |
|------|-------|-------|
| `idea.md` | Idea | Ý tưởng thô, problem statement |
| `clarify.md` | Clarify | Câu hỏi đã resolve, scope, assumptions |
| `spec.md` | Spec | Requirements + acceptance criteria |
| `architecture.md` | Architecture | Tech decisions, ADRs, data flow |
| `tasks.md` | Tasks | Task breakdown có thứ tự |
| `review.md` | Review | Kết quả review + test |

## 4. Gate policy

- AI **không được** nhảy phase khi gate chưa PASS.
- Gate FAIL → dừng, báo lỗi rõ ràng, hướng dẫn bước tiếp theo.
- User có thể override bằng cách ghi rõ: `GATE OVERRIDE: <lý do>`.

## 5. Language & communication

- Spec/artifacts: tiếng Anh hoặc tiếng Việt (nhất quán trong một feature).
- Code comments: theo convention project (mặc định: English).
- Commit message: imperative mood, mô tả "why" ngắn gọn.

## 6. Quality bar

- Acceptance criteria phải testable (Given/When/Then hoặc checklist).
- Architecture phải ghi rõ trade-offs đã cân nhắc.
- Task phải đủ nhỏ để implement trong 1 session.
- Review phải có evidence: test output, lint, checklist ticked.
