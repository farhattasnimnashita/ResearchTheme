# Academic profile and private supervision review update

## Added
- Student dashboard > My Profile: year-wise CGPA (1st–6th year), skills, and research interests editor.
- Student-owned API:
  - `GET /api/student-academic-profile`
  - `PUT /api/student-academic-profile`
- Teacher-only team review API:
  - `GET /api/projects/{id}/team-academics`
  - It returns member academic records only if the signed-in teacher is the teacher assigned to that project's supervision request.
- Teacher supervision review dialog now displays each member's year-wise CGPA, skills, and research interests before Accept/Decline.
- SQLite startup migration and Supabase/PostgreSQL schema tables are included.

## Privacy
Academic records are not included in the public student directory or general project listing. The team academic endpoint checks the logged-in role and exact `projects.teacher_id` before returning the information. A different teacher receives HTTP 403.

## Database
The Go application creates these tables automatically on startup:
- `student_academic_records`
- `student_skills`
- `student_academic_interests`

For Supabase, these tables are created in the private `researchbridge` schema by the Go application. Do not move them to `public`.

## Important validation note
The Go source was formatted and the changed JavaScript files passed `node --check`. A full `go test ./...` could not be completed in this environment because required Go modules were not cached and outbound network access was unavailable. Run `go test ./...` locally after dependencies are available, then test the profile save and teacher review flow with two teacher accounts to verify the 403 privacy boundary.


## Mandatory supervision-request validation

- The requester must have at least one year-wise CGPA record and one private Academic Profile skill. The browser redirects to My Profile when either is missing.
- The backend validates the project owner and every confirmed `team_members` row before accepting `POST /api/projects/{id}/supervisor`.
- If a teammate is incomplete, the API rejects the request and names the affected student(s); each student must update their own account.
- General profile `expertise` is not treated as the private `student_skills` list for this check.
- Teacher Assigned Teams and Pending Work / Revision now merge the project-member endpoint with the selected-teacher academic endpoint, deduplicate by user ID, and show the project title, complete team list and submitter.
