Teacher directory + student shortlist: teammate handoff

teachers.json: screenshot থেকে নেওয়া ১৮ জন শিক্ষকের structured data। এখনই teammate JSON parse করে directory UI, name/designation search এবং shortlist UI তৈরি করতে পারবে।
teacher_directory.sql: একই data shared PostgreSQL/Supabase database-এ রাখার script। Supabase SQL Editor-এ paste করে Run করতে পারবে। Existing account/profile tables বদলায় না; login account তৈরি করে না। Existing matching directory record overwrite করে না।

Screenshot-এর নাম/পদবি হুবহু রাখা হয়েছে; বর্তমান তথ্য independently verify করা হয়নি। Local-format ফোন +880 format-এ normalize হয়েছে। Website Link লেখা আছে কিন্তু URL নেই; ছবি, research interests, expertise এবং availability-ও পাওয়া যায়নি। Missing information অনুমান করে পূরণ করা হয়নি। verified=false মানে এই imported directory entry independently verified নয়, account approval status নয়।

প্রস্তাবিত integration (এগুলো এখনও তোমাদের backend-এ implement করা হয়নি):
GET /api/teacher-directory — authenticated student-কে teachers array ফেরত দেবে।
GET /api/me/teacher-shortlist — current student's saved teacher IDs ফেরত দেবে।
POST /api/me/teacher-shortlist — body: {"teacher_id":"sust-cse-003"}
DELETE /api/me/teacher-shortlist/sust-cse-003 — current student's shortlist থেকে বাদ দেবে।

Shortlist database table-এ student_id, teacher_id, created_at রাখবে। student_id হবে তোমাদের existing users table-এর ID-এর একই type ও foreign key; teacher_id হবে teacher_directory.id-এর foreign key। (student_id, teacher_id) unique বা composite primary key করবে, যাতে duplicate না হয়। Login করা student's ID server session থেকে নেবে, request body থেকে নয়। GET/DELETE-ও সেই student's records-এ সীমিত রাখবে। তোমাদের বর্তমান user schema না মিলিয়ে shortlist migration চালানো হয়নি।

Student frontend shared Go API দিয়ে পড়বে ও save করবে। localStorage-এ রাখলে দুই laptop-এ shortlist sync হবে না। SQL script RLS enable করে কোনো public write policy দেয় না; existing Go server-এর authorized database connection দিয়ে query করবে। Browser-এ DATABASE_URL বা service-role key রাখবে না।

এই directory IDs login user IDs নয়। Teacher register করলে verified account-এর সঙ্গে directory entry backend-এ link করতে হবে। শুধু imported email দেখে login/approval দেওয়া যাবে না। Teacher profile update-এর interests/expertise ওই linked directory-তে reflect করতে হবে। Shortlist করা আর supervision request accepted হওয়া আলাদা বিষয়।

প্রথমে নাম/পদবি/department দিয়ে filter ও manual shortlist করবে। Interest-based matching করতে হলে শিক্ষককে interests/expertise পূরণ করতে হবে; empty list মানে data unknown, teacher-এর কোনো interest নেই এমন নয়।

Example read (teammate development only):
const teachers = await fetch('/assets/teachers.json').then(r => r.json());
Production-এ authenticated GET /api/teacher-directory ব্যবহার করবে।

পরীক্ষা: JSON parse, unique email/ID count, phone format checked। Live database-এ import বা API integration এখানে করা হয়নি।
