# Spec: User Login

**Slug:** `user-login`
**Status:** approved
**Gate G2:** ✅ pass

## Overview

Allow registered users to authenticate with email and password and receive a persistent session.

## User stories / requirements

### REQ-001: Login with valid credentials

**As a** registered user  
**I want** to log in with email and password  
**So that** I can access protected pages

**Acceptance criteria:**

- [x] AC-001: Given valid email and password, When user submits login, Then user is redirected to dashboard and session cookie is set
- [x] AC-002: Given invalid password, When user submits login, Then show generic error "Invalid email or password" and HTTP 401

### REQ-002: Session persistence

**As a** logged-in user  
**I want** my session to survive page refresh  
**So that** I don't have to log in repeatedly

**Acceptance criteria:**

- [x] AC-003: Given active session, When user refreshes page, Then user remains authenticated for up to 7 days

### REQ-003: Logout

**As a** logged-in user  
**I want** to log out  
**So that** my session ends on shared devices

**Acceptance criteria:**

- [x] AC-004: Given logged-in user, When user clicks logout, Then session is invalidated and user sees login page

## Functional requirements

| ID | Requirement | Priority | Trace |
|----|-------------|----------|-------|
| FR-001 | Validate credentials server-side | Must | REQ-001 |
| FR-002 | Issue HTTP-only secure cookie | Must | REQ-002 |
| FR-003 | Invalidate session on logout | Must | REQ-003 |

## Edge cases & error scenarios

| Scenario | Expected behavior |
|----------|-------------------|
| Empty email | Client validation + server 422 |
| Rate limit exceeded | 429 with retry-after |
| Account disabled | Same generic error as wrong password |

## Non-functional requirements

| ID | Category | Requirement |
|----|----------|-------------|
| NFR-001 | Security | No user enumeration via error messages |
| NFR-002 | Performance | Login p95 < 3s |

## Dependencies

- External APIs: none
- Internal modules: user store, session middleware
- Data migrations: sessions table (if not exists)

## Gate G2 checklist

- [x] All requirements have testable AC
- [x] Edge cases documented
- [x] No implementation/tech details (defer to architecture)
- [x] Traceability: AC → REQ mapping complete
