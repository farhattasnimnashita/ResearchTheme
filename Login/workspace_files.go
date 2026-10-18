package main

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) workspaceFiles(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	pid := projectID(r)
	if !a.projectAccess(u, pid) {
		fail(w, 403, "This project is not assigned to you.")
		return
	}
	if r.Method == "POST" {
		r.Body = http.MaxBytesReader(w, r.Body, 9<<20)
		if e = r.ParseMultipartForm(9 << 20); e != nil {
			fail(w, 400, "Choose one file up to 8 MB.")
			return
		}
		defer r.MultipartForm.RemoveAll()
		f, header, e := r.FormFile("file")
		if e != nil {
			fail(w, 400, "Choose a file.")
			return
		}
		defer f.Close()
		name := filepath.Base(strings.ReplaceAll(header.Filename, "\\", "/"))
		ext := strings.ToLower(filepath.Ext(name))
		allowed := map[string]bool{".pdf": true, ".docx": true, ".pptx": true, ".xlsx": true, ".zip": true, ".txt": true, ".md": true, ".csv": true, ".png": true, ".jpg": true, ".jpeg": true, ".go": true, ".py": true, ".cpp": true, ".c": true, ".js": true, ".html": true}
		if !allowed[ext] || len(name) > 180 || strings.ContainsAny(name, "\r\n\x00") {
			fail(w, 400, "Use PDF, Office, ZIP, text, CSV, image or source-code files.")
			return
		}
		data, e := io.ReadAll(io.LimitReader(f, (8<<20)+1))
		if e != nil || len(data) == 0 || len(data) > 8<<20 {
			fail(w, 400, "File must be between 1 byte and 8 MB.")
			return
		}
		tx, e := a.db.Begin()
		if e != nil {
			fail(w, 500, "Could not save file.")
			return
		}
		defer tx.Rollback()
		var total int64
		if tx.QueryRow("SELECT COALESCE(SUM(length(data)),0) FROM workspace_files WHERE project_id=?", pid).Scan(&total) != nil {
			fail(w, 500, "Could not check storage.")
			return
		}
		if total+int64(len(data)) > 64<<20 {
			fail(w, 409, "This project has reached its 64 MB attachment limit. Use a repository or document link instead.")
			return
		}
		id, e := tx.InsertID("INSERT INTO workspace_files(project_id,author_id,name,media_type,data,created_at) VALUES(?,?,?,?,?,?)", pid, u.ID, name, "application/octet-stream", data, time.Now().Unix())
		if e != nil {
			fail(w, 500, "Could not upload file.")
			return
		}
		if tx.Commit() != nil {
			fail(w, 500, "Could not save file.")
			return
		}
		a.notifyProject(pid, u.ID, "submission", "File uploaded: "+name)
		reply(w, 201, map[string]any{"id": id, "link": fmt.Sprintf("/api/files/%d", id), "message": "File saved. Copy its link into a submission to request review. Uploads with the same name remain separate versions."})
		return
	}
	rows, e := a.db.Query("SELECT f.id,f.name,length(f.data),f.created_at,u.name FROM workspace_files f JOIN users u ON u.id=f.author_id WHERE f.project_id=? ORDER BY f.id DESC", pid)
	if e != nil {
		fail(w, 500, "Could not load files.")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, size, at int64
		var name, author string
		if rows.Scan(&id, &name, &size, &at, &author) != nil {
			fail(w, 500, "Could not read files.")
			return
		}
		out = append(out, map[string]any{"id": id, "name": name, "size": size, "created_at": at, "author": author, "link": fmt.Sprintf("/api/files/%d", id)})
	}
	if rows.Err() != nil {
		fail(w, 500, "Could not read files.")
		return
	}
	reply(w, 200, out)
}
func (a *App) downloadFile(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	var pid, at int64
	var name string
	id := projectID(r)
	if a.db.QueryRow("SELECT project_id,name,created_at FROM workspace_files WHERE id=?", id).Scan(&pid, &name, &at) != nil {
		fail(w, 404, "File not found.")
		return
	}
	if !a.projectAccess(u, pid) {
		fail(w, 403, "This file is private to its project.")
		return
	}
	var data []byte
	if a.db.QueryRow("SELECT data FROM workspace_files WHERE id=?", id).Scan(&data) != nil {
		fail(w, 500, "Could not load file.")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	http.ServeContent(w, r, name, time.Unix(at, 0), bytes.NewReader(data))
}
