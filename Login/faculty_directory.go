package main

import (
	"net/http"
	"time"
)

type facultySeed struct{ name, designation, phone, email string }

var sustFaculty = []facultySeed{
	{"Mohammad Shahidur Rahman, PhD", "Professor", "+8801914930807", "rahmanms@sust.edu"},
	{"Dr Mohammad Reza Selim", "Professor", "+8801972357830", "selim@sust.edu"},
	{"M. Jahirul Islam, PhD., PEng.", "Professor", "+8801770348185", "jahir-cse@sust.edu"},
	{"Mohammad Abdullah Al Mumin, PhD", "Professor", "+8801711445110", "mumin-cse@sust.edu"},
	{"Dr. Md Forhad Rabbi, SMIEEE", "Professor & Head", "+8801844175805", "frabbi-cse@sust.edu"},
	{"Md Masum", "Professor", "+8801919736248", "masum-cse@sust.edu"},
	{"Dr. Husne Ara Chowdhury", "Professor", "+8801558304163", "husna-cse@sust.edu"},
	{"Dr. Sadia Sultana", "Professor", "+8801911089612", "sadia-cse@sust.edu"},
	{"Mahruba Sharmin Chowdhury", "Associate Professor", "+8801917566699", "mahruba-cse@sust.edu"},
	{"Md. Eamin Rahman", "Assistant Professor", "+8801677014633", "eamin-cse@sust.edu"},
	{"A.K.M. Fakhrul Hossain", "Lecturer", "+8801521333411", "fakhrul-cse@sust.edu"},
	{"Md. Shadmim Hasan Sifat", "Lecturer", "+8801571291835", "shadmim-cse@sust.edu"},
	{"Md. Mehedi Hasan", "Lecturer", "+8801932981532", "mehedi-cse@sust.edu"},
	{"Abdullah Al Thaki", "Lecturer", "01516178219", "toki-cse@sust.edu"},
	{"Md. Shymon Islam", "Lecturer", "+8801718550548", "shymon-cse@sust.edu"},
	{"MD Fahimul Islam Fahim", "Lecturer", "01893639181", "fahim-cse@sust.edu"},
	{"Mainul Hasan Mohsin", "Lecturer", "01521565271", "mainul-cse@sust.edu"},
	{"Mufassir Ahmad Chowdhury", "Lecturer", "01771144308", "mufassir-cse@sust.edu"},
}

func setupFacultyDirectory(db *Database) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS faculty_directory (id BIGINT PRIMARY KEY, name TEXT NOT NULL, designation TEXT NOT NULL, phone TEXT NOT NULL DEFAULT '', email TEXT NOT NULL UNIQUE, website TEXT NOT NULL DEFAULT '', department TEXT NOT NULL DEFAULT 'Computer Science & Engineering', source TEXT NOT NULL DEFAULT 'User-provided SUST department screenshots')`,
		`CREATE TABLE IF NOT EXISTS faculty_ratings (faculty_id BIGINT NOT NULL REFERENCES faculty_directory(id) ON DELETE CASCADE, student_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, supervision BIGINT NOT NULL CHECK(supervision BETWEEN 1 AND 5), monitoring BIGINT NOT NULL CHECK(monitoring BETWEEN 1 AND 5), guidelines BIGINT NOT NULL CHECK(guidelines BETWEEN 1 AND 5), helpfulness BIGINT NOT NULL CHECK(helpfulness BETWEEN 1 AND 5), understanding BIGINT NOT NULL CHECK(understanding BETWEEN 1 AND 5), updated_at BIGINT NOT NULL, PRIMARY KEY(faculty_id,student_id))`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	// Stable IDs prevent duplicates across team laptops, restarts and deployments.
	for i, f := range sustFaculty {
		_, err := db.Exec(`INSERT INTO faculty_directory(id,name,designation,phone,email) VALUES(?,?,?,?,?) ON CONFLICT(id) DO NOTHING`, i+1, f.name, f.designation, f.phone, f.email)
		if err != nil {
			return err
		}
	}
	return nil
}
func (a *App) facultyDirectory(w http.ResponseWriter, r *http.Request) {
	if _, e := a.user(r); e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	rows, e := a.db.Query(`SELECT id,name,designation,phone,email,website,department FROM faculty_directory ORDER BY id`)
	if e != nil {
		fail(w, 500, "Faculty list unavailable")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var name, designation, phone, email, website, department string
		if rows.Scan(&id, &name, &designation, &phone, &email, &website, &department) != nil {
			fail(w, 500, "Faculty list unavailable")
			return
		}
		out = append(out, map[string]any{"id": id, "name": name, "designation": designation, "phone": phone, "email": email, "website": website, "department": department})
	}
	if rows.Err() != nil {
		fail(w, 500, "Faculty list unavailable")
		return
	}
	reply(w, 200, out)
}
func (a *App) facultyRatings(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if u.Role != "student" {
		fail(w, 403, "Student-only ratings")
		return
	}
	rows, e := a.db.Query(`SELECT f.id,f.name,f.designation,f.phone,f.email, COUNT(v.student_id), COALESCE(AVG((v.supervision+v.monitoring+v.guidelines+v.helpfulness+v.understanding)/5.0),0),COALESCE(AVG(v.supervision),0),COALESCE(AVG(v.monitoring),0),COALESCE(AVG(v.guidelines),0),COALESCE(AVG(v.helpfulness),0),COALESCE(AVG(v.understanding),0) FROM faculty_directory f LEFT JOIN faculty_ratings v ON v.faculty_id=f.id GROUP BY f.id,f.name,f.designation,f.phone,f.email ORDER BY f.id`)
	if e != nil {
		fail(w, 500, "Ratings unavailable")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, count int64
		var name, designation, phone, email string
		var overall, s, m, g, h, under float64
		if rows.Scan(&id, &name, &designation, &phone, &email, &count, &overall, &s, &m, &g, &h, &under) != nil {
			fail(w, 500, "Ratings unavailable")
			return
		}
		out = append(out, map[string]any{"id": id, "name": name, "designation": designation, "phone": phone, "email": email, "rating_count": count, "overall": overall, "supervision_average": s, "monitoring_average": m, "guidelines_average": g, "helpfulness_average": h, "understanding_average": under})
	}
	if rows.Err() != nil {
		fail(w, 500, "Ratings unavailable")
		return
	}
	reply(w, 200, out)
}
func (a *App) submitFacultyRating(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if u.Role != "student" || !u.Verified {
		fail(w, 403, "Verified students only")
		return
	}
	var data struct {
		FacultyID     int64 `json:"faculty_id"`
		Supervision   int64 `json:"supervision"`
		Monitoring    int64 `json:"monitoring"`
		Guidelines    int64 `json:"guidelines"`
		Helpfulness   int64 `json:"helpfulness"`
		Understanding int64 `json:"understanding"`
	}
	if !decode(w, r, &data) {
		return
	}
	for _, v := range []int64{data.Supervision, data.Monitoring, data.Guidelines, data.Helpfulness, data.Understanding} {
		if v < 1 || v > 5 {
			fail(w, 400, "All scores must be between 1 and 5")
			return
		}
	}
	var exists int
	err := a.db.QueryRow(`SELECT COUNT(*) FROM faculty_directory WHERE id=?`, data.FacultyID).Scan(&exists)
	if err != nil || exists != 1 {
		fail(w, 404, "Faculty member not found")
		return
	}
	_, err = a.db.Exec(`INSERT INTO faculty_ratings(faculty_id,student_id,supervision,monitoring,guidelines,helpfulness,understanding,updated_at) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(faculty_id,student_id) DO UPDATE SET supervision=excluded.supervision,monitoring=excluded.monitoring,guidelines=excluded.guidelines,helpfulness=excluded.helpfulness,understanding=excluded.understanding,updated_at=excluded.updated_at`, data.FacultyID, u.ID, data.Supervision, data.Monitoring, data.Guidelines, data.Helpfulness, data.Understanding, time.Now().Unix())
	if err != nil {
		fail(w, 500, "Could not save rating")
		return
	}
	reply(w, 200, map[string]string{"status": "saved"})
}
