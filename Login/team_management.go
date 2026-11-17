package main

import (
	"net/http"
	"time"
)

func (a *App) teamStatus(w http.ResponseWriter, r *http.Request) {
	u, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}
	id := projectID(r)
	if id <= 0 || !a.projectAccess(u, id) {
		fail(w, 403, "You cannot view this team's members.")
		return
	}
	var ownerID int64
	if err = a.db.QueryRow("SELECT student_id FROM projects WHERE id=?", id).Scan(&ownerID); err != nil {
		fail(w, 404, "Project not found.")
		return
	}
	rows, err := a.db.Query(`
      SELECT r.id, r.assignee_id, target.name, target.email, r.status, r.updated_at,
             CASE WHEN r.assignee_id=? THEN 1 ELSE 0 END
      FROM workspace_records r
      JOIN users target ON target.id=r.assignee_id
      WHERE r.project_id=? AND r.kind='invitation'
      ORDER BY r.updated_at DESC, r.id DESC`, ownerID, id)
	if err != nil {
		fail(w, 500, "Could not load team history.")
		return
	}
	defer rows.Close()
	invitations := []map[string]any{}
	for rows.Next() {
		var rid, uid, updated int64
		var name, email, status string
		var isOwner int
		if rows.Scan(&rid, &uid, &name, &email, &status, &updated, &isOwner) != nil {
			fail(w, 500, "Could not read team history.")
			return
		}
		invitations = append(invitations, map[string]any{
			"id": rid, "user_id": uid, "name": name, "email": email, "status": status, "updated": updated,
		})
	}
	if rows.Err() != nil {
		fail(w, 500, "Could not read team history.")
		return
	}
	if err := rows.Close(); err != nil {
		fail(w, 500, "Could not finish loading team history.")
		return
	}
	members, err := a.db.Query(`
      SELECT u.id,u.name,u.email,CASE WHEN u.id=p.student_id THEN 1 ELSE 0 END
      FROM projects p JOIN users u ON u.id=p.student_id OR u.id IN
        (SELECT user_id FROM team_members WHERE project_id=p.id)
      WHERE p.id=? ORDER BY CASE WHEN u.id=p.student_id THEN 0 ELSE 1 END,lower(u.name)`, id)
	if err != nil {
		fail(w, 500, "Could not load team members.")
		return
	}
	defer members.Close()
	team := []map[string]any{}
	for members.Next() {
		var uid int64
		var name, email string
		var owner int
		if members.Scan(&uid, &name, &email, &owner) != nil {
			fail(w, 500, "Could not read team members.")
			return
		}
		team = append(team, map[string]any{"id": uid, "name": name, "email": email, "owner": owner == 1})
	}
	if members.Err() != nil {
		fail(w, 500, "Could not read team members.")
		return
	}
	var pending int
	_ = a.db.QueryRow("SELECT COUNT(*) FROM workspace_records WHERE project_id=? AND kind='invitation' AND status='pending'", id).Scan(&pending)
	reply(w, 200, map[string]any{"owner_id": ownerID, "members": team, "invitations": invitations, "team_limit": 3, "pending_count": pending, "can_manage": u.Role == "student" && u.ID == ownerID, "updated_at": time.Now().Unix()})
}
