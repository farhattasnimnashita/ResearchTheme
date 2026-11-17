# Team Management ও Profile UI Update

## Implemented workflow
1. Student A-এর তৈরি project-এর owner-ই invitation পাঠাতে পারবে।
2. Verified student directory থেকে invitation পাঠানো যায়।
3. Student B-এর My Team পেজে incoming invitation দেখা যায়; Accept/Decline করা যায়।
4. Accept সফল হলে transaction-এর মধ্যে `team_members`-এ membership যোগ হয় এবং invitation status `accepted` হয়।
5. Teacher dashboard-এর proposal card-এ project owner ও accepted team members-এর নাম দেখায়।
6. Project owner ছাড়া অন্য কেউ member remove করতে পারে না। Remove করার সময় `team_members` থেকে সদস্য বাদ যায় এবং invitation history-তে status `removed` হয়।
7. Team size limit ৩ জন (owner-সহ)। Pending invitation-ও capacity reserve করে; invitation পাঠানো ও গ্রহণ—দুই জায়গায় backend check আছে।
8. Invitation history-তে `pending`, `accepted`, `declined`, `removed` আলাদা status হিসেবে দেখানো হয়।

## Database
আলাদা membership table migration লাগেনি। Existing `team_members` এবং `workspace_records` (kind=`invitation`) ব্যবহার করা হয়েছে। তাই পুরোনো project ও invitation data বজায় থাকে।

## Frontend
- Student dashboard → My Team: project selector, student search, invite, incoming invitation accept/decline, invitation history এবং owner-only remove।
- Student dashboard → My Profile: redesigned profile header, profile picture area, account detail cards এবং private academic editor।
- Teacher dashboard: proposal card-এ team member names।

## Run / verify
Project directory-তে:
```bash
go test ./...
go run .
```
এই environment-এ Go dependency download network timeout হওয়ায় `go test ./...` সম্পন্ন করা যায়নি। JavaScript files `node --check` দিয়ে syntax-checked হয়েছে এবং Go files `gofmt` করা হয়েছে।
