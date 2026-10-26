package main

import (
	"encoding/json"
	"testing"
)

func TestCapacityApprovalAndPlacement(t *testing.T) {
	a, owner, teacher, outsider, id := workspaceFixture(t)
	expect(t, call(t, a, "GET", "/api/supervision-capacity", nil, owner), 403)
	expect(t, call(t, a, "POST", "/api/supervision-capacity", map[string]int{"max_teams": 1, "max_students": 1}, teacher), 200)
	path := "/api/projects/" + itoa(id) + "/review"
	expect(t, call(t, a, "POST", path, map[string]string{"status": "approved"}, outsider), 409)
	expect(t, call(t, a, "POST", path, map[string]string{"status": "approved"}, teacher), 200)
	expect(t, call(t, a, "POST", path, map[string]string{"status": "approved"}, teacher), 409)
	var c SupervisionCapacity
	w := call(t, a, "GET", "/api/supervision-capacity", nil, teacher)
	expect(t, w, 200)
	if err := json.Unmarshal(w.Body.Bytes(), &c); err != nil {
		t.Fatal(err)
	}
	if c.UsedTeams != 1 || c.UsedStudents != 1 || !c.Full {
		t.Fatalf("bad counts: %+v", c)
	}
	var p map[string]any
	w = call(t, a, "GET", "/api/student-placement", nil, owner)
	expect(t, w, 200)
	json.Unmarshal(w.Body.Bytes(), &p)
	if p["supervision"] != "Supervisor assigned" {
		t.Fatal(p)
	}
	// Lowering limits never removes an already-approved team.
	expect(t, call(t, a, "POST", "/api/supervision-capacity", map[string]int{"max_teams": 0, "max_students": 0}, teacher), 200)
	var status string
	a.db.QueryRow("SELECT status FROM projects WHERE id=?", id).Scan(&status)
	if status != "approved" {
		t.Fatal(status)
	}
}
func TestCapacityRejectsOversizedTeamAndReleasesPending(t *testing.T) {
	a, owner, teacher, _, id := workspaceFixture(t)
	var tid, uid int64
	if e := a.db.QueryRow("SELECT teacher_id,student_id FROM projects WHERE id=?", id).Scan(&tid, &uid); e != nil {
		t.Fatal(e)
	}
	member, e := a.db.InsertID("INSERT INTO users(name,email,password_hash,role,verified,created_at) VALUES(?,?,?,'student',1,0)", "Member", "member@example.com", "unused")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = a.db.Exec("INSERT INTO team_members(project_id,user_id) VALUES(?,?)", id, member); e != nil {
		t.Fatal(e)
	}
	expect(t, call(t, a, "POST", "/api/supervision-capacity", map[string]int{"max_teams": 1, "max_students": 1}, teacher), 200)
	path := "/api/projects/" + itoa(id) + "/review"
	expect(t, call(t, a, "POST", path, map[string]string{"status": "approved"}, teacher), 409)
	c, e := readCapacity(a.db, tid)
	if e != nil {
		t.Fatal(e)
	}
	if c.UsedTeams != 0 || c.UsedStudents != 0 {
		t.Fatal(c)
	}
	expect(t, call(t, a, "POST", "/api/supervision-capacity", map[string]int{"max_teams": 1, "max_students": 2}, teacher), 200)
	expect(t, call(t, a, "POST", path, map[string]string{"status": "approved"}, teacher), 200)
	c, e = readCapacity(a.db, tid)
	if e != nil || c.UsedStudents != 2 || !c.Full {
		t.Fatalf("%+v %v", c, e)
	}
	tx, e := a.db.Begin()
	if e != nil {
		t.Fatal(e)
	}
	if e = checkJoinCapacity(tx, id, 99999); e == nil {
		t.Fatal("joining a full supervisor allowed")
	}
	tx.Rollback()
	// A member counted as both owner and member must still only count once.
	a.db.Exec("INSERT INTO team_members(project_id,user_id) VALUES(?,?)", id, uid)
	c, e = readCapacity(a.db, tid)
	if e != nil || c.UsedStudents != 2 {
		t.Fatalf("duplicate count: %+v %v", c, e)
	}
	w := call(t, a, "GET", "/api/student-placement", nil, owner)
	expect(t, w, 200)
	var p map[string]any
	json.Unmarshal(w.Body.Bytes(), &p)
	if p["membership"] != "Team joined" {
		t.Fatal(p)
	}
}
func TestCapacityFullCannotBeBypassedAndCorrectionsStillAllowed(t *testing.T) {
	a, _, teacher, _, id := workspaceFixture(t)
	expect(t, call(t, a, "POST", "/api/supervision-capacity", map[string]int{"max_teams": 0, "max_students": 10}, teacher), 200)
	path := "/api/projects/" + itoa(id) + "/review"
	expect(t, call(t, a, "POST", path, map[string]string{"status": "approved"}, teacher), 409)
	expect(t, call(t, a, "POST", path, map[string]string{"status": "changes_requested"}, teacher), 400)
	expect(t, call(t, a, "POST", path, map[string]string{"status": "changes_requested", "feedback": "Please clarify the methodology."}, teacher), 200)
	expect(t, call(t, a, "POST", "/api/supervision-capacity", map[string]int{"max_teams": -1, "max_students": 10}, teacher), 400)
}

func TestConcurrentApprovalsRespectOneRemainingTeam(t *testing.T) {
	a, _, teacher, _, first := workspaceFixture(t)
	var tid int64
	if e := a.db.QueryRow("SELECT teacher_id FROM projects WHERE id=?", first).Scan(&tid); e != nil {
		t.Fatal(e)
	}
	uid, e := a.db.InsertID("INSERT INTO users(name,email,password_hash,role,verified,created_at) VALUES(?,?,?,'student',1,0)", "Second owner", "second@example.com", "unused")
	if e != nil {
		t.Fatal(e)
	}
	second, e := a.db.InsertID("INSERT INTO projects(student_id,teacher_id,title,area,abstract,status,updated_at) VALUES(?,?,?,?,?,'submitted',0)", uid, tid, "Second project", "Data", "A detailed research abstract")
	if e != nil {
		t.Fatal(e)
	}
	expect(t, call(t, a, "POST", "/api/supervision-capacity", map[string]int{"max_teams": 1, "max_students": 10}, teacher), 200)
	codes := make(chan int, 2)
	for _, id := range []int64{first, second} {
		go func(id int64) {
			codes <- call(t, a, "POST", "/api/projects/"+itoa(id)+"/review", map[string]string{"status": "approved"}, teacher).Code
		}(id)
	}
	one, two := <-codes, <-codes
	if !((one == 200 && two == 409) || (one == 409 && two == 200)) {
		t.Fatalf("concurrent results %d %d", one, two)
	}
	c, e := readCapacity(a.db, tid)
	if e != nil || c.UsedTeams != 1 {
		t.Fatalf("capacity oversubscribed %+v %v", c, e)
	}
}
