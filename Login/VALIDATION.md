# Validation — updated authentication and profiles

Passed:

- 18 Go integration/protocol tests, including role-specific profile validation/storage, student-only fields, old user-schema migration and preservation of account data/projects/sessions, repeat startup, missing-account signup guidance, wrong-password/role handling, verification, session isolation, proposal reviews and SMTP STARTTLS handling.
- `go test -race ./... -count=1`.
- `go vet ./...`.
- Linux and Windows amd64 source builds.
- JavaScript syntax check (`node --check`).
- DOM interaction checks in jsdom: default login page, a visible Register link for both roles plus an extra missing-account signup CTA, institution for both roles, student-only academic fields, preservation of entered fields when changing roles, numeric year/semester API payload, registration confirmation and existing-account errors without signup suggestions.
- All linked HTML/CSS/JS image/style/script assets exist locally.

Limits:

- Visual browser QA and live Gmail/Resend delivery were not performed. This environment has no suitable browser-control plugin and no user email credentials.
- Legacy accounts retain their login data but previously uncollected profile fields remain blank; the profile page labels those as Not provided. Editing legacy profiles is not included.
- The project is packaged with a GitHub workflow; the ResearchTheme source was read through the GitHub connector; no GitHub changes were pushed.

The enclosing execution workspace required `GOFLAGS=-buildvcs=false` for builds. Normal Go commands are documented for the extracted project.

Latest frontend-only update: direct Register navigation, preserved email/role, cleared passwords, return to login, valid decorative SVG orbit references, reduced-motion CSS, JavaScript syntax, and Linux/Windows builds checked. SVG motion timing/appearance was not visually browser-tested. Go/backend files were unchanged in this update.


ResearchTheme integration: protected Student HTML route, unauthenticated redirects, Teacher isolation, Student proposal workspace and Teacher page continuity tested. DOM checks cover imported navigation, safe profile rendering, live proposal and feedback summaries, workspace links and removal of inline JavaScript/fake Meet URL. Visual browser QA remains unavailable.
