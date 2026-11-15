# ResearchBridge: Two-stage supervision workflow

## Student team
1. Create/select an existing team project.
2. Open **Supervisor** and choose the project plus a verified teacher.
3. Click **Send Supervision Request**. Only the project title and team membership are presented as the initial request.
4. Wait for the supervisor's Accept/Decline decision. Every project member receives a dashboard notification.
5. After acceptance, open **Proposal**, complete the research details, click **Save Draft**, then **Submit for Review**.
6. Every team member receives the full-proposal decision and feedback notification.

## Supervisor
1. Open **Supervision Requests**.
2. Open a pending request and inspect the project title and team members.
3. Accept or decline (a decline requires a reason).
4. Accepted requests unlock the student's full proposal form.
5. Full proposals appear separately under **Proposal Review** and can be approved, returned for changes, or rejected.

## Database migration
- SQLite: new `supervision_status` and `supervision_feedback` columns are added automatically at startup.
- Supabase/PostgreSQL: the startup schema migration adds the same columns automatically.
- Existing projects start with `supervision_status = none`; send a new supervision request to use the new workflow.

## Validation note
JavaScript syntax checks and `gofmt` completed. `go test ./...` could not run in this environment because Go dependencies were not cached and outbound access to `proxy.golang.org` was unavailable. Run `go test ./...` locally before deployment.
