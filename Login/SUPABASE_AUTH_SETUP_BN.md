# ResearchTheme: Supabase Auth + Gmail Custom SMTP

এই update signup, login, verification code এবং password reset-কে Supabase Auth-এর সঙ্গে যুক্ত করে। Student/Teacher dashboard, team, proposal এবং workspace-এর code অপরিবর্তিত রাখা হয়েছে।

## 1. Update বসানো

1. চালু Go server Ctrl+C দিয়ে বন্ধ করো।
2. বর্তমান project-এর backup রাখো, project folder-এর বাইরে।
3. ZIP-এর Login folder-এর ফাইলগুলো তোমার ResearchTheme-GitHub/Login folder-এ কপি করে একই নামের ফাইল replace করো। পুরো পুরোনো Login folder delete করবে না: এই ZIP শুধু পরিবর্তিত ফাইলের update।
4. নিজের `.env` এবং database রেখে দেবে। ZIP-এ এগুলো নেই।

## 2. Supabase dashboard

Authentication → Emails → SMTP Settings:
- Enable custom SMTP: ON
- Sender email ও Username: একই Gmail address
- Sender name: ResearchTheme
- Host: smtp.gmail.com
- Port: 587
- Password: Gmail App Password (সাধারণ Gmail password নয়)
- Save changes

Authentication → Sign In / Providers:
- Allow new users to sign up: ON
- Confirm email: ON
- Email: enabled
- Password policy: অন্তত 8 অক্ষর।

## 3. জরুরি: email templates-এ code বসাও

এই update link-এর বদলে email code ব্যবহার করে। Authentication → Emails → Templates থেকে দুইটি template বদলাও। Template-এর `{{ .Token }}` হুবহু রাখতে হবে।

### Confirm signup
Subject: Verify your ResearchTheme email
Body:
```html
<h2>Verify your ResearchTheme email</h2>
<p>Your verification code is:</p>
<p><strong>{{ .Token }}</strong></p>
<p>Return to ResearchTheme, open Verify email, and enter this code with your email address.</p>
<p>If you did not request this, ignore this email. Do not share this code.</p>
```

### Reset password
Subject: Reset your ResearchTheme password
Body:
```html
<h2>Reset your ResearchTheme password</h2>
<p>Your password recovery code is:</p>
<p><strong>{{ .Token }}</strong></p>
<p>Return to ResearchTheme's Reset password page and enter this code with your email and new password.</p>
<p>If you did not request this, ignore this email. Do not share this code.</p>
```

Save দুইবার করো। Supabase-এ নির্ধারিত OTP expiry প্রযোজ্য; পুরোনো code নয়, সর্বশেষ code ব্যবহার করবে। Reset fail হলে নতুন code চাইবে।

Authentication → URL Configuration → Site URL:
- Local test: http://localhost:8080
- Render live হওয়ার পর: Render-এর actual HTTPS URL
- এই OTP flow-তে redirect callback URL দরকার নেই।

## 4. Project URL এবং publishable key

Supabase project-এর Connect dialog বা Project Settings → API Keys থেকে Project URL এবং Publishable key নাও। Project URL `https://PROJECT_REF.supabase.co` ধরনের হবে। এটি PostgreSQL DATABASE_URL নয়।

Publishable key `sb_publishable_...` ব্যবহার করো। Secret key বা service_role key ব্যবহার করবে না। Key/password screenshot বা chat-এ পাঠাবে না।

তোমার `.env`-এ নিচের মানগুলো বসাও/বদলাও (duplicate key রেখো না):
```env
APP_ENV=development
APP_ADDR=127.0.0.1:8080
APP_URL=http://localhost:8080
EMAIL_MODE=supabase
SUPABASE_URL=https://YOUR_PROJECT_REF.supabase.co
SUPABASE_PUBLISHABLE_KEY=YOUR_PUBLISHABLE_KEY
```
আগের DATABASE_URL লাইন ঠিক যেমন আছে রাখবে। SMTP_USER/SMTP_PASS আর Go backend ব্যবহার করবে না যখন EMAIL_MODE=supabase; Gmail credentials শুধু Supabase dashboard-এ প্রয়োজন।

## 5. Local test

Login folder-এর terminal-এ:
```powershell
go test ./...
go run .
```
তারপর http://localhost:8080 খোলো।
1. নিজের ব্যবহারযোগ্য email দিয়ে signup করো।
2. Email-এর code Verify email page-এ বসাও।
3. আগের signup password ও সঠিক Student/Teacher role দিয়ে login করো।
4. Logout করে Forgot password পরীক্ষা করো। Email-এর recovery code ও নতুন password বসাও।
5. পুরোনো password fail এবং নতুন password login সফল হচ্ছে যাচাই করো।
6. Student ও Teacher দুটির dashboard এবং team/proposal access পরীক্ষা করো।

Mail না এলে Inbox/Spam, Supabase Authentication logs ও SMTP credentials দেখো। Gmail sending limit এবং Supabase rate limit প্রযোজ্য। এই package থেকে তোমার live Gmail delivery পরীক্ষা করা হয়নি।

## 6. পুরোনো account

`researchbridge.users`-এর পুরোনো account নিজে থেকে `auth.users`-এ তৈরি হয় না। পুরোনো user একই email দিয়ে নতুন signup করবে, email code verify করবে এবং password দিয়ে login করবে। সফল verified login-এর পর পুরোনো local profile ID-এর সঙ্গে Auth ID যুক্ত হবে: আগের name, role, team/proposal থাকে। Signup-এ অন্য role দিলেও পুরোনো role বদলাবে না। পুরোনো account-এর password ভুলে গেলে কিন্তু Auth-এ এখনও signup না করলে আগে signup লাগবে।

একবার link হলে local password login বন্ধ হয়; Auth user delete/recreate বা email পরিবর্তন করলে স্বয়ংক্রিয়ভাবে নতুন identity link হবে না। এই app-এ email পরিবর্তনের feature নেই। Owner-কে mapping reconcile করতে হবে। Local password mode-এ ফেরত যেও না—migrated account সেখানে login করতে পারবে না।

## 7. Render

Local test সফল হলে update ফাইলগুলো feature/online-deployment-এ commit/push করো।
- Branch: feature/online-deployment
- Root Directory: Login
- Build Command: go build -o app .
- Start Command: ./app
- APP_ENV: production
- APP_ADDR: 0.0.0.0:10000
- APP_URL: প্রকৃত Render HTTPS URL, শেষে slash নয়
- DATABASE_URL: আগের Supabase PostgreSQL connection string
- EMAIL_MODE: supabase
- SUPABASE_URL: project URL
- SUPABASE_PUBLISHABLE_KEY: publishable key

Render থেকে Supabase Auth-এ HTTPS request যায়; Gmail SMTP connection করে Supabase। Render-এ Gmail App Password প্রয়োজন নেই। SMTP setup successful না হলে website খুললেও signup/reset mail fail হতে পারে।

## 8. Session ও data নোট

- Password Supabase Auth-এ যাচাই হয়; নতুন password application DB-এ রাখা হয় না।
- App session HttpOnly cookie, server DB-তে token hash; সর্বোচ্চ এক ঘণ্টা। তারপর আবার login করতে হবে (refresh flow নেই)।
- এই app দিয়ে password reset করলে সব app session revoke হয়।
- Supabase dashboard দিয়ে user disable/delete বা অন্য client দিয়ে password পরিবর্তন করলে বিদ্যমান app session সর্বোচ্চ এক ঘণ্টা পর্যন্ত থাকতে পারে। জরুরি revoke-এর জন্য server-side sessions মুছতে হবে।
- পুরোনো local session Supabase mode-এ গ্রহণ করা হয় না।
- Dashboard permissions backend-এর local profile role দিয়ে নির্ধারিত; metadata পুরোনো role overwrite করতে পারে না।
- File uploads-এর permanent storage এই Auth update-এর অংশ নয়। Render-এর local upload file স্থায়ী নয়; পরে object storage setup করতে হবে।

## Sources
https://supabase.com/docs/guides/auth/auth-smtp
https://supabase.com/docs/guides/auth/auth-email-templates
https://supabase.com/docs/reference/javascript/auth-verifyotp
