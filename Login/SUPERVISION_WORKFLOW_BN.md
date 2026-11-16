# ResearchBridge: Two-stage supervision workflow

## Student team
1. Create/select an existing team project.
2. Invite the teammate. The invitation becomes membership only after the invited student accepts it.
3. Every confirmed team member must complete their own **Academic Profile** before requesting supervision:
   - At least one year-wise CGPA record (year 1–6, CGPA 0.00–4.00).
   - At least one skill in the private Academic History & Skills section.
4. Open **Supervisor** and choose the project plus a verified teacher.
5. Click **Send Supervision Request**. The browser checks the requester's own academic profile first. If it is incomplete, no request is sent and the student is redirected to **My Profile → Academic History & Skills**.
6. The backend independently validates the project owner and every confirmed teammate. If any member is incomplete, the request is rejected and the response identifies the member(s) who need to update their own profile. A frontend-only bypass cannot send the request.
7. Wait for the supervisor's Accept/Decline decision. Every project member receives a dashboard notification.
8. After acceptance, open **Proposal**, complete the research details, click **Save Draft**, then **Submit for Review**.
9. Every team member receives the full-proposal decision and feedback notification.

## Supervisor
1. Open **Supervision Requests**.
2. Open a pending request and inspect the project title, all confirmed team members, and each member's year-wise CGPA, skills and research interests.
3. Academic information is returned only to the teacher assigned to that exact request. Other teachers must receive HTTP 403.
4. Accept or decline (a decline requires a reason).
5. Accepted requests unlock the student's full proposal form.
6. Full proposals appear separately under **Proposal Review** and can be approved, returned for changes, or rejected.
7. In **Assigned Teams**, see the project title and all confirmed members. In **Pending Work / Revision**, see the record title, project title, all team members, submitter and revision parent.

## Database migration
- SQLite: academic profile tables and supervision fields are created/migrated automatically at startup.
- Supabase/PostgreSQL: the startup schema migration creates the academic profile tables and adds supervision fields automatically.
- Existing projects start with `supervision_status = none`; send a new supervision request to use the two-stage workflow.

## Important distinction
- General Profile `Expertise` and private Academic Profile `Skills` are separate data fields.
- Supervision-request validation uses private `student_skills` and `student_academic_records`, not the general `researcher_profiles.expertise` field.
- A pending invitation is not a confirmed member; only an accepted invitation is stored in `team_members`.

## Verification note
JavaScript syntax checks and `gofmt` completed. The member-list and academic-validation SQL queries were smoke-tested against a two-student SQLite fixture. `go test ./...` could not run in this environment because Go dependencies were not cached and outbound access to `proxy.golang.org` was unavailable. Run `go test ./...` locally before deployment.
