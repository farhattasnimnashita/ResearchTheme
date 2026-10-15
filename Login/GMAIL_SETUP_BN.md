# Gmail SMTP — শুরু থেকে সেটআপ

এটি এই chat-এ দেওয়া ResearchBridge-Go-Starter.zip-এর supplementary update। Original project-এর পাশাপাশি এই update রাখুন। Domain বা Resend API key লাগবে না। Gmail account-এর App Password লাগবে। Gmail-এর sending limits এবং anti-abuse controls প্রযোজ্য; unlimited email service নয়।

## ১. Google Account প্রস্তুত করুন

যে Gmail address থেকে verification পাঠাবেন, সেটিতে login করুন। চাইলে project-এর জন্য আলাদা Gmail account ব্যবহার করুন।

https://myaccount.google.com/security খুলুন → 2-Step Verification নির্বাচন করুন → Google-এর ধাপ অনুসরণ করে চালু করুন।

এরপর https://myaccount.google.com/apppasswords খুলুন। দরকার হলে আবার login করুন। App name হিসেবে ResearchBridge দিয়ে Create করুন। Google যে ১৬ অক্ষরের App Password দেখাবে, সেটি নিজের কাছে copy করুন। Gmail-এর সাধারণ password ব্যবহার করবেন না। Secret এখানে chat-এ বা screenshot-এ পাঠাবেন না।

App Passwords না থাকলে 2-Step Verification চালু কি না দেখুন। School/work account policy, security-key-only setup অথবা Advanced Protection-এর কারণে অপশনটি না-ও থাকতে পারে। এ ক্ষেত্রে account protection বন্ধ না করে ব্যক্তিগত Gmail account-এর উপযোগিতা যাচাই করুন অথবা OAuth-ভিত্তিক integration প্রয়োজন হবে।

## ২. Code update দিন

Server চললে সেই terminal-এ Ctrl+C দিন। আগে নিজের পরিবর্তিত code-এর backup রাখুন।

এই update folder থেকে নিচের চারটি file আপনার project-এর root-এ দিন—যেখানে go.mod আছে:

- main.go (আগেরটি replace)
- email.go (আগেরটি replace)
- smtp.go (নতুন)
- smtp_test.go (নতুন tests)

এটি আগের দেওয়া starter-এর জন্য। নিজের main.go/email.go-তে অন্য পরিবর্তন করে থাকলে backup রেখে পরিবর্তনগুলো মিলিয়ে নিন। Existing web folder, projects.go, go.mod, go.sum, main_test.go রাখুন। এই update কোনো database schema বদলায় না। নিজের .env ও researchbridge.db delete/replace করবেন না।

## ৩. নিজের .env edit করুন

.env.smtp.example শুধু নমুনা; সেটি সরাসরি আপনার secret settings-এর ওপর copy করবেন না। Project-এর .env-এ নিচের settings বসান:

```dotenv
APP_ENV=development
APP_ADDR=127.0.0.1:8080
APP_URL=http://localhost:8080
DB_PATH=researchbridge.db
EMAIL_MODE=smtp
SMTP_USER=youraddress@gmail.com
SMTP_PASS=your_16_character_app_password
EMAIL_FROM=ResearchBridge <youraddress@gmail.com>
```

- দুই জায়গার youraddress@gmail.com-এ একই নিজের পূর্ণ Gmail address বসান (ছোট হাতের অক্ষরে)। এটি sender account; Student/Teacher recipient email আলাদা হতে পারে।
- SMTP_PASS-এ Google-এর ১৬ অক্ষরের App Password দিন। উদাহরণের placeholder ব্যবহার করলে server সেটি গ্রহণ করবে না। Google চার অক্ষর করে space দিয়ে দেখালে space বাদ দিন (code সাধারণ spaces-ও সরিয়ে নেয়)।
- EMAIL_MODE=smtp রাখুন। RESEND_API_KEY থাকলেও এই mode-এ তা ব্যবহার হবে না।
- Gmail SMTP host smtp.gmail.com এবং port 587 এই update-এ নির্ধারিত আছে। আলাদা SMTP_HOST/SMTP_PORT লাগবে না। STARTTLS ও certificate যাচাই বাধ্যতামূলক।
- .env file save করুন (Ctrl+S)। এটি GitHub-এ দেবেন না।

## ৪. চালান ও পরীক্ষা করুন

VS Code-এর project terminal-এ:

```powershell
go test ./...
go run .
```

Email mode: smtp দেখাবে। এরপর http://localhost:8080 খুলুন। New user হলে Sign up করুন। আগে account তৈরি করা থাকলে Resend verification চাপুন (এখানে Resend শব্দটির অর্থ আবার link পাঠানো, provider-এর নাম নয়)। Recipient inbox ও Spam folder দেখুন → email-এর link খুলুন → Verify email চাপুন → সঠিক Student/Teacher role দিয়ে login করুন।

Verification email project-এর sender Gmail থেকে যাবে। প্রতিটি Student/Teacher-এর নিজের Gmail App Password দিতে হবে না। শুধু server-এর sender account configure করবেন।

## ৫. সমস্যা হলে

- Port 8080 already in use: আগের server চলা terminal-এ Ctrl+C দিন, তারপর আবার চালান। কোন process জানেন না হলে সব process বন্ধ করবেন না; port-এর process আগে শনাক্ত করুন।
- Email mode: resend/console দেখালে: পুরোনো terminal-এর environment variables .env-কে override করতে পারে। নতুন PowerShell terminal খুলুন অথবা নির্দিষ্ট EMAIL_MODE/SMTP_USER/SMTP_PASS/EMAIL_FROM পুরোনো environment values সরিয়ে আবার চালান। Secret values print করবেন না।
- Account already exists: নতুন registration নয়, Resend verification ব্যবহার করুন।
- SMTP_PASS error: normal Gmail password নয়; App Password তৈরি করুন। Google account password বদলালে App Password বাতিল হতে পারে, তখন নতুনটি দিন।
- Delivery failed: একই sender Gmail ব্যবহার হচ্ছে কি না, App Password ঠিক কি না এবং firewall/network outbound port 587 block করছে কি না দেখুন। Gmail Sent folder/Google security alerts দেখুন। Provider গ্রহণ করলেও inbox-এর বদলে Spam-এ যেতে পারে।
- Too many attempts: কিছুক্ষণ অপেক্ষা করুন। Gmail sending restrictions এবং starter-এর request limit প্রযোজ্য।

## ৬. দুই laptop ও GitHub

Code commit করুন; .env, SMTP_PASS, database commit করবেন না। Teammate-এর laptop-এ code চালাতে আলাদা করে local .env configure করতে হবে। Google password/App Password GitHub দিয়ে share করবেন না।

localhost verification link যে কম্পিউটারে খোলা হয় তার server-এ যায়। Teammate-এর অন্য কম্পিউটার থেকে একই account verify করতে চাইলে একটি shared hosted server/address দরকার। Email পাঠানো মানেই localhost website public হয়ে যাওয়া নয়।

## Verification performed

Go tests cover existing auth and project workflows, Gmail config validation, email MIME/header encoding, cleartext SMTP rejection and SMTP delivery through a local STARTTLS test server. No actual Gmail email was sent; real delivery requires your local App Password. All Go tests, go vet, and the Windows amd64 build passed.

## Official references

https://support.google.com/accounts/answer/185833
https://developers.google.com/workspace/gmail/imap/imap-smtp
