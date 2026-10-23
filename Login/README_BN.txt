ResearchBridge Landing Page

ফলাফল:
http://localhost:8080/ খুললে নতুন পরিচিতি page আসবে।
Start Here চাপলে /login/student-এ আগের sign-in page আসবে।
সেখানে Student/Teacher role বেছে নেওয়া যায়।
Supervisor card-এর button সরাসরি /login/teacher খুলবে।

সহজ ইনস্টল:
১. ZIP extract করুন।
২. Extract করা folder-এ PowerShell খুলুন (যেখানে Install-Landing.ps1 আছে)।
৩. চালান:
   .\Install-Landing.ps1

আপনার Login folder অন্য জায়গায় হলে:
   .\Install-Landing.ps1 -LoginPath 'D:\your-project\Login'

Script চালানো blocked হলে policy বদলানোর দরকার নেই; নিচের manual ধাপ করুন।

Manual ইনস্টল:
১. ZIP-এর Login/web/landing.html আপনার Login/web/ folder-এ রাখুন।
২. ZIP-এর Login/web/assets/landing.css এবং landing-campus.png আপনার Login/web/assets/ folder-এ রাখুন।
৩. Login/main.go-তে এই লাইন খুঁজুন:
   page := "web/index.html"

তার ঠিক নিচে যোগ করুন:
   if p == "/" {
       page = "web/landing.html"
   }

আগে এই অংশ যোগ করা থাকলে দ্বিতীয়বার যোগ করবেন না।
কোনো পুরোনো main.go বা app.js পুরো replace করতে হবে না।

Server চালু থাকলে তার terminal-এ Ctrl+C, তারপর:
   gofmt -w main.go
   go run .

Browser-এ http://localhost:8080/ খুলে Ctrl+Shift+R চাপুন।
শুধু refresh যথেষ্ট নয়: Go embed-এর জন্য server restart জরুরি।

যাচাই:
- Homepage-এ sign-in form-এর আগে landing page দেখা যায়।
- Start Here -> sign-in page।
- Teacher workspace -> teacher sign-in।
- How it works -> নিচের research journey section।
- পুরোনো register/verify/reset/dashboard routes আগের মতো আছে।

এটি আপনার আগে দেওয়া Login(5) code-এর routing অনুযায়ী তৈরি।
HTML/CSS ও link/static asset checks করা হয়েছে। এই পরিবেশে Go এবং browser runtime দিয়ে চালিয়ে পরীক্ষা করা যায়নি।

অনলাইনে update-এর জন্য repository root থেকে:
   git add Login/main.go Login/web/landing.html Login/web/assets/landing.css Login/web/assets/landing-campus.png
   git diff --cached --stat
   git commit -m "Add ResearchBridge public landing page"
   git push origin main
