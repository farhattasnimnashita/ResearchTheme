# ResearchBridge — Go research collaboration platform

Go + SQLite + embedded HTML/CSS/JavaScript. No frontend build or external database server is required. This update adds the sky-blue Teacher dashboard, persistent light/dark themes, email password recovery and a shared research workspace. The imported maroon Student dashboard remains available after Student login.

See **TEACHER_UPDATE_BN.md** for installation into the user's existing project and a full feature walkthrough. **APPLY_UPDATE.ps1** preserves the target `.env` and active database and creates a local backup. **QUICK_START_BN.md** covers first-time setup. No real credentials or database are included.

## Run

Go 1.23+ is required. For an existing installation, keep its `.env`, stop the running server and copy all Go sources plus the complete web folder using the updater. New installations copy `.env.example` to `.env` and supply Gmail App Password settings (or configure Resend). Run `go run .` and open APP_URL, normally http://localhost:8080. Startup refuses to silently select console mode when neither `.env` nor EMAIL_MODE exists.

## Workflows

- Role-specific register/login, institution and Student academic fields, email verification and sessions.
- Student dashboard imported from `farhattasnimnashita/ResearchTheme` (`dashboard.html` and `static/dashboard.css`), with real profile/proposal data; protected proposal workspace at `/workspace/student`.
- Sky-blue Teacher dashboard matching the Student layout: geometric hover/tap banner, four live summary cards, Current Projects, Next Meeting and activity panels. Eight primary sidebar links plus an expandable supervision-tools list expose every module directly, including clear no-team states. Project selection preserves the current feature.
- Supervision/proposal approval, corrections and rejection with feedback; assigned team members; opt-in team invitations.
- Project tasks, milestones, progress, contributions, paper/experiment notes, submissions and versions, feedback, blockers, mentoring, final submissions and corrections, publication tracking and private showcases.
- Meeting requests and scheduling with one-hour conflict checks across the supervisor's assigned projects. Meeting notes and separately assignable follow-up tasks.
- Criterion-based evaluation and deadline-extension decisions require responsibility delegated by the Student project owner to the current supervisor. Grants do not transfer when the supervisor changes. This is project-level delegation, not an institution-admin appointment.
- Private file uploads (8 MB/file, 64 MB/project) stored in SQLite, protected downloads, preserved earlier uploads and supporting URLs.
- Teacher-owned resource posts: paid/unpaid opportunities, lab/group pages (members/projects/recruitment in the description), campus research problems; publish to verified users. Unpublished posts and editing remain owner-only.
- Profile bio/interests/expertise and in-app request/submission/meeting notification preferences.
- Browser-persisted light/dark preference on auth and dashboards, respecting the OS preference initially.
- Forgot/reset password uses the same Gmail SMTP or Resend sender as verification. Queued email delivery avoids response-time account enumeration; random hashed single-use links expire in 30 minutes, per-address cooldown, invalidation of all sessions on reset, no automatic login or verification.

All module actions persist in SQLite and use server-side project/role checks. Teacher review preserves the student's original submission content, while review status/feedback are stored separately and each saved version is retained in an audit history. Optimistic revisions prevent overwriting concurrent edits.

## Scope

Research records use structured titles/status/dates, text, links, optional assignees, progress/criteria and review feedback. They are not integrations with an external calendar, publisher, lab administration or recruitment/payment system. Meeting links are supplied by the users. Progress is maintained by the supervisor; there is no automatic member scoring. Notifications are in-app, with email reserved for account verification and recovery. Files are private attachments, not executable previews; old versions are preserved rather than deleted. Showcases remain project-private.

Teacher signup remains self-selected; email verification proves mailbox access, not employment. Institution/admin approval is outside this starter. Existing legacy profile fields not previously collected remain blank. Do not merge unrelated Go server sources into this project.

## Security and persistence

Parameterized SQL; bcrypt; hashed verification/session/reset tokens; bounded request and file sizes; strict same-origin/custom-header checks; project membership and teacher assignment checks on every workspace/file action; safe text rendering; attachment-only downloads; CSP without inline scripts; notification preferences and authenticated access; transactional migrations and atomic password resets. A small in-process recovery queue is appropriate for local development; queued jobs may be lost on process exit, in which case the user requests another link.

Keep `.env`, `*.db`, `*.db-*`, executables and `backup*/` out of Git. Update backups include local secrets and database files and must not be committed. Stop the server before copying a live SQLite database.

## Validation

`go test -race ./...`, `go vet ./...`, `go build .`, Windows amd64 cross-build, JavaScript syntax and jsdom interaction tests. See VALIDATION.md for outcomes and verification limits. The enclosing build workspace requires GOFLAGS=-buildvcs=false; extracted normal Go projects do not.
