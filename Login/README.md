# ResearchBridge

A Go + SQLite university research collaboration starter with a full-screen ivory/charcoal/gold discovery background, verified accounts, role-specific signup fields, a login-first flow, and separate Student/Teacher dashboards.

**Updating an existing copy? Read UPDATE_BN.md first. Keep your .env and database.**

## What works

- Full-screen responsive authentication page: login first, with a small Register link beneath the login button and an additional signup suggestion when an email has no account. Separate `/login/student` and `/login/teacher` routes remain available; `/signup` is a direct signup route.
- Decorative gold orbital lines and light points animate around the background globe; reduced-motion settings show the still background.
- Institution required for both roles; Department, Academic year and Semester required only for Student signup. Details are saved in SQLite and displayed in My profile.
- In-place, repeatable migration adds profile columns to existing databases without deleting users or projects. Legacy profiles display Not provided for previously uncollected fields.
- Persistent SQLite users, password hashes, verification tokens, sessions, proposals and reviews.
- Email verification required before login; resend links; expiry after 30 minutes; single-use verification.
- Database-backed sessions with logout and expiry after 24 hours.
- Role enforced by the server on every protected action, not just hidden buttons.
- Student: create private drafts, select a verified teacher, submit proposals, revise an abstract after requested changes, see feedback.
- Teacher: view assigned submitted proposals, approve/reject/request changes with feedback, see the review queue.
- Real dashboard counts from the database. No seeded users or invented project statistics.
- CSRF origin/custom-header checks, bounded input, parameterized SQL, bcrypt, hashed bearer tokens, request limits, security headers.
- GitHub Actions checks and integration tests.

This is the initial research proposal workflow. Team membership inside the app, chat, file uploads, milestones, password reset, profile editing and administrator approval are not implemented. Anyone can choose Teacher at signup; email verification proves email ownership, not employment. Add institution/admin approval before using it as a trusted teacher directory.

## 1. Run on Windows / VS Code

Install Go 1.23 or later and Git. Extract this ZIP into its own folder, for example `D:\project-250\ResearchBridge`. Open that folder in VS Code. Keep `go.mod`, `go.sum` and the Go files in the same root folder.

Do not overwrite an existing project or database with this starter without a backup. The profile migration supports earlier versions of this chat's ResearchBridge starter. An unrelated ResearchTheme schema needs separate review.

In the VS Code PowerShell terminal:

```powershell
Copy-Item .env.example .env
go mod download
go run .
```

Open **http://localhost:8080**. Use that exact origin, not `127.0.0.1:8080`, because origin protection checks `APP_URL`.

Alternatively right-click/open `START.ps1` in PowerShell if your local execution policy permits it. No Node, npm, MySQL, PostgreSQL or C compiler is needed. CSS, JavaScript and the image are served by Go and embedded in the built executable. Internet access is needed for the initial Go dependency download.

### First local test (no email service required)

The default `.env` uses `EMAIL_MODE=console`. **It does not send an email.**

1. On the login page choose Teacher and try a new email. When No account found appears, click Sign up as teacher. Enter your name, institution, email and chosen website password.
2. Find the `LOCAL DEVELOPMENT ONLY — verification ...` link in the Go terminal.
3. Open the entire link in your browser and click **Verify email**.
4. Log in through Teacher login.
5. In another browser profile or after logging out, repeat the login-to-signup flow and verification as Student using a different email. Enter Institution, Department, Academic year and Semester.
6. Student: New proposal → Save draft → Submit for review → choose the teacher.
7. Teacher: Review queue → Review proposal → Request changes / Approve / Reject.
8. Student: see the decision and supervisor feedback. A changes-requested proposal can be revised and resubmitted.

An email address belongs to one account and one role. Use two addresses to test both roles. Teacher drafts are not shown in the review queue; students must submit first. You may need to refresh an already-open dashboard after another user makes a change.

The SQLite database is created automatically at `researchbridge.db`. It remains after restarting. Each teammate has their own local database. To share live users and proposals, both need access to one hosted Go server; GitHub does not synchronize databases.

## 2. Real email verification with Gmail SMTP (no custom domain)

Gmail support from the earlier update is included. Your existing SMTP `.env` continues to work. Use the same Gmail account for SMTP_USER and EMAIL_FROM, and a Google App Password with 2-Step Verification enabled:

```dotenv
EMAIL_MODE=smtp
SMTP_USER=youraddress@gmail.com
SMTP_PASS=replace_with_16_character_google_app_password
EMAIL_FROM=ResearchBridge <youraddress@gmail.com>
```

The Gmail sender uses smtp.gmail.com:587 with mandatory STARTTLS, certificate validation and timeouts. The App Password must be configured locally, never committed. Gmail limits apply. No Gmail email was sent from this environment. See https://support.google.com/accounts/answer/185833 for account eligibility.

## 3. Alternative: real email verification with Resend

Create a Resend account/API key, verify a sending domain you control, and configure an allowed sender. The default resend.dev testing sender sends only to the email associated with your Resend account; use your verified domain for normal student/teacher delivery. Check your account's current sending limits in Resend.

Edit **your local `.env`**:

```dotenv
APP_ENV=development
APP_ADDR=127.0.0.1:8080
APP_URL=http://localhost:8080
DB_PATH=researchbridge.db
EMAIL_MODE=resend
RESEND_API_KEY=replace_with_your_real_key_locally
EMAIL_FROM=ResearchBridge <verify@your-verified-domain.example>
```

Replace the example address with a real sender on your verified domain. Stop Go with `Ctrl+C`, then `go run .`. Register and check your inbox/spam folder. A delivered link points to the configured `APP_URL`; a localhost link works only on the computer running the app. It does not become public because an email was delivered.

Go automatically reads `.env`; existing PowerShell environment variables take precedence. If old settings override the file, remove those specific variables (`Remove-Item Env:EMAIL_MODE`, etc.) or reopen the terminal, then restart Go. Never paste keys in chat or commit `.env`.

If delivery fails, the account stays saved and unverified. Fix the email settings, then click **Resend verification**. You do not need to register again. Successful API acceptance does not guarantee inbox delivery; investigate bounces in the provider dashboard if needed.

Reference: https://resend.com/docs/api-reference/emails/send-email
Password hashing reference: https://pkg.go.dev/golang.org/x/crypto/bcrypt

## 4. Suggested two-person contribution roles

| Person | Development responsibility | Branch |
| --- | --- | --- |
| You | Landing page, auth UI, Student/Teacher dashboard UI, responsive CSS | `feature/frontend` |
| Teammate | Go auth, SQLite, verification delivery, proposal/review APIs, integration tests | `feature/backend` |

These are development responsibilities, separate from Student/Teacher account roles in the application. Both developers can test both account roles using separate test accounts. Agree on API field names before changing them. Each developer should make their own commits; do not share GitHub accounts or passwords.

## 5. Push this project to GitHub

Create an **empty GitHub repository** (no README, license or .gitignore generated). In this project's root terminal, substitute your real repository URL:

```powershell
git init
git branch -M main
git add .
git status
git commit -m "Build ResearchBridge authentication and role dashboards"
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPOSITORY.git
git push -u origin main
```

Before committing, `git status` should not list `.env`, `researchbridge.db`, database sidecars or executables. `.env.example`, `go.mod`, **go.sum**, code, tests and web assets should be committed. This ZIP contains no Git history or remote, no credentials and no runtime database.

Invite your teammate via repository **Settings → Collaborators**. After accepting, they run:

```powershell
git clone https://github.com/YOUR_USERNAME/YOUR_REPOSITORY.git
cd YOUR_REPOSITORY
git switch -c feature/backend
Copy-Item .env.example .env
go run .
```

For your work, create `feature/frontend`. On each person's own branch:

```powershell
git add .
git commit -m "Describe your actual change"
git push -u origin feature/frontend
```

Use `feature/backend` in that last command on the backend branch. Open a Pull Request into `main`, let the other teammate review it, and merge after checks pass. Before beginning new work, commit current changes, switch to `main`, `git pull --ff-only`, then create/update your feature branch. Do not force-push shared branches. If your repository already contains a project, clone it and integrate this starter in a separate branch instead of running new-repository commands blindly.

**The project has not been pushed to a GitHub account by this download.** The repository URL/ownership must be selected before a remote push.

## 6. Files

```text
ResearchBridge/
  main.go          # Startup, configuration, SQLite schema, auth/session handlers
  email.go         # Console, Resend or Gmail sender routing
  smtp.go          # TLS-protected Gmail SMTP delivery
  smtp_test.go     # SMTP protocol/config tests
  profile.go       # Role-specific signup validation and safe schema migration
  profile_test.go  # New profile and migration integration tests
  projects.go      # Student proposal and Teacher review handlers
  main_test.go     # Auth, persistence, CSRF and role/ownership integration tests
  go.mod / go.sum  # Go dependencies and checksums
  .env.example     # Example configuration only
  .gitignore       # Keeps secrets, databases and builds out of Git
  START.ps1        # Windows local start helper
  web/
    index.html
    assets/
      app.js
      style.css
      auth.css
      discovery-background.png
      library.png  # Generated professional research-library background
  .github/workflows/go.yml
```

Background produced with the built-in image generation tool: a serene modern university research library, warm oak and glass, daylight, muted forest green and neutrals, without people, text or logos. The asset is included at `web/assets/library.png`; there are no external image/font dependencies.

## 7. Check and build

```powershell
go test ./...
go vet ./...
go build -o researchbridge.exe .
.\researchbridge.exe
```

For race detection (`go test -race ./...`), use a supported platform with the compiler required by Go's race detector. The GitHub workflow uses Ubuntu for this.

## 8. When deploying later

Use a host capable of running Go and keeping a persistent SQLite disk. **GitHub Pages cannot run this Go backend.** Do not commit the live database to GitHub. Configure HTTPS through the host/reverse proxy and set:

```dotenv
APP_ENV=production
APP_ADDR=0.0.0.0:8080
APP_URL=https://your-real-domain.example
DB_PATH=/persistent-disk/researchbridge.db
EMAIL_MODE=resend
RESEND_API_KEY=set_via_your_host_secret_settings
EMAIL_FROM=ResearchBridge <verify@your-real-sending-domain.example>
```

Set `APP_URL` without a trailing slash. The server rejects production without HTTPS and real email mode. HTTPS sets secure session cookies; a local HTTP server is intended for development. Keep the instance behind your HTTPS proxy, protect persistent database backups, and use one Go instance with its SQLite file. Verification links can contain account secrets: do not share terminal links. The in-memory request limit is a small single-server safeguard; public deployments should also use ingress rate limits and monitoring.

Expired database records are checked on use; a long-running deployment can periodically delete expired `sessions` and `verifications`. Startup creates the initial tables and adds missing institution/department/year/semester columns in a transaction. Existing records are preserved. Back up your live database before upgrading.
