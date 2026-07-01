# Clarify: User Login

**Slug:** `user-login`
**Status:** complete
**Gate G1:** ✅ pass

## Resolved questions

| # | Question | Answer | Decided by |
|---|----------|--------|------------|
| 1 | Which actors can log in? | Registered end users only | user |
| 2 | Session duration? | 7 days sliding expiration | user |
| 3 | Brute-force protection? | Rate limit 5 attempts / 15 min per IP | user |

## Open questions (blocking)

| # | Question | Owner | Blocking? |
|---|----------|-------|-----------|

## Scope

### In scope

- Email + password login form
- Server-side credential validation
- HTTP-only session cookie
- Logout

### Out of scope

- Registration, password reset, OAuth

## Actors / users

| Actor | Role | Key actions |
|-------|------|-------------|
| End user | Authenticated customer | Login, logout |
| Anonymous visitor | Unauthenticated | View login page |

## Assumptions

1. Users already have accounts created by admin or registration (out of scope).
2. HTTPS is enforced in all environments.

## Risks & constraints

| Risk | Impact | Mitigation |
|------|--------|------------|
| Credential stuffing | Account takeover | Rate limiting + generic errors |

## Gate G1 checklist

- [x] No blocking open questions
- [x] Scope bounded (in/out explicit)
- [x] Actors identified
- [x] Assumptions listed
