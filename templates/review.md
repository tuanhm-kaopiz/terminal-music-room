# Review: {feature-name}

**Slug:** `{slug}`
**Status:** draft | approved
**Gate G6:** ⬜ pending | ✅ pass | ❌ fail

## Summary

Brief: what was built, scope of changes.

## Acceptance criteria verification

| AC ID | Requirement | Status | Evidence |
|-------|-------------|--------|----------|
| AC-001 | | ✅ pass / ❌ fail / ⚠️ waived | test name, screenshot, manual step |

## Code review findings

| # | Severity | Finding | Resolution |
|---|----------|---------|------------|
| | critical / major / minor | | fixed / tracked / waived |

## Test evidence

```bash
# Command run:
npm test

# Output (summary):
# X passed, 0 failed
```

```bash
# Lint:
npm run lint
# Result: pass
```

## Security checklist

- [ ] Input validation
- [ ] Auth/authz checked
- [ ] No secrets in diff
- [ ] SQL injection / XSS considered

## Performance notes

- 

## Ship decision

- [ ] **SHIP** — all gates pass
- [ ] **HOLD** — issues remain (list below)

### Remaining issues

1. 

## Gate G6 checklist

- [ ] All AC verified or waived with reason
- [ ] Review findings resolved or tracked
- [ ] Test evidence attached
- [ ] No critical security issues open
