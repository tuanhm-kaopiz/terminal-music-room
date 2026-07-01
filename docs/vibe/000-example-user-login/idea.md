# Idea: User Login

**Slug:** `user-login`
**Created:** 2026-06-30
**Status:** draft

## Problem statement

End users cannot access protected areas of the app without signing in. Support needs a standard login flow with clear error messages when credentials are wrong.

## Proposed solution (high level)

Add email/password login with session persistence and redirect to the dashboard after success.

## Success looks like

- [ ] Users can log in with valid credentials in under 3 seconds (p95)
- [ ] Invalid credentials show a safe, non-enumerating error message
- [ ] Session persists across browser refresh until logout

## Out of scope

- Social login (Google/GitHub) — future feature
- Password reset flow — separate ticket
- MFA — not in v1

## References

- Figma: (link)
- Docs: workflow/PIPELINE.md
- Related issues: example only

## Raw notes

Gold-standard example feature for Vibe DevKit onboarding. Do not implement in production code.

## Gate G0 checklist

- [x] Problem statement clear
- [x] Success metric defined
- [x] Out of scope listed
- [x] Feature folder created
