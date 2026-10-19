# ResearchBridge Dashboard Update

## নতুন ফিচার
- Teacher sidebar ও student sidebar: প্রতিটি menu-তে আলাদা visual icon; selected/active menu color; hover/click feedback।
- Teacher header: notification bell, unread count, profile shortcut। Bell থেকে review queue-তে যাওয়া যাবে।
- Teacher/student research workspace profile: profile photo (JPG/PNG/WebP, 300 KB max) ও optional educational qualification; photo এবং qualification account database-এ save হয়।
- Supervision requests: বিদ্যমান proposal review flow-এর Accept / Request Corrections / Reject কাজ আগের মতোই থাকবে।
- Student team invitations: বিদ্যমান Review queue-র Accept / Decline flow আগের মতোই থাকবে। Teacher তার assigned team ও supervision requests দেখতে পারবেন।
- Notification preferences এবং mark all read আগের মতোই রাখা হয়েছে।

## চালানোর নিয়ম
1. আপনার project folder-এর backup রাখুন।
2. ZIP extract করে নতুন Login folder ব্যবহার করুন, অথবা source/assets files বর্তমান project-এ copy করুন।
3. আগে থেকে থাকা `.env`, database, এবং environment-specific settings আলাদা করে সংরক্ষণ করুন। এই ZIP-এ কোনো secret/config `.env` অথবা database data রাখা হয়নি।
4. `go run .` দিয়ে চালান। App start হলে নতুন `researcher_identity` table তৈরি হবে।
5. Teacher login করুন, `My Profile` > ছবি/qualification যোগ করে `Save profile & settings` চাপুন।
6. Teacher `Supervision Requests` > submitted proposal > `Review proposal` > Accept/Reject করতে পারবেন। Student-রা `Review queue`-তে team invite Accept/Decline করতে পারবেন।

## Verification
- `node --check` দিয়ে modified teacher dashboard script-এর syntax পরীক্ষা করা হয়েছে।
- Go dependencies অনলাইনে download করতে না পারায় এই environment-এ Go unit tests চালানো যায়নি। নিজের PC-তে `go test ./...` চালিয়ে যাচাই করুন।
