# ResearchBridge Faculty Directory and Student Rating

## What's included
- Removed the generated woman portrait from teacher banner. The original sky-blue styling remains.
- Faculty directory seeded from 9 user-supplied SUST CSE screenshots; 18 teachers; phone and email shown as provided; website URLs not provided, so deliberately not fabricated.
- Faculty data: `GET /api/faculty-directory` (authenticated users including teammates via shared Supabase DB).
- Student-only rating API: `GET /api/faculty-ratings` and `POST /api/faculty-ratings` (CSRF-protected by existing backend; verified students submit).
- Scores: supervision, monitoring, guidelines, helpfulness, understanding (1-5 stars each). One rating per student+faculty; submitting again updates it.
- New section in the student dashboard: `Faculty Ratings`. Teacher dashboard does not display ratings.

## How to apply
1. Back up your working project and existing .env. Do NOT commit .env to GitHub.
2. Open this ZIP's Login folder in VS Code and copy your existing .env if needed.
3. Keep `DATABASE_URL` pointed at the same Supabase Postgres DB for both teammates. If absent, data will only save in local SQLite and NOT automatically sync.
4. Run `go run .` in the `Login` directory. Startup creates faculty_directory and faculty_ratings tables if needed and inserts the 18 screenshot records once. No other teachers or existing ratings are overwritten.
5. Sign in as a verified student, open /dashboard/student and scroll to Faculty Directory & Student Ratings.
6. To verify on Supabase: use SQL editor (`SELECT name,email FROM researchbridge.faculty_directory;`), or query through backend. For teammate access, use the same connected database, not a shared secret sent in chat.
7. Before going live, verify phone numbers, names, and current designations with university; these were transcribed from screenshots, not independently verified.

## Notes
- Database records are not linked automatically to registered teacher user IDs. Teammate can create mapping separately if needed.
- No real students' ratings were invented. All rating values start empty.
- The previous frontend assets and existing API routes remain in place.
- Go test could NOT complete in the tool environment because required Go modules cannot be downloaded here; run `go test ./...` locally before deployment.
- Ratings are student-only in the Go app API; do not expose these tables through public Supabase policies or public SQL endpoints.
