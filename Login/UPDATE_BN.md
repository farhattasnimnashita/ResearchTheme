# ResearchBridge — নতুন login ও signup update

## সর্বশেষ ছোট update: Register link ও moving orbit

Login button-এর নিচে New here? Register an account link সব সময় দেখা যায়। Student/Teacher role বেছে link চাপলে সেই role-এর signup খুলবে। Globe-এর চারপাশে animated gold orbit tracks ও travelling light points আছে। Device-এ reduced motion চাইলে স্থির background থাকে।

**আগের full-screen/profile update ইতিমধ্যে বসানো থাকলে এবার শুধু web/assets/app.js ও web/assets/auth.css replace করুন।** তারপর Go server restart এবং browser-এ Ctrl+F5 দিন। .env/database/Go files বদলানোর প্রয়োজন নেই।


## এই update-এ কী আছে

1. Login/signup page-এর background পুরো browser viewport জুড়ে। দ্বিতীয় reference image-এর ivory, charcoal ও gold theme, books ও glowing globe থেকে নতুন background তৈরি করা হয়েছে। এটি live HTML form-এর পেছনে থাকে; screenshot-এর মতো একটি ছোট পাশের ছবিতে সীমাবদ্ধ নয়। ছোট screen-এ background cover করতে প্রান্ত কিছুটা crop হতে পারে। Form scroll করা যায়।
2. Student এবং Teacher দুজনের নতুন signup-এ Institution required। তালিকা থেকে বেছে নিতে বা নিজের institution লিখতে পারেন।
3. শুধু Student signup-এ Department, Academic year (1–6), Semester (1–12) required। Teacher signup-এ এগুলো দেখায় না বা store হয় না।
4. প্রথম page-এ Login এবং নিচে ছোট Register link দেখায়। Email database-এ না থাকলে বাড়তি Sign up করার suggestion এবং button দেখায়। Existing account-এর password বা role ভুল হলে signup বলতে থাকে না; নির্দিষ্ট login error দেখায়।
5. নতুন তথ্য database-এ সংরক্ষিত থাকে এবং dashboard-এর My profile-এ দেখা যায়।
6. SMTP/Resend email support আছে। SMTP-র resend message-এ আর console instructions দেখায় না।

## আগের project-এ update দেওয়ার নিয়ম

1. Server-এর terminal-এ Ctrl+C দিন। অন্য terminal-এ server চলছে কি না নিশ্চিত করুন।
2. পুরো project folder-এর একটি backup রাখুন।
3. ZIP extract করুন। ভিতরের ResearchBridge folder থেকে **সব .go files**, **go.mod**, **go.sum** এবং **web folder-এর contents** আপনার বর্তমান project folder-এ copy/replace করুন। নতুন profile.go এবং profile_test.go অবশ্যই দেবেন। web/index.html, web/assets/app.js, web/assets/auth.css, web/assets/discovery-background.png একসঙ্গে update দরকার।
4. নিজের **.env**, **researchbridge.db** এবং তার sidecar files delete/replace করবেন না। ZIP-এ আসল .env/database/credentials নেই। .env.example শুধু নমুনা।
5. Project root (go.mod থাকা folder)-এর terminal-এ চালান:

```powershell
go test ./...
go run .
```

6. Browser-এ http://localhost:8080 খুলে Ctrl+F5 দিন। Server embedded frontend serve করে, তাই code/image বদলানোর পর Go restart প্রয়োজন।

নিজের Go/JS/CSS-তে আলাদা custom changes থাকলে আগে diff মিলিয়ে নিন। এটি এই chat-এর Go starter ও Gmail SMTP update-এর ওপর ভিত্তি করে তৈরি।

## Database: পুরোনো data রাখা হবে

Startup-এ institution, department, academic_year ও semester columns প্রয়োজন হলে transaction-এর ভিতরে যোগ হয়। পুরোনো users, password hashes, verification, sessions, proposals ও feedback মুছে যায় না। বারবার server চালালেও columns duplicate হয় না।

পুরোনো signup-এ নতুন তথ্য আগে নেওয়া হয়নি, তাই তাদের My profile-এ Not provided দেখা স্বাভাবিক। পুরোনো account আগের password/role দিয়েই login করতে পারে। এই update-এ পুরোনো profile edit করার feature যোগ হয়নি। নতুন signup-এর fields বাধ্যতামূলক।

## নতুন Student পরীক্ষা

- Default Login page-এ একটি এখনও register না করা email ও পছন্দের password দিন → Log in চাপুন।
- No account found message-এর নিচে Sign up as student button চাপুন।
- Full name, Institution, Department, Year, Semester, Email ও Password দিন → Create account।
- Email verify করুন → Login করুন → My profile-এ নতুন তথ্য দেখুন।
- Password হবে ResearchBridge signup password; SMTP App Password নয়।

## নতুন Teacher পরীক্ষা

- Teacher role বেছে একটি নতুন email দিয়ে Login চেষ্টা করুন।
- Account না থাকলে Sign up as teacher চাপুন।
- Full name, Institution, Email, Password দিন। Student-only fields থাকবে না।
- Email verify করে Teacher role দিয়ে login করুন।

## Email settings

নিজের working .env বদলানোর দরকার নেই যদি ইতিমধ্যে Email mode: smtp দিয়ে email পৌঁছে থাকে। এই ZIP-এ smtp.go ও email.go আছে; আলাদা SMTP update আবার প্রয়োগ করবেন না।

Sender Gmail SMTP_USER ও EMAIL_FROM-এ একই হতে হবে। Verification recipient হলো user-এর signup email। নিজের API key/App Password chat বা GitHub-এ পাঠাবেন না।

## Background provenance

Built-in image generation দিয়ে দ্বিতীয় attached reference-এর style ধরে নতুন full-screen background বানানো হয়েছে। Prompt: ivory etched university architecture and charcoal/gold scholarly books, luminous orbital globe, no text/logos, dark upper-left heading space, quiet ivory right side for a live login form. Asset: web/assets/discovery-background.png. Original library.png retained for dashboard banners.

## Verification limits

Automated tests cover profile storage/validation, old-schema migration and data preservation, login error codes and signup suggestions, SMTP, auth and proposal flows. DOM interaction checks cover role-specific signup fields and the login-first flow. Actual Gmail delivery and visual browser QA need local verification; no user email credentials were used here.
