# ResearchBridge — Student, Teacher ও Coordinator workflow

এই নথি পূর্ণ feature তালিকার বাস্তবায়ন পরিকল্পনা। সব feature বর্তমান code-এ সম্পূর্ণ তৈরি হয়নি। এই update মূলত laptop, tablet ও Android browser-এর responsive UI যোগ করে। এটি native Android APK নয়।

## 1. এখন কোন অবস্থায় আছে

- কাজ করছে: signup/login, email verification, Gmail SMTP, password reset, role-based dashboard, profile-এর research তথ্য, proposal তৈরি/submit/review, invitation accept করে project-এ যোগ, private files, record/version history, in-app notifications এবং dark mode।
- আংশিক: tasks, milestones, paper, experiment, meeting, contribution, publication, final submission/correction, evaluation, lab ও opportunity—এগুলো database-backed manual record/form। পূর্ণ domain workflow, structured fields ও institutional approval এখনও লাগবে।
- বাকি: স্বতন্ত্র team/recruitment system, matching, team agreement, student-only anonymous supervisor reviews, private shortlist, supervisor change approval, idea/thesis archive, defense/result publication, resource booking, পূর্ণ portfolio, ভাষা/privacy settings, persistent reminders এবং Coordinator/Admin portal।
- বর্তমান Student dashboard-এর কিছু মূল tab পুরোনো preview। কার্যকর shared tools পেতে **Research workspace →** ব্যবহার করতে হয়। পরের ধাপে preview tab-গুলো আসল API ও page-এর সঙ্গে যুক্ত করতে হবে।
- বর্তমান demo-তে project owner evaluation/extension দায়িত্ব দিতে পারে। এটি নতুন institutional policy পূরণ করে না। Admin/Coordinator role ও scoped permission তৈরি করে এই grant API বদলাতে হবে; পুরোনো grants review/revoke করতে হবে। সাধারণ Teacher যেন নিজের বা Student-এর দেওয়া অনুমতিতে institutional reviewer না হন।

## 2. পুরো journey

1. Coordinator department, batch, project type, team size, submission deadlines এবং supervisor capacity ঠিক করবেন।
2. Student/Teacher signup → email verify → login → profile complete। প্রয়োজন হলে institutional approval আলাদা state হবে; email verification তার বিকল্প নয়।
3. Student team তৈরি বা join করবেন → members invitation accept করবেন → agreement ও দায়িত্ব ঠিক করবেন।
4. Team idea/topic বাছবে → private supervisor shortlist → proposal draft → supervision request পাঠাবে।
5. Teacher request দেখে accept / decline / correction দেবেন। Capacity ও allocation policy অনুযায়ী Coordinator assignment নিশ্চিত করবেন।
6. Proposal approved হলে team workspace-এ tasks, milestones, files, papers, experiments ও meetings চলবে।
7. Student work submit করবেন → Teacher feedback/approval দেবেন → correction হলে নতুন version submit হবে।
8. Final submission → supervisor readiness review → Coordinator reviewer ও defense schedule করবেন।
9. Assigned reviewer criteria অনুযায়ী marks দেবেন → Coordinator অনুমোদিত result প্রকাশ করবেন → final corrections review হবে।
10. Completion approved → permission অনুযায়ী thesis archive, portfolio ও showcase → publication/continuation track হবে।

## 3. Student-এর প্রতিটি feature

| Feature | Student-এর কাজ ও ফলাফল | বাস্তবায়নের প্রয়োজন |
|---|---|---|
| Profile | পরিচয়, department/year, skill ও interest পূরণ → save | বর্তমান research profile আছে; structured skills ও academic edit flow লাগবে |
| Team Creation | team name/project type → invite → recipient accept → member | বর্তমানে project-based invite আছে; আলাদা teams এবং join-request flow লাগবে |
| Membership Status | নিজের team, invitation ও searching status দেখা | membership dashboard; প্রতি project cycle-এ membership rules |
| Recruitment Board | skill, vacant seats, description দিয়ে post → applicant review | recruitment posts/applications; accept করার সময় seat transaction |
| Team Availability | open teams filter করে available seats দেখা | confirmed members + reserved seats থেকে availability গণনা |
| Skill Matching | skill/interest অনুযায়ী team বা member suggestion | weighted matching; score-এর কারণ দেখানো; automatic joining নয় |
| Skill Coverage | required skills বনাম accepted members-এর skill দেখা | team requirement ও skill coverage comparison |
| Compatibility | available time, workload, role preference দেওয়া → মিল দেখা | consent-based preferences; private schedule details সীমিত রাখা |
| Team Agreement | roles, meeting rules, contribution rules → সব member accept | agreement version; edit হলে আবার acceptance |
| Supervisor Reviews | eligible student review লিখবে; student-only page-এ anonymous review দেখবে | backend role/eligibility check; author/email/identifying metadata response-এ না দেওয়া |
| Supervisor Shortlist | team-এর পছন্দের supervisor save/rank | team-private API; teacher/co-supervisor read deny |
| Supervision Request | team + proposal + selected teacher → submit | proposal revisions ও request state আলাদা; teacher capacity check |
| Request Tracking | pending/accepted/declined/correction state দেখা | timestamps, decision reason ও request history |
| Supervisor Change Request | কারণ দিয়ে আবেদন → Coordinator সিদ্ধান্ত | আগের supervision চলবে decision পর্যন্ত; reassignment audit + access revoke |
| Idea Board | idea publish → collaborator interest → invite | idea visibility, moderation ও interest records |
| Topic Selection | idea ও permitted archive খুঁজে topic বাছাই | selected topic/version; similar work reference |
| Proposal | problem/objectives/method/timeline/PDF → submit | বর্তমান proposal আছে; structured fields এবং PDF/version association লাগবে |
| Proposal Revision | feedback অনুযায়ী edit → নতুন revision → resubmit | পুরোনো version immutable; reviewer latest ও old দুটো দেখবেন |
| Team Workspace | tasks/files/notes/history ব্যবহার | verified accepted member-only access; draft boundary বজায় রাখা |
| Tasks & Milestones | অনুমতি অনুযায়ী task assign, deadline, completion evidence | Student task permissions লাগবে; final milestone approval Supervisor-এর |
| Progress Tracking | approved milestones-এর ভিত্তিতে progress দেখা | approved weights / total weights; বর্তমান manual percentage-কে official progress বলবেন না |
| Version History | submission/version খুলে আগের কাজ দেখা | uploads replace নয়; version chain ও timestamp |
| Feedback | comments/corrections/approval পড়ে follow-up | comment author/time ও resolve/reopen flow |
| Deadline Extension | proposed date + reason → দায়িত্বপ্রাপ্ত reviewer | pending/approved/rejected; approval ছাড়া মূল deadline বদলাবে না |
| Contribution Log | কাজ, সময়, evidence/link → submit | contribution records আছে; verified contribution আলাদা রাখতে হবে |
| Paper Library | title/DOI/reference/PDF/summary/notes | বর্তমান paper records থেকে structured library; citation export পরে |
| Experiment Journal | dataset/version, parameters, code commit, method, result লেখা | structured experiment runs; reproducibility ও immutable versions |
| Meeting Booking | supervisor-এর available slot দেখে request → confirmation | availability model; pending reservation policy; overlapping booking transaction |
| Meeting Notes | decisions ও action items লেখা/দেখা | note থেকে task/assignee/deadline link |
| Blocker Reports | problem, urgency, attempted fixes → help request | pending/in-progress/resolved; supervisor advice ও follow-up |
| Waiting Status | pending review কতদিন ধরে আছে দেখা | submitted_at থেকে elapsed days; timezone-safe date handling |
| Final Submission | report/slides/code/files checklist পূরণ → submit | final package snapshot; required file validation |
| Defense Information | scheduled date, room/link, instructions দেখা | Coordinator-published schedule; student read-only |
| Result & Final Corrections | published result দেখা → corrections submit | unpublished marks লুকানো; correction version ও completion gate |
| Thesis Archive | অনুমোদিত thesis search/read | completion, consent/license, visibility ও protected download |
| Research Portfolio | approved outputs/skills/contributions দেখানো | verification provenance ও public/private settings |
| Opportunity Board | paid/unpaid post দেখে apply | Teacher post আছে; application/status flow লাগবে |
| Mentorship | junior help request / approved senior mentor feedback | mentor approval, pairing, scope ও interaction records |
| Resource Booking | equipment/GPU/dataset/lab slot request | resource permissions; Coordinator/manager approval; conflict lock |
| Publication Tracking | draft/submitted/revision/accepted/published update | DOI/venue/submission dates ও supporting evidence |
| Showcase & Continuation | approved demo publish / previous work continue request | owner consent, license ও permission; private showcase এখন আছে |
| Settings & Notifications | theme, language, password, privacy, preferences | dark/reset/in-app preferences আছে; language/privacy controls ও delivery preferences লাগবে |

## 4. Teacher / Supervisor-এর প্রতিটি feature

| Feature | Teacher-এর কাজ | সীমা/ফলাফল |
|---|---|---|
| Profile | পরিচয়, interests, expertise update | নিজের profile; institutional role নিজে বদলাতে পারবেন না |
| Supervision Requests | team ও proposal পড়ে accept/decline | capacity ও allocation policy মেনে; unrelated requests read deny |
| Assigned Teams | assigned teams ও members দেখা | assignment revoked হলে access-ও বাতিল |
| Proposal Review | approve/reject/correction + feedback | student evidence অপরিবর্তিত; review history সংরক্ষণ |
| Team Workspace | files/tasks/notes দেখা ও permitted action | assigned team scope; personal shortlist/review নয় |
| Milestone Review | evidence যাচাই → approval/correction; next task | approved milestone থেকে official progress |
| Progress Monitoring | stage, approved progress, delays/blockers দেখা | team comparison permission অনুযায়ী |
| Submission Review | current ও old versions review | submitted files overwrite নয় |
| Feedback | comment/correction/approval দেওয়া | author/time ও target record যুক্ত |
| Deadline Extension Review | delegated হলে decision | institutional permission ছাড়া 403; decision audit |
| Contribution Review | evidence দেখে member contribution মূল্যায়ন | student log edit নয়; assessment আলাদা |
| Paper & Experiment Review | notes/method/result review | experiment data preserve; review field আলাদা |
| Meeting Management | availability publish; request confirm/reschedule | conflicting slots atomically reject; students notified |
| Meeting Notes | decisions লেখা; action items assign | notes-to-task link ও revision history |
| Blocker Support | পরামর্শ, resource referral, follow-up | resolved হলে reason/evidence রাখা |
| Pending Work Dashboard | pending requests/reviews/upcoming meeting | actionable counts; কোনো team না থাকলে empty state |
| Final Submission Review | report/slides/code/files completeness যাচাই | readiness approval; result publication নয় |
| Evaluation | assigned reviewer হিসেবে criteria অনুযায়ী marks | Coordinator assignment আবশ্যক; private draft marks; submit/lock |
| Final Corrections Review | revised work যাচাই | scoped completion authority; final audit |
| Opportunity Board | paid/unpaid openings publish | applications access নিজের post অনুযায়ী |
| Mentorship Support | guidance ও feedback | approved mentorship relationship scope |
| Lab Pages | lab members/projects/recruitment update | lab manager হিসেবে explicit grant; সাধারণ Teacher নয় |
| Campus Problem Bank | campus problem থেকে research idea post | ownership, moderation এবং visibility |
| Publication Tracking | team paper/revision/publication review | evidence-backed status; indexing integration আলাদা |
| Project Showcase | approved demo ও continuation feedback | consent ও publication policy মানতে হবে |
| Settings & Notifications | profile/theme/password ও notification prefs | requests/submissions/meetings scope; student-private data inaccessible |

## 5. Coordinator/Admin workflow

Admin account public signup থেকে তৈরি হবে না। Trusted initial bootstrap অথবা existing privileged admin invitation; প্রতিটি change audited। Coordinator-এর department/batch scope থাকবে।

| Feature | Workflow |
|---|---|
| Academic Setup | department → batch/cycle → project type → team size → deadlines → publish |
| Unmatched Students | searching list → suggest/invite/assist → student consent নিয়ে placement |
| Supervisor Allocation | capacity + interests + preferences → assign/co-assign → notify → audit |
| Defense Scheduling | reviewer eligibility → panel → time/room → conflict check → publish |
| Result Publication | all required evaluations → moderation/approval → lock → publish authorized fields |
| Cross-department Work | participating departments → coordinator scope → supervision/assessment policy |
| Resource Administration | resource + capacity + availability → approve/reject booking → usage record |
| Department Reports | scoped teams/allocation/pending/completion filters → CSV export; formula injection-safe cells |
| Planning Insights | aggregated interests, resource demand, outcomes; no private individual review/shortlist leakage |
| Access Control | role + department + assignment + delegated capability; expiry/revocation |
| Audit History | actor/action/target/time/reason; append-only records; restricted access |

## 6. গুরুত্বপূর্ণ state ও data rules

- Team: draft → recruiting → complete → active → completed/archived। Invite: pending → accepted/declined/expired। Join request: pending → accepted/rejected/withdrawn।
- Supervision request: draft → submitted → accepted/declined/changes_requested/withdrawn। Acceptance ও official allocation policy আলাদা রাখা যাবে।
- Proposal: draft → submitted → changes_requested → resubmitted → approved/rejected।
- Milestone/submission: planned → in_progress → submitted → approved/changes_requested। Rejected/old versions audit-এ থাকবে।
- Meeting/resource: requested → confirmed → completed/cancelled। Pending slot reserve করলে expiry রাখতে হবে।
- Evaluation: assigned → draft → submitted → locked। Result: unpublished → approved_for_publication → published।
- Completion: final_submitted → reviewed → defense_done → corrections_required → corrections_approved → completed।
- Submission author, reviewer এবং institutional approver আলাদা fields; last edit দিয়েই evidence overwrite করা যাবে না।

## 7. Database ও Go backend-এর নির্মাণ পরিকল্পনা

Existing users/projects/team_members/workspace_records/history/files/notifications পুনর্ব্যবহার করুন যেখানে অর্থ মেলে। নিচের model-গুলো incremental migration-এ যোগ হবে:

| Module | নতুন structured tables / সম্পর্ক |
|---|---|
| Academic/roles | departments, academic_cycles, project_types, role_assignments, capability_grants |
| Team formation | teams, memberships, team_invitations, join_requests, recruitment_posts, applications |
| Skills/agreement | skills, user_skills, team_required_skills, availability_preferences, agreements, agreement_acceptances |
| Supervisor | supervisor_capacity, supervision_requests, allocations, change_requests |
| Private student tools | supervisor_reviews, review_eligibility, team_shortlists; teacher APIs থেকে সম্পূর্ণ পৃথক |
| Research | ideas, topics, proposal_versions, milestones, submissions, submission_versions, comments, contribution_reviews |
| Meetings/resources | availability_slots, meetings, meeting_attendees, resources, resource_bookings |
| Evaluation | defense_panels, reviewer_assignments, rubrics, evaluations, published_results, correction_packages |
| Outputs | thesis_archive, portfolio_items, opportunities, opportunity_applications, mentor_pairs, labs, publications, showcases |
| System | notification_outbox, reminder_jobs, delivery_attempts, audit_events |

প্রতিটি module-এর sequence: migration → Go model/validation → permission helper → API → UI → allowed/denied workflow test। আগে backend authorization; UI hide করা অতিরিক্ত সুবিধা।

## 8. Student-only reviews ও shortlist-এর privacy

1. Verified Student হওয়া এবং নির্ধারিত eligibility পূরণ করা প্রয়োজন। Eligible review-এর জন্য বাস্তব supervision relationship/completed cycle যাচাই করুন।
2. Go middleware-এ teacher role deny; Student হলেও অন্য team-এর shortlist deny। Admin-ও default access পাবেন না; abuse moderation-এর জন্য আলাদা audited permission নীতিগতভাবে ঠিক করতে হবে।
3. Response DTO-তে author_id, email, raw eligibility record, identifying attachment বা internal metadata দেবেন না। Anonymous display মানেই database থেকে author মুছে ফেলা নয়।
4. Frontend-এ Teacher-এর menu/component থাকবে না। Direct URL/API, exports, search, notifications, attachment download এবং cache response দিয়েও তথ্য বের হতে পারবে না।
5. Review editing/deletion window, abuse-reporting, minimum cohort size ও small-group re-identification policy নির্ধারণ করুন। ছোট group-এ anonymity নিশ্চিত বলা যাবে না।
6. Tests: teacher GET/POST forbidden; unrelated student shortlist forbidden; eligible student review allowed; anonymous payload-এ identifying fields absent।

## 9. Reminder, notification ও double booking

- Go background worker + SQLite persistent outbox/jobs; server restart হলেও pending jobs থাকবে।
- Event: request submitted, review decision, invitation, new submission, meeting reschedule, published result। প্রত্যেকটিতে authorized recipient ও preferences check হবে।
- Reminder: configurable 24h/1h-before meeting, approaching deadline এবং overdue review; event/user/channel key দিয়ে duplicate delivery ঠেকানো।
- SMTP retry bounded exponential backoff; sent/failed status; expired/cancelled event skip। বর্তমান email transport reuse করা যাবে, কিন্তু reminder worker এখনও তৈরি করতে হবে।
- Booking conflict: একই resource/person-এর start < existing_end AND end > existing_start। একই transaction-এ availability check ও insert/confirm; একসঙ্গে দুই request হলেও একটি সফল হবে।
- বর্তমান meeting tool supervisor-এর একঘণ্টার overlap পরীক্ষা করে; পূর্ণ participant/resource conflict ও concurrency guarantees-এর বিকল্প নয়।
- Database-এ UTC; UI-তে ব্যবহারকারীর local timezone।

## 10. দুই teammate-এর কাজ ভাগ ও order

| Phase | Teammate A | Teammate B | Done হওয়ার শর্ত |
|---|---|---|---|
| 1: ভিত্তি | auth/role/scoped permission migrations | responsive shell, auth/profile UI | Student/Teacher/Admin boundary tests |
| 2: team | teams/invite/join/recruitment APIs | team formation, skills, recruitment UI | accept/reject/seat-race complete flow |
| 3: supervision | request/allocation/proposal versions | supervisor discovery/shortlist/request UI | request → correction → approval; privacy tests |
| 4: workspace | files/tasks/milestones/audit APIs | research workspace ও progress UI | submit → review → revision → approved progress |
| 5: collaboration | meetings/resources/outbox | booking/calendar/notifications UI | concurrent conflict test; restart-safe reminders |
| 6: completion | reviewer/evaluation/results/archive | defense/final/correction/portfolio UI | only assigned reviewer; only authorized publish |
| 7: community | mentorship/labs/publications/showcase APIs | community/search/preferences UI | ownership, visibility ও integration tests |

দুজন একই main.go-তে একসঙ্গে বড় edit করবেন না। Feature অনুযায়ী Go handler files ও JS modules আলাদা করুন। Branch examples: feature/team-api, feature/team-ui। API request/response আগে লিখে একমত হন; ছোট pull request-এ merge করুন।

## 11. Responsive update বসানো ও পরীক্ষা

এই ZIP-এর পুরো **web** folder বর্তমান project-এর **web** folder-এ copy/replace করুন। এতে 4টি HTML page-এ responsive stylesheet/script যুক্ত আছে; শুধু CSS copy করলে mobile menu আসবে না। পুরোনো .env/database একই থাকবে। Server বন্ধ করে copy → go run . → Ctrl+F5।

- Laptop >1024px: sidebar স্থায়ী; available space অনুযায়ী card grid।
- Tablet 761–1024px: Menu দিয়ে navigation expand/collapse; full-width content।
- Android ≤760px: touch-friendly forms/buttons; cards one/two column; ≤420px-এ single column। Native select/file picker browser/OS অনুযায়ী দেখাবে।
- Android-এ hover নেই; banner-এর tap button দিয়ে details দেখা যায়। System reduced-motion setting মানা হয়।
- 360, 390, 768, 1024, 1366px-এ login/signup, long name/title, menu, modal, file upload, dark mode, landscape এবং keyboard open অবস্থায় পরীক্ষা করুন।
- Automated DOM checks layout pixel বা বাস্তব Android hardware পরীক্ষা নয়। এই environment-এ real-device visual QA হয়নি।

Phone-এ local project দেখতে হলে phone-এর localhost নয়, laptop-এর LAN address ব্যবহার হবে। একই Wi-Fi-তে laptop server listen address, APP_URL এবং Windows private-network firewall configuration মেলাতে হবে; email reset/verification link-ও APP_URL ব্যবহার করে। Network-এর বাইরে ব্যবহারের জন্য HTTPS hosting প্রয়োজন।
