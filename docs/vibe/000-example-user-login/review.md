# Review: User Login

**Slug:** `user-login`
**Status:** approved
**Gate G6:** ✅ pass

## Summary

Example feature demonstrating full Vibe pipeline. Artifacts only — no production code in this repo.

## Acceptance criteria verification

| AC ID | Requirement | Status | Evidence |
|-------|-------------|--------|----------|
| AC-001 | Valid login → dashboard | ✅ pass | Example: tasks T-003, T-004 |
| AC-002 | Invalid credentials → generic error | ✅ pass | ADR-002, spec edge cases |
| AC-003 | Session persists 7 days | ✅ pass | T-001, T-002 |
| AC-004 | Logout clears session | ✅ pass | T-005 |

## Code review findings

| # | Severity | Finding | Resolution |
|---|----------|---------|------------|
| — | — | Example artifacts only | N/A |

## Test evidence

```bash
# Command run:
npm test

# Output (summary):
# Example — run project test command from vibe.config.yaml
```

```bash
# Lint:
npm run lint
# Result: pass (when project configured)
```

## Security checklist

- [x] Input validation
- [x] Auth/authz checked
- [x] No secrets in diff
- [x] SQL injection / XSS considered

## Performance notes

- Login p95 target documented in NFR-001

## Ship decision

- [x] **SHIP** — all gates pass (example/reference)

## Gate G6 checklist

- [x] All AC verified or waived with reason
- [x] Review findings resolved or tracked
- [x] Test evidence attached
- [x] No critical security issues open
