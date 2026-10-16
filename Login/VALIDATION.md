# Validation — Teacher sky-blue update (2026-10-16)

Passed:

- 26 Go integration/protocol tests with the race detector: `go test -race ./... -timeout 120s`.
- `go vet ./...`, Linux amd64 build and Windows amd64 cross-build.
- Existing registration, institution/academic-field validation and migration, email verification, session lifecycle, role enforcement, proposal reviews, SMTP MIME encoding/STARTTLS and Student dashboard access tests remain passing.
- Recovery tests cover same generic known/unknown response, queued delivery, correct reset purpose in both email formats, hashed token storage, per-address cooldown, invalid/expired/used tokens, password length, old-password rejection, all-session revocation and successful new-password login.
- Workspace tests cover unauthenticated/unassigned denials, Student self-approval prevention, preservation of student evidence during Teacher review, optimistic conflict detection, audit versions, revision chains, invalid URLs/parents/attachment links, private unpublished resources, ownership of edits, evaluation/extension grants, revocation on supervisor reassignment and one-hour meeting conflicts.
- Team/files tests cover opt-in invitation acceptance, prevention of Teacher-added members, membership access, authenticated private downloads, attachment disposition and preserved uploaded content.
- Profile/notification tests cover persistence, preference filtering and suppression of self-notifications.
- Every project/global module kind persists and remains readable after repeat schema initialization.
- jsdom interaction checks cover the Teacher's eight main navigation items, four live overview cards, 20 additional direct module links, hover/tap reveal control, no-team states, feature-preserving team selection, project module navigation, review form permissions and payload, safe rendering, profile saves, meeting forms, pending work queue, paid/unpaid opportunity form and Student responsibility controls.
- Theme checks cover toggling and persistence. Recovery DOM checks cover email submission, password confirmation, removal of the token from URL history, correct payload and success clearing.
- Earlier login/signup, Register link, student-only fields, role/email preservation, orbit SVG references and Student proposal workspace DOM checks remain passing.
- JavaScript syntax checks pass. All referenced local assets exist. HTML uses external scripts and contains no inline styles/event handlers under the strict CSP.

Verification limits:

- Browser visual QA was not performed: no suitable control-browser capability was available under the local Sites workflow. DOM tests do not verify exact pixel layout or animation appearance.
- Live Gmail/Resend delivery was not performed without the user's credentials; the existing transport is retained and its protocol/message tests pass. In-app notifications are separate from account emails.
- PowerShell updater was reviewed but could not be executed on Windows in this Linux environment. Stop the server before its backup/copy operation; it preserves `.env` and the active database and excludes backups from Git.
- The broad research features are database-backed manual forms, record/status reviews and file/link workflows, as described in README.md and TEACHER_UPDATE_BN.md. External calendars, institutional permissions, payments, automatic scoring and publication indexing are not implemented.
- No GitHub changes were pushed. The Student source attribution remains `farhattasnimnashita/ResearchTheme`.

The enclosing execution workspace requires GOFLAGS=-buildvcs=false; ordinary extracted Go checkouts use the documented commands.

## Student-layout refinement (2026-10-16)

Updated Teacher HTML/CSS/JS and navigation; backend handlers and email configuration are unchanged. Teacher, recovery, imported Student and proposal-workspace jsdom checks pass. Visual pixel/animation QA remains unverified; the layout was matched to the supplied Student screenshots and existing CSS.
