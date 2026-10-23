# ResearchBridge: Teacher / Supervisor-এর কাজ

আপনার ২২ অক্টোবরের source দেখে তৈরি। “Basic” মানে database-backed form/record আছে; পুরো academic workflow যাচাই বা সম্পূর্ণ হওয়ার দাবি নয়।

| Student-এর কাজ/ঘটনা | Teacher কী করবেন | বর্তমান অবস্থান |
|---|---|---|
| নিজের জন্য supervisor খোঁজা | পরিচয়, bio, qualification, research interests, expertise, photo রাখা | Profile আছে; directory-এর সঙ্গে সব field auto-sync পূর্ণ যাচাই বাকি |
| Supervision request submit | request ও proposal দেখে accept/correction/reject | Code আছে; accept ও proposal approval বর্তমানে একই decision |
| Team নিয়ে project করা | নিজের assigned project/team ও members দেখা | আছে |
| Detailed proposal submit | problem/objectives/method/outcomes/scope পর্যালোচনা | Student-এর structured fields আছে; Teacher review dialog প্রধানত abstract দেখায়—পূর্ণ proposal layout বাকি |
| Proposal correction জমা | সংশোধন দেখে পুনরায় decision | Code আছে; review history পূর্ণ workflow বাড়াতে হবে |
| Milestone work | milestone তৈরি, progress ও approval | Basic record |
| Task progress | দায়িত্ব, due date ও next action দেওয়া | Basic record |
| Report/code জমা | submission review ও feedback | Basic record + file access |
| সংশোধন প্রয়োজন | comment/correction/approval দেওয়া | Feedback আছে; threaded discussion/resolve flow বাকি |
| সময় বাড়ানোর আবেদন | কারণ দেখে extension decision | Basic record; permission দরকার, approved date প্রয়োগের পূর্ণ workflow বাকি |
| Individual contribution | কে কী কাজ করেছে তা দেখা | Basic contribution record; automatic contribution scoring নেই |
| Papers/reading notes | related work ও paper summary review | Basic record; citation/DOI library automation নেই |
| Experiment results | method/results দেখে advice | Basic record; structured experiment comparison নেই |
| Meeting request | meeting schedule/decision | Basic record ও এক ঘণ্টার conflict check; availability slots/সম্পূর্ণ calendar নেই |
| Meeting notes | সিদ্ধান্ত ও follow-up নির্দেশনা | Basic note record |
| Blocker/problem report | সাহায্য ও feedback দেওয়া | Basic blocker record |
| Files upload | নিজের project-এর files ও আগের versions দেখা | Code আছে |
| Project progress | milestones/tasks ধরে progress দেখা | Basic monitoring; weighted approved milestone calculation পূর্ণ নয় |
| Final submission | report/slides/code review | Basic record; required-file checklist/locked final package বাকি |
| Reviewer assignment | criteria ও marks দেওয়া | Basic score/criteria; independent reviewer allocation বাকি |
| Final correction | সংশোধন যাচাই করে status দেওয়া | Basic record; official completion gate বাকি |
| Research opportunity | paid/unpaid opening post | Basic post; application selection workflow বাকি |
| Mentorship | guidance ও feedback | Basic record; formal mentor pairing নেই |
| Lab/group activity | members/projects/recruitment তথ্য | Basic lab record; institution-controlled lab ownership বাকি |
| Campus problem | research idea/problem দেওয়া | Basic post |
| Publication | paper submission/revision/publication progress | Basic record; structured venue/DOI/status history বাড়ানো দরকার |
| Showcase | project demo link ও নির্দেশনা | Basic workspace record; public publishing/consent workflow বাকি |
| নতুন request/work/meeting | notification দেখা ও relevant কাজ খোলা | Existing stored notifications + এই update-এ polling, popup, sound |
| নিজের settings | dark mode, password reset, notification preferences | আছে; full privacy/language controls বাকি |

## Teacher-এর জন্য পরবর্তী গুরুত্বপূর্ণ কাজ

১. Supervision capacity: কত team নেবেন, কত slot খালি, accepting students on/off।
২. Supervision acceptance ও proposal approval আলাদা status/decision।
৩. সম্পূর্ণ detailed proposal viewer: সব structured fields ও attached files।
৪. Teacher profile থেকে student directory sync।
৫. Meeting availability ও booking calendar।
৬. Rubric অনুযায়ী marks এবং দায়িত্বপ্রাপ্ত reviewer assignment।
৭. Scheduled deadline/meeting reminders—বর্তমান notification নতুন event-এ আসে, calendar reminder নয়।
৮. Help & Support ticket/reply; আলাদা admin dashboard ছাড়াও support operator Supabase দিয়ে চালাতে পারে।

## কার কাছে কোন ক্ষমতা

- সাধারণ Teacher শুধু assigned project-এর অনুমোদিত তথ্য/কাজ করবেন।
- Teacher student-এর private shortlist ও student-only supervisor reviews দেখতে পারবেন না।
- Coordinator-এর দায়িত্ব: supervisor allocation/change, defense schedule, reviewer assignment এবং result publication।
- Support admin: account/access সমস্যা ও support ticket; স্বয়ংক্রিয়ভাবে academic marks/approval দেওয়ার অধিকার নয়।
- বর্তমান code-এ evaluation/extension permission project owner দেয়। Institutional workflow-এর জন্য এটি coordinator-controlled করা বাকি।
