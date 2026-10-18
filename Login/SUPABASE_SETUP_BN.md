# Supabase database update — ধাপ ১ ও ২

ভিত্তি: feature/online-deployment, commit 2a4dfe148429b9b4472a6d5c4039ada37e50f1d7।
এটি database update; Render deployment বা email API update নয়।

## কোথায় রাখবে

ZIP extract করে তার Login folder-এর সব ফাইল তোমার
D:\project-250\ResearchTheme-GitHub\Login folder-এ copy করো।
একই নামের এই update-এর ফাইলগুলো replace করবে; পুরো পুরোনো Login folder delete করবে না।
আগে Go server চললে Ctrl+C দিয়ে বন্ধ করো।
তোমার নিজের code-এ এই সময় অন্য পরিবর্তন থাকলে আগে backup করে diff মিলিয়ে নেবে।
এই ZIP-এ .env, .db, credentials, frontend design বা cmd/check-db নেই।
তোমার বিদ্যমান .env, SQLite database, dashboard files এবং cmd/check-db রেখে দেবে।

## .env settings

নিজের Login/.env-এ existing SMTP settings রেখে শুধু সঠিক DATABASE_URL আছে নিশ্চিত করো:

```env
DATABASE_URL=postgresql://postgres.PROJECT_REF:ENCODED_PASSWORD@SESSION_POOLER_HOST:5432/postgres?sslmode=require
EMAIL_MODE=smtp
```

উপরের DATABASE_URL শুধুই কাঠামো। তোমার সফলভাবে পরীক্ষিত আসল URI ব্যবহার করবে।
Password URL-encoded হবে। SMTP_USER, SMTP_PASS, EMAIL_FROM আগের মতো থাকবে।
APP_URL আপাতত http://localhost:8080 এবং APP_ENV=development রাখো।
Terminal-এর environment variables .env-এর চেয়ে অগ্রাধিকার পায়।

## চালানো

```powershell
cd D:\project-250\ResearchTheme-GitHub\Login
go mod tidy
go test ./...
go run .
```

একই port-এ পুরোনো server চালু থাকলে সেটির terminal-এ Ctrl+C দিয়ে বন্ধ করো।
সফল startup-এ দেখবে:

```text
Database mode: postgres (Supabase)
ResearchBridge: http://localhost:8080
Email mode: smtp
```

Startup নিজেই একটি transaction-এর মধ্যে প্রয়োজনীয় tables তৈরি করে।
Supabase SQL Editor-এ আলাদা করে postgres_schema.sql paste করার দরকার নেই।
আবার চালালে existing data রেখে schema setup হয়।
DATABASE_URL দেওয়া কিন্তু ভুল হলে app বন্ধ হবে; গোপনে SQLite-এ ফিরে যাবে না।
DATABASE_URL না থাকলে আগের DB_PATH-এর SQLite চলবে।

## Tables কোথায়

Supabase Table Editor-এর schema selector-এ researchbridge বেছে নাও।
public schema-তে tables দেখাবে না। Custom auth-এর users/password hashes/sessions
public Data API থেকে আলাদা রাখার জন্য private researchbridge schema ব্যবহার করা হয়েছে।
এই schema Data API exposed schemas-এ যোগ করবে না।
Supabase Authentication > Users-এ এই custom Go users দেখাবে না।

Tables: users, verifications, sessions, projects, password_resets,
reset_requests, researcher_profiles, team_members, project_permissions,
workspace_records, workspace_history, workspace_files, workspace_notifications।

## প্রথম manual check

1. নতুন Student account register করো; email link verify করে login করো।
2. নতুন Teacher account register করে একইভাবে verify/login করো। Admin approval নেই।
3. Profile edit, student search, team invitation এবং accept/reject পরীক্ষা করো।
4. Proposal submit/review এবং workspace notes/files পরীক্ষা করো।
5. Forgot password দিয়ে আসল email delivery পরীক্ষা করো।
6. Server বন্ধ করে আবার চালিয়ে দেখো account/project আছে কি না।

## পুরোনো data এবং limits

পুরোনো SQLite users/data এই update নিজে থেকে Supabase-এ copy করে না।
আগের .db নিরাপদে রাখবে। Supabase-এ শুরুতে নতুন signup করতে হবে,
অথবা পুরোনো data রাখার জন্য আলাদা tested data-transfer step দরকার।

Attachments এখন PostgreSQL BYTEA-তে থাকে; Supabase Storage যোগ করা হয়নি।
ফলে attachment-ও database quota ব্যবহার করে। বর্তমান প্রতি-file 8 MB এবং
প্রতি-project 64 MB limit বহাল আছে। বড় files-এর জন্য repository/document link ব্যবহার করো।

Multi-step transactions PostgreSQL advisory lock দিয়ে ধারাবাহিকভাবে চালানো হয়।
এতে আগের SQLite single-writer আচরণ বজায় থাকে—team capacity, one-time token ও
meeting conflict checks-এর জন্য। বড় traffic হলে scoped row locks-এ উন্নয়ন দরকার।

Bundled pgx v5.7.6 দিয়ে পরীক্ষা করা হয়েছে; Go 1.23 বা নতুন version প্রয়োজন।
তোমার আগে install করা pgx v5.11.0-এর বদলে ZIP-এর go.mod/go.sum এই tested version রাখে।

## Validation

SQLite test suite পাস। একই HTTP/API suite PostgreSQL WASM (PGlite) runtime-এ পাস:
signup/verification, both-role login, reset tokens/session revocation, profile flags,
student search, proposals, team invitations, protected files, review permissions,
meeting conflict, notifications, workspace history। Migration restart persistence tested।
Dashboard access test-এর পুরোনো script filename assertion teammate-এর বর্তমান
nashita-dashboard.js-এর সঙ্গে মিলানো হয়েছে; frontend বদলানো হয়নি।

তোমার live Supabase credentials আমাদের কাছে নেই। Supabase-specific connection,
actual Gmail delivery এবং দুই laptop থেকে shared data manual check বাকি।
Render-এর email/API এবং port settings এই update-এর অংশ নয়।

Developer PostgreSQL integration tests: disposable test database ব্যবহার করে
POSTGRES_TEST_URL সেট করে go test ./... চালানো যায়। Tests unique rb_test_* schema
তৈরি/সরিয়ে নেয়। Normal go test SQLite ব্যবহার করে।

.env ও .db GitHub-এ commit করবে না।
