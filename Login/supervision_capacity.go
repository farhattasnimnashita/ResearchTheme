package main

import (
	"database/sql"
	"errors"
	"net/http"
)

type capacityReader interface{ QueryRow(string, ...any) *sql.Row }
type SupervisionCapacity struct {
	MaxTeams          int  `json:"max_teams"`
	MaxStudents       int  `json:"max_students"`
	UsedTeams         int  `json:"used_teams"`
	UsedStudents      int  `json:"used_students"`
	RemainingTeams    int  `json:"remaining_teams"`
	RemainingStudents int  `json:"remaining_students"`
	Full              bool `json:"full"`
}

// Owners count as students too; UNION prevents counting a person twice.
const supervisedStudentsSQL = `SELECT COUNT(DISTINCT uid) FROM (
 SELECT student_id AS uid FROM projects WHERE teacher_id=? AND status='approved'
 UNION SELECT tm.user_id AS uid FROM team_members tm JOIN projects p ON p.id=tm.project_id WHERE p.teacher_id=? AND p.status='approved'
) members`

func readCapacity(q capacityReader, teacherID int64) (SupervisionCapacity, error) {
	c := SupervisionCapacity{MaxTeams: 5, MaxStudents: 20}
	err := q.QueryRow("SELECT max_teams,max_students FROM supervision_capacity WHERE teacher_id=?", teacherID).Scan(&c.MaxTeams, &c.MaxStudents)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return c, err
	}
	if err = q.QueryRow("SELECT COUNT(*) FROM projects WHERE teacher_id=? AND status='approved'", teacherID).Scan(&c.UsedTeams); err != nil {
		return c, err
	}
	if err = q.QueryRow(supervisedStudentsSQL, teacherID, teacherID).Scan(&c.UsedStudents); err != nil {
		return c, err
	}
	c.RemainingTeams = max(0, c.MaxTeams-c.UsedTeams)
	c.RemainingStudents = max(0, c.MaxStudents-c.UsedStudents)
	c.Full = c.RemainingTeams == 0 || c.RemainingStudents == 0
	return c, nil
}
func (a *App) supervisionCapacity(w http.ResponseWriter, r *http.Request) {
	u, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if u.Role != "teacher" {
		fail(w, 403, "Only teachers can manage supervision capacity.")
		return
	}
	if r.Method == "POST" {
		var in struct {
			MaxTeams    int `json:"max_teams"`
			MaxStudents int `json:"max_students"`
		}
		if !decode(w, r, &in) {
			return
		}
		if in.MaxTeams < 0 || in.MaxTeams > 1000 || in.MaxStudents < 0 || in.MaxStudents > 10000 {
			fail(w, 400, "Enter teams from 0 to 1000 and students from 0 to 10000.")
			return
		}
		tx, e := a.db.Begin()
		if e != nil {
			fail(w, 500, "Could not save capacity.")
			return
		}
		defer tx.Rollback()
		// Uses the same serialized transaction as approvals and team joins.
		if _, e = tx.Exec(`INSERT INTO supervision_capacity(teacher_id,max_teams,max_students) VALUES(?,?,?) ON CONFLICT(teacher_id) DO UPDATE SET max_teams=excluded.max_teams,max_students=excluded.max_students`, u.ID, in.MaxTeams, in.MaxStudents); e != nil {
			fail(w, 500, "Could not save capacity.")
			return
		}
		if e = tx.Commit(); e != nil {
			fail(w, 500, "Could not save capacity.")
			return
		}
	}
	c, err := readCapacity(a.db, u.ID)
	if err != nil {
		fail(w, 500, "Could not load capacity.")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	reply(w, 200, c)
}

// Approval checks run inside the same transaction as the status change.
func checkApprovalCapacity(tx *Transaction, teacherID, projectID int64) error {
	var conflicts int
	err := tx.QueryRow(`SELECT COUNT(*) FROM projects other WHERE other.status='approved' AND other.id<>? AND EXISTS (
 SELECT 1 FROM users u WHERE (u.id=other.student_id OR EXISTS(SELECT 1 FROM team_members t WHERE t.project_id=other.id AND t.user_id=u.id))
 AND (u.id=(SELECT student_id FROM projects WHERE id=?) OR EXISTS(SELECT 1 FROM team_members t WHERE t.project_id=? AND t.user_id=u.id)))`, projectID, projectID, projectID).Scan(&conflicts)
	if err != nil {
		return err
	}
	if conflicts > 0 {
		return errors.New("A team member already has an approved project. Resolve team membership before accepting.")
	}
	c, err := readCapacity(tx, teacherID)
	if err != nil {
		return err
	}
	var size int
	err = tx.QueryRow(`SELECT COUNT(*) FROM users WHERE id=(SELECT student_id FROM projects WHERE id=?) OR id IN(SELECT user_id FROM team_members WHERE project_id=?)`, projectID, projectID).Scan(&size)
	if err != nil {
		return err
	}
	if c.UsedTeams+1 > c.MaxTeams || c.UsedStudents+size > c.MaxStudents {
		return errors.New("Not enough supervision capacity for this team. Update your capacity in Profile or decline the request.")
	}
	return nil
}
func checkJoinCapacity(tx *Transaction, projectID, studentID int64) error {
	var tid int64
	var status string
	if err := tx.QueryRow("SELECT COALESCE(teacher_id,0),status FROM projects WHERE id=?", projectID).Scan(&tid, &status); err != nil {
		return err
	}
	if status != "approved" || tid == 0 {
		return nil
	}
	c, err := readCapacity(tx, tid)
	if err != nil {
		return err
	}
	var exists int
	err = tx.QueryRow(`SELECT COUNT(*) FROM projects p WHERE p.teacher_id=? AND p.status='approved' AND (p.student_id=? OR EXISTS(SELECT 1 FROM team_members t WHERE t.project_id=p.id AND t.user_id=?))`, tid, studentID, studentID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 && c.UsedStudents+1 > c.MaxStudents {
		return errors.New("The supervisor's student capacity is full. Ask the supervisor to increase capacity before joining.")
	}
	return nil
}
func (a *App) studentPlacement(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if u.Role != "student" {
		fail(w, 403, "Students only.")
		return
	}
	var id int64
	var title, status, teacher string
	var size int
	e = a.db.QueryRow(`SELECT p.id,p.title,p.status,COALESCE(t.name,''),(SELECT COUNT(*) FROM users m WHERE m.id=p.student_id OR m.id IN(SELECT user_id FROM team_members WHERE project_id=p.id)) FROM projects p LEFT JOIN users t ON t.id=p.teacher_id WHERE p.student_id=? OR EXISTS(SELECT 1 FROM team_members m WHERE m.project_id=p.id AND m.user_id=?) ORDER BY CASE WHEN p.status='approved' THEN 0 WHEN EXISTS(SELECT 1 FROM team_members m WHERE m.project_id=p.id) THEN 1 ELSE 2 END,p.updated_at DESC,p.id DESC LIMIT 1`, u.ID, u.ID).Scan(&id, &title, &status, &teacher, &size)
	if e != nil && !errors.Is(e, sql.ErrNoRows) {
		fail(w, 500, "Could not load team status.")
		return
	}
	membership := "Looking for a team"
	if size > 1 {
		membership = "Team joined"
	} else if id > 0 {
		membership = "Project created — invite teammates"
	}
	supervision := "No supervisor assigned"
	if status == "approved" {
		supervision = "Supervisor assigned"
	} else if status == "submitted" {
		supervision = "Awaiting teacher approval"
	} else if status == "changes_requested" {
		supervision = "Teacher requested changes"
	} else if status == "rejected" {
		supervision = "Request declined"
	}
	w.Header().Set("Cache-Control", "no-store")
	reply(w, 200, map[string]any{"project_id": id, "title": title, "membership": membership, "supervision": supervision, "teacher": teacher, "team_size": size, "status": status})
}
