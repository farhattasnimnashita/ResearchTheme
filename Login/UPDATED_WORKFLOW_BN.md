# ResearchBridge — Updated Standard Workflow (বাংলা)

এই নথি ZIP-এর বর্তমান Go backend, Student dashboard, Teacher dashboard এবং supervision workflow দেখে তৈরি। এটি future feature-এর প্রতিশ্রুতি নয়; কোন ধাপগুলো বর্তমান code-এ আছে এবং কোথায় উন্নতি দরকার তা আলাদা করে দেখানো হয়েছে।

## A. Student-এর জন্য প্রস্তাবিত কাজের ক্রম

1. **Account তৈরি ও email verify**
   - Student account খুলে email verification সম্পন্ন করবে।
   - Verification-এর পর login করবে।

2. **My Profile সম্পূর্ণ করা**
   - Name, institution, department, academic year, semester ঠিক করবে।
   - Profile photo / bio / research interests / general expertise চাইলে যোগ করবে।
   - এরপর **Academic History & Skills (Private)** অংশে:
     - অন্তত একটি academic year-এর CGPA যোগ করবে (0.00–4.00)।
     - অন্তত একটি skill যোগ করবে (যেমন C++, Python, Java, Research Writing)।
     - Research interests যোগ করতে পারবে।
   - Academic Profile-এর Save button চাপবে। শুধু general Profile-এর Expertise field পূরণ করলেই supervision request-এর skill requirement পূরণ হবে না; request validation `student_skills` table-এর Academic Profile skills দেখে।

3. **Research topic / project তৈরি**
   - Project title, research area ও short abstract/description লিখবে।
   - Project draft হিসেবে save করবে।
   - Topic খুব বড় না করে একটি নির্দিষ্ট research problem বেছে নেবে।

4. **Team তৈরি ও সদস্য যোগ**
   - Project owner teammate-এর verified account-এ invitation পাঠাবে।
   - Invited student নিজের dashboard-এর Invitations থেকে Accept করবে।
   - Invitation পাঠানো মানেই member হওয়া নয়; Accept করার পরই `team_members`-এ membership তৈরি হয়।
   - Teacher-এর কাছে এবং workspace-এ শুধু project owner ও confirmed/accepted members দেখানো হবে।

5. **প্রত্যেক team member-এর Academic Profile সম্পূর্ণ করা**
   - Owner এবং প্রত্যেক confirmed member-এর অন্তত একটি year-wise CGPA record এবং অন্তত একটি skill থাকতে হবে।
   - একজন member-এরও তথ্য অসম্পূর্ণ থাকলে supervision request backend থেকে reject হবে।
   - Error message-এ অসম্পূর্ণ member-এর নাম দেখানো হবে। প্রত্যেক student নিজের account-এ login করে নিজের Academic Profile update করবে।

6. **Supervisor নির্বাচন ও request পাঠানো**
   - Supervisor page-এ project ও verified teacher নির্বাচন করবে।
   - Send Supervision Request চাপলে frontend আগে requester-এর Academic Profile পরীক্ষা করবে।
   - requester-এর academic record বা skill না থাকলে request পাঠানো হবে না; dashboard স্বয়ংক্রিয়ভাবে My Profile-এর Academic History & Skills অংশে যাবে।
   - Backend আবার project owner ও সব confirmed member-এর তথ্য যাচাই করবে। Frontend bypass করলেও incomplete profile-এ request পাঠানো যাবে না।
   - Academic records ও skills public student directory-তে প্রকাশ করা হবে না।

7. **Teacher-এর সিদ্ধান্তের জন্য অপেক্ষা**
   - Request Pending থাকবে।
   - নির্বাচিত teacher-ই team academic information দেখতে পারবেন।
   - Teacher Accept বা Decline করবেন; Decline-এর জন্য reason দিতে হবে।
   - অন্য teacher-এর team academic data access করা উচিত নয়; API exact assigned teacher ID যাচাই করে।

8. **Full proposal লেখা**
   - Supervision Accepted হলে Proposal page unlock হবে।
   - Problem Statement, Objectives, Methodology, Expected Outcomes, Scope & Limitations লিখবে।
   - আগে Save Draft, তারপর সব section প্রস্তুত হলে Submit for Review।

9. **Proposal review ও revision**
   - Teacher proposal approve, changes requested, অথবা reject করতে পারবেন।
   - Changes requested হলে feedback পড়ে proposal update করে আবার submit করবে।
   - পুরোনো submission/version মুছে না দিয়ে নতুন version রাখতে হবে।

10. **Research execution**
    - Approved proposal-এর পর team ও supervisor মিলে milestones, tasks, meetings ও deadlines ঠিক করবে।
    - কাজের প্রমাণ, files, experiment notes, paper notes ও contribution log নিয়মিত update করবে।
    - Blocker হলে দ্রুত report করবে।

11. **Submission, feedback ও correction**
    - Student submission upload করবে।
    - Teacher review করে feedback/status দেবেন।
    - Correction চাইলে Student নতুন version submit করবে।
    - Teacher Revision/Pending Work card-এ project title, complete team list এবং কে submission করেছে তা যাচাই করবেন।

12. **Final submission ও completion**
    - Final report, presentation, code/data ও references প্রস্তুত করবে।
    - Supervisor-এর final readiness review নেবে।
    - Department/Coordinator-এর defense, marks publication ও archive ধাপ institution policy অনুযায়ী হবে।

## B. Teacher-এর কাজের ক্রম

1. Profile-এ research interests, expertise এবং supervision capacity ঠিক করা।
2. Supervision Requests-এ pending request খোলা।
3. Project title, প্রত্যেক team member-এর নাম, year-wise CGPA, skills এবং research interests দেখা।
4. Capacity ও research fit বিবেচনা করে Accept/Decline করা; decline করলে reason দেওয়া।
5. Accept করার পর full proposal review করা।
6. Proposal approve / changes requested / reject করা।
7. Accepted team-এর workspace-এ milestones, tasks, meetings ও submissions পরিচালনা করা।
8. প্রতিটি submission/revision-এ project title, full team, submitter এবং previous version যাচাই করা।
9. Final submission ও corrections review করা।

## C. বর্তমান code-এ এই update-এ যা সংশোধন করা হয়েছে

- Teacher Assigned Teams: `/api/workspace/{id}/members` এবং teacher-authorized `/api/projects/{id}/team-academics` থেকে member list merge করে duplicate বাদ দিয়ে দেখানো।
- Teacher Pending Work / Revision: একই merged team list, project title, record title, submitter এবং revision parent তথ্য দেখানো।
- Backend member query: project owner এবং `team_members`-এর accepted members-কে explicit `UNION` দিয়ে একবার করে ফেরত দেওয়া।
- Student Profile editor: ভুল করে supervisor-request logic Profile Edit form-এ বসে থাকার bug সরিয়ে basic profile save handler ঠিক করা।
- Student Send Supervision Request: আসল request handler-এ Academic Profile check বসানো; incomplete requester হলে request না পাঠিয়ে Profile-এ redirect করা।
- Backend supervision request: project owner এবং প্রতিটি confirmed teammate-এর academic record ও skill বাধ্যতামূলক করা; incomplete member-দের নামসহ request reject করা।
- Teacher academic privacy: academic endpoint কেবল নির্বাচিত supervisor-এর জন্য সীমাবদ্ধ রাখার existing authorization বজায় রাখা।
- Browser cache: modified JS-এর asset version update করা হয়েছে।

## D. এখন সবচেয়ে জরুরি উন্নতির জায়গা

### Priority 1 — Team membership ও data correctness
- দুইজন member দেখাতে হলে দ্বিতীয় student-কে invitation Accept করতে হবে। Pending invitation `team_members`-এ থাকে না।
- দুই account দিয়ে test করো: owner invite করবে → second student accept করবে → owner ও teacher—দুই dashboard-এ দুইজনের নাম দেখাবে।
- Team member remove/leave, owner transfer এবং team size limit-এর explicit workflow এখনো শক্ত করা দরকার।
- `team_members`-এ unique constraint `(project_id, user_id)` এবং team-size validation transaction-এর মধ্যে রাখা উচিত।

### Priority 2 — Academic Profile
- Academic skills-এর একাধিক জায়গা আছে: general profile-এর `expertise` এবং private Academic Profile-এর `student_skills`। Supervision request-এর জন্য private Academic Profile-এর skills-ই authoritative।
- UI-তে এই পার্থক্য স্পষ্ট রাখো, যাতে student ভুল field পূরণ করে request blocked না হয়।
- CGPA কোন grading scale (4.00/5.00) এবং কোন year/semester সম্পন্ন হয়েছে—department policy অনুযায়ী configurable করা দরকার।
- Profile save ও request validation-এর integration tests লিখতে হবে।

### Priority 3 — Authorization ও privacy
- অন্য teacher-এর `GET /api/projects/{id}/team-academics` request অবশ্যই 403 পাবে।
- Student যেন অন্য student-এর private CGPA/skills API দিয়ে পড়তে না পারে।
- Teacher শুধু assigned project-এর workspace/files/records দেখতে পারবে।
- সব গুরুত্বপূর্ণ সিদ্ধান্তে actor, timestamp, previous status, new status ও feedback audit log রাখতে হবে।

### Priority 4 — Workflow consistency
- `Draft`, `Supervision Pending`, `Supervision Accepted`, `Proposal Submitted`, `Changes Requested`, `Approved`, `Rejected`—state transition এক জায়গায় document ও test করতে হবে।
- Supervision acceptance এবং full proposal approval আলাদা ধাপ হিসেবেই থাকবে।
- Duplicate request, repeated Accept, stale form submission এবং simultaneous capacity booking-এর race condition test করতে হবে।

### Priority 5 — Product completeness
- Existing feature plan অনুযায়ী Student dashboard-এর কিছু tab এখনো preview/partial। Research workspace-এর database-backed feature এবং preview feature আলাদা করে label করো।
- Recruitment board, team matching, availability, skill coverage, team agreement, supervisor shortlist, student-only anonymous reviews, Coordinator/Admin allocation, defense scheduling ও thesis archive এখনো পূর্ণ institutional workflow নয়।
- আগে core path (team → supervisor → proposal → work → revision → final review) স্থিতিশীল করো; তারপর optional advanced features যোগ করো।

## E. Local verification checklist

1. `node --check web/assets/teacher-dashboard.js`
2. `node --check web/assets/nashita-dashboard.js`
3. `node --check web/assets/academic-profile.js`
4. `gofmt -d proposal.go workspace.go academic_profile.go`
5. `go test ./...`
6. দুই student account দিয়ে invitation accept এবং team membership test।
7. Academic record missing / skill missing / all complete—এই তিন অবস্থায় request test।
8. Team-এর দ্বিতীয় member-এর profile incomplete রেখে request করলে backend block করে কি না test।
9. Assigned teacher academic profile দেখতে পারে কি না এবং অন্য teacher 403 পায় কি না test।
10. Teacher Assigned Teams ও Revision card-এ দুই member, project title এবং submitter ঠিকমতো দেখায় কি না test।

**পরীক্ষার সীমাবদ্ধতা:** এই environment-এ JavaScript syntax check ও Go formatting check চালানো গেছে। Go dependencies আগে থেকে cache-এ না থাকায় এবং network access না থাকায় `go test ./...` সম্পন্ন করা যায়নি। নিজের computer-এ dependencies available রেখে অবশ্যই test চালাবে।
