# Supabase Storage setup — ResearchTheme

এই আপডেটে workspace attachment Storage-এ রাখা যায়। পুরোনো database attachment একই URL দিয়ে download হবে; স্বয়ংক্রিয় migration বা deletion করা হয় না। এই সংস্করণে profile picture upload backend নেই; এই পরিবর্তন নতুন profile-photo feature যোগ করে না।

## ১. কোড বসানো
ZIP extract করে Login folder-এর ফাইলগুলো বর্তমান Login folder-এ copy/replace করুন। বর্তমান .env মুছবেন না। ZIP-এ credentials, database, backup বা executable রাখা হয়নি। পরিবর্তিত ফাইল: main.go, workspace.go, workspace_files.go, postgres_schema.sql। নতুন ফাইল: supabase_storage.go, supabase_storage_test.go, এই নির্দেশিকা। অন্য source/assets অপরিবর্তিত।

## ২. Supabase
Storage → research-files bucket → Public OFF নিশ্চিত করুন। Bucket তৈরি আছে। নতুন permissive RLS policy প্রয়োজন নেই: Go backend user session ও project membership যাচাই করে server-only credential দিয়ে ফাইল upload/download করে। Bucket public করবেন না।
Project Settings → API Keys → Secret keys থেকে একটি secret key তৈরি/কপি করুন (sb_secret_ দিয়ে শুরু)। Legacy service_role key-ও সমর্থিত, কিন্তু publishable/anon key storage credential হিসেবে ব্যবহার করবেন না। এই key browser JS, GitHub, screenshot বা chat-এ দেবেন না।

## ৩. Render Environment
নতুন কোড push করার পরে Render → Environment-এ যোগ করুন:

| Key | Value |
|---|---|
| STORAGE_MODE | supabase |
| SUPABASE_STORAGE_BUCKET | research-files |
| SUPABASE_STORAGE_KEY | নিজের server-only Supabase secret key |

আগের SUPABASE_URL, SUPABASE_PUBLISHABLE_KEY, DATABASE_URL, EMAIL_MODE=supabase, APP_URL রাখুন। SUPABASE_PUBLISHABLE_KEY-তে secret key দেবেন না। Save, rebuild, and deploy করুন। Startup-এ workspace_storage table স্বয়ংক্রিয়ভাবে তৈরি হবে; SQL Editor-এ manual migration দরকার নেই।

Local test করতে নিজের .env-এ একই তিনটি নতুন variable যোগ করুন। STORAGE_MODE না দিলে database mode বজায় থাকে। পুরোনো Storage file পড়তে Storage key ও URL রাখতে হবে, এমনকি নতুন upload-এর mode database করলেও।

## ৪. Commit / push
PowerShell, repository root:

```powershell
cd D:\project-250\ResearchTheme-GitHub
cd Login
go test ./...
cd ..
git add Login/main.go Login/workspace.go Login/workspace_files.go Login/postgres_schema.sql Login/supabase_storage.go Login/supabase_storage_test.go Login/SUPABASE_STORAGE_SETUP_BN.md
git commit -m "Store workspace attachments in Supabase Storage"
git push origin main
```

Auto-Deploy On Commit চালু আছে। Deploy successful হওয়ার পরে Environment সেট করে redeploy করুন।

## ৫. পরীক্ষা
নিজের project workspace-এ একটি ছোট TXT/PDF upload করুন। Supabase Storage → research-files → projects-এ object এসেছে দেখুন। Website থেকে download করে মূল ফাইলের সঙ্গে মিলান। অন্য project-এর account ফাইল পাবে না। পুরোনো attachment-ও download করে দেখুন। Redeploy-এর পরে আবার download করুন।

ফাইল প্রতি 8 MiB এবং project প্রতি 64 MiB limit রাখা হয়েছে। Browser-এ object-এর public URL বা privileged key পাঠানো হয় না। Download একই /api/files/{id} route দিয়ে, permission check-এর পরে হয়। Upload বা database failure-এ object cleanup best-effort; process crash বা network failure-এ orphan object থাকতে পারে। পুরোনো database files সেখানেই থাকবে। Supabase quota এখনও প্রযোজ্য।

Mock Storage server দিয়ে upload/download, authorization, quota ও failure cleanup পরীক্ষা করা হয়েছে। আসল Supabase credential দিয়ে live test এই পরিবেশে করা হয়নি।
