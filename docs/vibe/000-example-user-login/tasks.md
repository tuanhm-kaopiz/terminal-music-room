# Tasks: User Login

**Slug:** `user-login`
**Status:** complete
**Gate G4:** ✅ pass

## Task list

### Phase 1 — Foundation

- [x] **T-001:** Add session model and migration
  - Deliverable: Session table + model
  - Maps to: AC-003, FR-002
  - Validation: `npm test -- session`
  - Files: `src/models/session`, `migrations/`

- [x] **T-002:** Session middleware
  - Deliverable: Middleware loads user from cookie
  - Maps to: AC-003
  - Validation: `npm test -- middleware`
  - Files: `src/middleware/session.ts`

### Phase 2 — Core implementation

- [x] **T-003:** Login endpoint
  - Deliverable: POST /api/auth/login
  - Maps to: AC-001, AC-002, FR-001
  - Validation: `npm test -- auth.login`
  - Files: `src/auth/controller.ts`

- [x] **T-004:** Login UI
  - Deliverable: Form + error display
  - Maps to: AC-001, AC-002
  - Validation: `npm run lint`
  - Files: `src/auth/LoginForm.tsx`

### Phase 3 — Integration & polish

- [x] **T-005:** Logout endpoint + UI
  - Deliverable: POST /api/auth/logout + button
  - Maps to: AC-004, FR-003
  - Validation: `npm test -- auth.logout`
  - Files: `src/auth/`

## Progress

| Phase | Total | Done | Blocked |
|-------|-------|------|---------|
| 1 | 2 | 2 | 0 |
| 2 | 2 | 2 | 0 |
| 3 | 1 | 1 | 0 |

## Gate G4 checklist

- [x] Tasks ordered by dependency
- [x] Each task has deliverable + validation
- [x] Each task maps to AC/requirement
- [x] No vague tasks
