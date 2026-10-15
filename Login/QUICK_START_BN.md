**আগের project update করলে আগে UPDATE_BN.md পড়ুন। নিজের .env ও database রাখুন।**

# ResearchBridge চালানোর সহজ নিয়ম

## ১. Folder খুলুন

ZIP extract করে `ResearchBridge` folder VS Code-এ খুলুন। সব Go file, `go.mod`, `go.sum` আর `web` folder একসঙ্গে থাকতে হবে। এটি নতুন project; আগের project/database-এর ওপর সরাসরি replace করবেন না।

## ২. VS Code-এর PowerShell terminal-এ দিন

```powershell
Copy-Item .env.example .env
go mod download
go run .
```

Go install না থাকলে আগে install করে VS Code আবার খুলুন। এরপর browser-এ **http://localhost:8080** খুলুন। Server চালু থাকা terminal বন্ধ করবেন না।

## ৩. প্রথমে Teacher account পরীক্ষা করুন

- Login page-এ Teacher বেছে নতুন email/password দিয়ে Login চাপুন। No account found দেখালে Sign up as teacher চাপুন → name, institution, email ও নতুন website password দিন।
- Default development mode-এ inbox-এ email যাবে না। Go terminal-এ verification link দেখাবে।
- পুরো link browser-এ খুলে **Verify email** চাপুন।
- Teacher login থেকে login করুন। Teacher dashboard দেখাবে।

## ৪. Student account পরীক্ষা করুন

- Logout করুন অথবা অন্য browser profile খুলুন।
- Student বেছে আলাদা নতুন email দিয়ে Login চেষ্টা করুন; signup suggestion থেকে Institution, Department, Year, Semester-সহ register করুন।
- Terminal-এর নতুন link খুলে verify করুন, তারপর Student login করুন।
- **New proposal** → তথ্য লিখে **Save draft**।
- **Submit for review** → verified Teacher নির্বাচন → Submit।
- Teacher login করে Review queue থেকে review ও feedback দিন।
- Student dashboard refresh করলে feedback দেখাবে।

একই email দিয়ে দুটি role-এর account হবে না। Database নিজের কম্পিউটারে `researchbridge.db` file-এ থাকে। Server বন্ধ করলেও account থাকবে।

## ৫. সত্যিকারের email পাঠাতে

আগে SMTP দিয়ে email পৌঁছে থাকলে নিজের .env যেমন আছে তেমন রাখুন। Gmail SMTP-র নমুনা:

```dotenv
EMAIL_MODE=smtp
SMTP_USER=youraddress@gmail.com
SMTP_PASS=your_16_character_google_app_password
EMAIL_FROM=ResearchBridge <youraddress@gmail.com>
```

একই sender Gmail দুই জায়গায় দিন। App Password নিজে স্থানীয়ভাবে বসাবেন, chat/GitHub-এ দেবেন না। নতুন code-এ Gmail SMTP support ইতিমধ্যে আছে। Resend বিকল্প ব্যবহার করলে README-এর section 3 পড়ুন। Configuration বদলালে Ctrl+C দিয়ে Go বন্ধ করে আবার go run . দিন।

localhost link অন্য কম্পিউটারে আপনার server খুলবে না। শুরুতে laptop-এর browser থেকেই email link খুলে verify করুন।

## ৬. GitHub-এ কাজ ভাগ

- একজন frontend: `web/index.html`, `web/assets/app.js`, `web/assets/style.css`।
- অন্যজন backend: `main.go`, `email.go`, `projects.go`, `main_test.go`।
- একজন `feature/frontend`, অন্যজন `feature/backend` branch ব্যবহার করতে পারেন।
- GitHub repository create, collaborator invite, clone ও push-এর পূর্ণ commands README section 5-এ আছে।
- `.env` এবং database push করবেন না; `.gitignore` এগুলো বাদ দেয়।
- দুজনের localhost database আলাদা থাকবে। GitHub code ভাগ করে, live database নয়।

GitHub-এ স্বয়ংক্রিয় push করা হয়নি। সঠিক repository URL ছাড়া কোন repository-তে পাঠাতে হবে তা নির্ধারিত নয়।
