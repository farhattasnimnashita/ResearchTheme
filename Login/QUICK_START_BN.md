# ResearchBridge চালানো

## আগে থেকে project এবং Gmail settings থাকলে

`TEACHER_UPDATE_BN.md` অনুসরণ করুন। বর্তমান `.env` ও database রাখুন। নতুন update-এ সব `.go` files ও `web` folder লাগে।

```powershell
cd "D:\project-250\project test"
go run .
```

http://localhost:8080 খুলুন। Student আগের মেরুন dashboard, Teacher নতুন sky-blue dashboard পাবে। Dark mode toggle আছে।

## প্রথমবার setup

Go 1.23+ install করুন। ZIP extract করে ResearchBridge folder-এ terminal খুলুন:

```powershell
Copy-Item .env.example .env
```

শুধু নতুন `.env`-এ Gmail settings দিন:

```env
EMAIL_MODE=smtp
SMTP_USER=youraddress@gmail.com
SMTP_PASS=your_16_character_google_app_password
EMAIL_FROM=youraddress@gmail.com
```

SMTP_PASS হলো Gmail App Password, সাধারণ Gmail password নয়। আগে সেটআপ করা থাকলে নতুন password তৈরি করবেন না; আগের `.env` ব্যবহার করুন। কখনো এটি GitHub-এ দেবেন না।

```powershell
go mod download
go run .
```

Register করে email-এর link দিয়ে verify করুন, তারপর role বেছে login করুন। Forgot password থেকে email reset link পাওয়া যায়। Development-এ ইচ্ছাকৃত offline test করতে চাইলে `EMAIL_MODE=console` ব্যবহার করা যায়; সেটি email পাঠায় না। নতুন code `.env` না পেলে নীরবে console mode চালু করবে না।

## একই origin রাখুন

`.env`-এর APP_URL এবং browser URL মিলবে, default http://localhost:8080। Phone-এর localhost laptop নয়; phone test-এর জন্য laptop-এর LAN address দিয়ে আলাদা APP_URL/APP_ADDR configuration দরকার।
