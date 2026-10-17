# Sky Blue Teacher Dashboard — update 2026-10-16

## বর্তমান project-এ update বসানো

আপনার সর্বশেষ কাজ করা project: `D:\project-250\project test`। Server-এর terminal-এ আগে **Ctrl+C** দিন। একই port-এ অন্য server চালু রাখবেন না।

1. নতুন ZIP extract করুন। ভিতরের `ResearchBridge` folder VS Code-এ খুলুন।
2. সেই extracted folder-এর PowerShell terminal-এ চালান:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\APPLY_UPDATE.ps1 -Target "D:\project-250\project test"
```

Script নতুন Go files ও web assets copy করবে। বর্তমান `.env` ও database overwrite করবে না। আগের code/settings/database-এর local backup `backup-update-...` folder-এ রাখবে। Backup কখনও GitHub-এ দেবেন না; script `.gitignore`-এ backup/database/secret exclusions যোগ করে। Target-এ `.env` না থাকলে script থামবে—console settings বানাবে না।

3. এরপর:

```powershell
cd "D:\project-250\project test"
go run .
```

Terminal-এ `Email mode: smtp` থাকবে যদি আগের Gmail `.env` ব্যবহার করেন। Browser-এ http://localhost:8080 খুলে Ctrl+F5 দিন। Teacher account দিয়ে login করুন।

Manual copy করতে চাইলে extracted `ResearchBridge` থেকে **সব `.go` files**, `go.mod`, `go.sum`, `START.ps1` ও **পুরো web folder** বর্তমান project-এ copy/replace করুন। `.env` বা database copy করবেন না। শুধু main.go replace করলে এই update সম্পূর্ণ হবে না।

## Dashboard ও dark mode

Student dashboard-এর মতো একই layout: বড় sky-blue banner, hover/tap-এ supervision details, চারটি summary card এবং Current Projects / Next Meeting / Recent Activity panels। Sidebar-এ Dashboard, My Profile, Assigned Teams, Supervision Requests, Proposal Review, Milestones, Files & Feedback, Meetings আছে।

**More supervision tools**-এ অন্য সব review, feedback, progress, evaluation, publication, opportunity, lab ও campus problem options আছে। নিচের ডানদিকে Dark mode button। কোনো team না থাকলেও option দেখা যাবে; team-ভিত্তিক কাজের জন্য student-এর submitted proposal প্রয়োজন। Opportunity, Lab ও Campus Problem সরাসরি তৈরি করা যায়।

Sidebar থেকে feature খুলে project selector দিয়ে team বেছে নিন। Team বদলালেও নির্বাচিত feature খোলা থাকবে।

| কাজ | ব্যবহার |
| --- | --- |
| Supervision / Proposal review | Student-এর submitted proposal approve, changes requested বা reject; সিদ্ধান্তের সঙ্গে feedback |
| Assigned Teams | Project owner ও invitation accept করা members দেখা |
| Team Workspace | Tasks তৈরি, member assign, notes, status ও progress update |
| Milestones / Progress | Milestone review, শতাংশ progress, completed tasks ও pending blockers |
| Submissions / Files | Reports, slides, code ও files upload/download; upload-এর আগের versions অক্ষত থাকে |
| Feedback | Comment, correction ও approval record |
| Deadline extension | Student request review; project owner দায়িত্ব দিলে supervisor সিদ্ধান্ত দেবেন |
| Contribution | Member-এর work/evidence record এবং supervisor-এর review notes |
| Papers / Experiments | Reading notes, methods, results, links ও review feedback |
| Meetings | Meeting request, scheduling, link, cancellation; একই teacher-এর scheduled meetings-এর মধ্যে অন্তত এক ঘণ্টা |
| Meeting Notes | সিদ্ধান্ত, follow-up note; আলাদা Task-এ member assign |
| Blocker Support | Student সমস্যা জমা দেয়, supervisor advice/status দেয় |
| Final submission / Corrections | Report, slides, code submission ও correction-এর review |
| Evaluation | Project owner reviewer দায়িত্ব দিলে criteria লিখে ১০০-এর মধ্যে score |
| Mentorship | Guidance, status এবং follow-up feedback |
| Publication | Writing/submission/revision-এর record, target date ও progress |
| Project Showcase | Project-এর members এবং supervisor-এর জন্য demo link ও notes |

Project selector-এর **Research resources** বেছে Opportunity Board, Lab Pages ও Campus Problem Bank খুলবে। Teacher নিজের record লিখে **published** করলে verified users দেখতে পাবে। Paid/unpaid, eligibility, apply instructions, lab members/projects/recruitment ও problem details সংশ্লিষ্ট description-এ লিখবেন। অন্য Teacher-এর unpublished record দেখা বা edit করা যায় না।

My profile থেকে পরিচিতি, interests, expertise এবং in-app notification preferences save করা যায়। Review queue-তে pending proposals, অন্য member-এর pending work, upcoming meetings ও notifications আছে।

## Student কীভাবে কাজ পাঠাবে

Student-এর আগের মেরুন dashboard আছে। **Research workspace →** link থেকে নতুন shared workspace খুলবে।

- Project owner নিজের project-এ verified Student-এর email দিয়ে invitation পাঠাবে। অন্য Student Research workspace → Requests & invitations থেকে accept করলে project access পাবে।
- Files & Versions থেকে file upload করুন। Upload-এর পর **Submit for review** চাপলে protected file link-সহ submission form খুলবে। Title/body পূরণ করে save করুন।
- Paper, experiment, blocker, meeting request, contribution, deadline request, final submission ও correction-এর record জমা দেওয়া যায়। Student নিজের কাজ approved করতে পারে না।
- **Submit new version** আগের record রেখে নতুন submission তৈরি করে। Version history থেকে আগের saved content ও review দেখা যায়।
- **Review Responsibilities**-এ শুধু project owner evaluation ও deadline-extension review-এর দায়িত্ব দিতে পারে। Supervisor বদলালে আগের অনুমতি নতুন supervisor-এর কাছে যায় না; owner আবার দায়িত্ব দেবে। এই permission বিশ্ববিদ্যালয়/admin appointment নয়।

## Forgot password

Login-এর নিচে Forgot password? → signup email → Gmail-এর **Reset your ResearchBridge password** email → link খুলে নতুন password দুইবার দিন।

একই Gmail SMTP অথবা Resend settings ব্যবহার হবে; আবার service setup করার প্রয়োজন নেই। Link ৩০ মিনিট কার্যকর, একবার ব্যবহারযোগ্য। Reset হলে আগের সব sessions বন্ধ হয়; নতুন password দিয়ে login করতে হবে। Password reset email account verification সম্পন্ন করে না—unverified account-কে verification-ও করতে হবে। প্রতি email-এ এক মিনিটে একটি request; Inbox ও Spam দেখবেন। Unknown email-এও একই generic confirmation থাকে।

## এই implementation-এর পরিধি

সব module database-backed forms, status, notes, links, review এবং history হিসেবে কাজ করে। এটি Zoom/Google Calendar booking, automatic plagiarism checking, institution-admin appointment, external publication indexing বা lab membership identity verification করে না। Scheduled meeting-এর link নিজে দিন। Meetings-এর সময় browser-এর local timezone দিয়ে লিখবেন। Progress teacher-maintained; automatic contribution scoring নেই। Notifications in-app; account verification ও reset-এর জন্য email যায়।

File সর্বোচ্চ 8 MB; প্রতি project-এ মোট 64 MB। বড় dataset/code repository-এর জন্য URL ব্যবহার করুন। Files SQLite database-এ private attachment হিসেবে থাকে। একই নামের নতুন upload আলাদা version; file delete/replace করে পুরনো evidence মুছে ফেলার feature নেই।

নতুন schema existing project-এর users/sessions/proposals অক্ষত রেখে table যোগ করে। Update-এর আগে server বন্ধ রেখে backup দরকার। পুরনো app-এর আলাদা `database.go` বা অন্য project-এর Go source নতুন folder-এ মেশাবেন না।
