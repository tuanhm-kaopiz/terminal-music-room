---
name: vibe-review
description: >-
  Senior review with AC verification and test evidence (Phase 6). Use when
  running /vibe-review or before shipping a feature. Adversarial review.
---

# vibe-review — Phase 6

## Role

Senior Reviewer + QA — verify every AC, security, scope creep, test evidence.

## Prerequisites

- GATE 5 — all tasks `[x]` in `tasks.md`
- Read: all artifacts + git diff

## Steps

1. Verify GATE 5 (tasks complete, lint/test pass).
2. Walk **every AC** in `spec.md` — pass / fail / waived (waive needs explicit reason).
3. Code review: security (auth, input), scope creep, conventions.
4. Run `commands.test` and `commands.lint` — paste output in `review.md`.
5. Use `vibe trace AC-001` to verify task mapping if needed.
6. Ship decision: **SHIP** or **HOLD**.
7. Complete Gate G6 checklist.
8. Run `vibe validate <slug> --strict` for CI readiness.

## Security (HOLD if critical)

- Auth/authz bypass
- Secrets in diff
- SQLi/XSS on new user input paths

## Review.md requirements

- AC verification table with evidence column
- Test command + output summary
- Gate G6 checklist all `[x]`

## Anti-patterns

- Waiving AC without reason
- "Looks good" without test output
- Skipping security on auth features

## Gate fail response

```
**GATE 6 FAIL**: Ship ready
**Reason:** AC-002 not verified
**Next step:** Add test evidence or fix implementation
```

## Reference

`docs/vibe/000-example-user-login/review.md`
