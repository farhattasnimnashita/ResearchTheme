# Teacher notification update

এই update আপনার ২২ অক্টোবর দেওয়া Login(5).zip-এর ভিত্তিতে।

## কোন ফাইল লাগবে

আপনার project-এ নিচের ৬টি ফাইল একই জায়গায় copy/replace করুন:

- web/teacher-dashboard.html
- web/student-dashboard.html
- web/assets/teacher-dashboard.js
- web/assets/nashita-dashboard.js
- web/assets/teacher-notifications.js (নতুন)
- web/assets/teacher-notifications.css (নতুন)

নতুন ZIP-এ পুরো source আছে, কিন্তু এই কাজের জন্য উপরোক্ত ৬টি ফাইলই পরিবর্তিত। নিজের .env রেখে দিন। ZIP-এ .env, local database, executable ও backup রাখা হয়নি।

## কী পরিবর্তন হয়েছে

- Teacher dashboard প্রতি ১০ সেকেন্ডে নিজের authenticated notification API পরীক্ষা করে।
- নতুন unread notification এলে bell count ও sidebar count বাড়ে এবং popup দেখায়।
- Enable sound একবার চাপলে test chime শোনা যাবে; এরপর নতুন update-এর জন্য sound হবে। Mute দিয়ে বন্ধ করা যায়। Page reload হলে আবার Enable sound চাপতে হবে।
- প্রথমবার পুরোনো notification load হলে sound হবে না। একই notification polling বা navigation-এর কারণে আবার sound করবে না।
- Request popup-এর View update supervision request page খুলবে; অন্যান্য record notification তার সংশ্লিষ্ট module খুলবে।
- Request, submission/file, meeting/note, extension ও blocker-এর existing backend notification-গুলো এই ব্যবস্থায় আসে। Profile-এর notification preference প্রযোজ্য।
- Student supervisor form-এর missing title/area/abstract element access এবং undefined draft.id সরানো হয়েছে।
- Supervisor dropdown এখন directory ID নয়, linked verified account_id ব্যবহার করে। Teacher directory load হওয়ার পর project list load হয়। Project বদলালে সেই project-এর selection দেখায়।

## Request পাঠানোর সঠিক ধাপ

১. Student একটি project তৈরি/বাছাই করবে।
২. Supervisor অংশে teacher নির্বাচন করে Save Supervisor Selection চাপবে। এটি শুধু selection save করে, request পাঠায় না।
৩. Proposal অংশে problem, objectives, method, expected outcomes, scope ইত্যাদি পূরণ/save করবে।
৪. Project owner Submit for review চাপলে request submit হবে। Backend notification তখন তৈরি করে। শুধু draft save করলে teacher notification পাবে না।
৫. Teacher অন্য browser/InPrivate window-তে login করে dashboard খুলে Enable sound চাপবে। Profile-এ Requests notification চালু রাখবে।
৬. নতুন request submit করার প্রায় ১০ সেকেন্ডের মধ্যে bell count ও popup দেখবে; tab সক্রিয় থাকলে sound শুনবে। View update থেকে request review করতে পারবে।
৭. Teacher approve/correction/reject করলে student তার project status/feedback refresh করে দেখবে।

Browser tab বন্ধ থাকলে sound বা background push নেই। Background tab-এ browser timer/audio সীমিত করতে পারে। ফিরে এলে নতুন updates পরীক্ষা হবে। একাধিক teacher tab খোলা রাখলে প্রতি tab alert দিতে পারে।

## চালানো

Go server বন্ধ করে Login folder-এ:

    go run .

তারপর Ctrl+Shift+R দিয়ে Student ও Teacher page refresh করুন। Go embed-এর জন্য restart প্রয়োজন।

## যাচাই

JavaScript syntax ও simulated tests PASS: initial history silent, new notification sound, duplicate polling silent, read notification silent, mute, badge clearing, teacher role guard, request destination এবং student form-এর broken references।

এই পরিবেশে Go নেই, তাই backend compile বা live Supabase/browser end-to-end test করা হয়নি। Backend notification API ও database schema বদলানো হয়নি। আপনার terminal-এ go test ./... চালিয়ে এবং উপরের দুই account flow দিয়ে যাচাই করুন।

## সীমা

এখানে notification delivery frontend-এ যোগ হয়েছে; সব Teacher feature নতুন করে সম্পূর্ণ implement করা হয়নি। TEACHER_FEATURES_BN.md-এ বর্তমান code-এর সুবিধা, আংশিক কাজ ও বাকি অংশ দেওয়া আছে।
