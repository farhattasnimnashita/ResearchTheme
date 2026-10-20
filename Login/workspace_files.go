package main

import (
	"bytes"
	"database/sql"
	"errors"
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

		allowed := map[string]bool{
			".pdf": true, ".docx": true, ".pptx": true,
			".xlsx": true, ".zip": true, ".txt": true,
			".md": true, ".csv": true, ".png": true,
			".jpg": true, ".jpeg": true, ".go": true,
			".py": true, ".cpp": true, ".c": true,
			".js": true, ".html": true,
		}

		if !allowed[ext] || len(name) > 180 ||
			strings.ContainsAny(name, "\r\n\x00") {
			fail(w, 400, "Use PDF, Office, ZIP, text, CSV, image or source-code files.")
			return
		}

		data, e := io.ReadAll(io.LimitReader(f, (8<<20)+1))
		if e != nil || len(data) == 0 || len(data) > 8<<20 {
			fail(w, 400, "File must be between 1 byte and 8 MB.")
			return
		}

		useStorage := a.cfg.StorageMode == "supabase"

		tx, e := a.db.Begin()
		if e != nil {
			fail(w, 500, "Could not save file.")
			return
		}
		defer tx.Rollback()

		var total int64
		e = tx.QueryRow(`
			SELECT COALESCE(SUM(COALESCE(s.byte_size, length(f.data))), 0)
			FROM workspace_files f
			LEFT JOIN workspace_storage s ON s.file_id = f.id
			WHERE f.project_id = ?
		`, pid).Scan(&total)

		if e != nil {
			fail(w, 500, "Could not check storage.")
			return
		}

		if total+int64(len(data)) > 64<<20 {
			fail(w, 409, "This project has reached its 64 MB attachment limit. Use a repository or document link instead.")
			return
		}

		// Supabase mode stores the actual bytes remotely.
		// Local mode continues to store the bytes in the database.
		dbData := data
		if useStorage {
			dbData = []byte{}
		}

		id, e := tx.InsertID(`
			INSERT INTO workspace_files
				(project_id, author_id, name, media_type, data, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, pid, u.ID, name, "application/octet-stream", dbData, time.Now().Unix())

		if e != nil {
			fail(w, 500, "Could not upload file.")
			return
		}

		if useStorage {
			objectKey := fmt.Sprintf("projects/%d/%d/%s", pid, id, name)

			// Upload the bytes to Supabase Storage.
			if e = a.storagePut(r.Context(), a.cfg.StorageBucket, objectKey, data); e != nil {
				// Best-effort cleanup in case the remote service partially
				// created the object before reporting the failure.
				_ = a.storageDelete(r.Context(), a.cfg.StorageBucket, objectKey)
				fail(w, 502, "Could not upload file to storage.")
				return
			}

			// Store where the remote bytes live and how large they are.
			_, e = tx.Exec(`
				INSERT INTO workspace_storage(file_id, bucket, object_key, byte_size)
				VALUES (?, ?, ?, ?)
			`, id, a.cfg.StorageBucket, objectKey, len(data))

			if e != nil {
				_ = a.storageDelete(r.Context(), a.cfg.StorageBucket, objectKey)
				fail(w, 500, "Could not save file metadata.")
				return
			}
		}

		if e = tx.Commit(); e != nil {
			if useStorage {
				objectKey := fmt.Sprintf("projects/%d/%d/%s", pid, id, name)
				_ = a.storageDelete(r.Context(), a.cfg.StorageBucket, objectKey)
			}
			fail(w, 500, "Could not save file.")
			return
		}

		a.notifyProject(pid, u.ID, "submission", "File uploaded: "+name)

		reply(w, 201, map[string]any{
			"id":      id,
			"link":    fmt.Sprintf("/api/files/%d", id),
			"message": "File saved. Copy its link into a submission to request review. Uploads with the same name remain separate versions.",
		})
		return
	}

	rows, e := a.db.Query(`
		SELECT f.id, f.name, COALESCE(s.byte_size, length(f.data)),
		       f.created_at, u.name
		FROM workspace_files f
		JOIN users u ON u.id = f.author_id
		LEFT JOIN workspace_storage s ON s.file_id = f.id
		WHERE f.project_id = ?
		ORDER BY f.id DESC
	`, pid)

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

		out = append(out, map[string]any{
			"id":         id,
			"name":       name,
			"size":       size,
			"created_at": at,
			"author":     author,
			"link":       fmt.Sprintf("/api/files/%d", id),
		})
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

	id := projectID(r)

	var pid, at int64
	var name string

	if e = a.db.QueryRow(
		"SELECT project_id, name, created_at FROM workspace_files WHERE id = ?",
		id,
	).Scan(&pid, &name, &at); e != nil {
		fail(w, 404, "File not found.")
		return
	}

	if !a.projectAccess(u, pid) {
		fail(w, 403, "This file is private to its project.")
		return
	}

	var data []byte
	var bucket, objectKey string
	var size int64

	// Look for remote-storage metadata first.
	e = a.db.QueryRow(`
		SELECT bucket, object_key, byte_size
		FROM workspace_storage
		WHERE file_id = ?
	`, id).Scan(&bucket, &objectKey, &size)

	if e == nil {
		// This file lives in Supabase Storage.
		data, e = a.storageGet(r.Context(), bucket, objectKey, size)
		if e != nil {
			fail(w, 502, "Could not download file from storage.")
			return
		}
	} else if errors.Is(e, sql.ErrNoRows) {
		// Backward-compatible local database file.
		if e = a.db.QueryRow(
			"SELECT data FROM workspace_files WHERE id = ?", id,
		).Scan(&data); e != nil {
			fail(w, 500, "Could not load file.")
			return
		}
	} else {
		fail(w, 500, "Could not load file metadata.")
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set(
		"Content-Disposition",
		mime.FormatMediaType("attachment", map[string]string{"filename": name}),
	)

	http.ServeContent(w, r, name, time.Unix(at, 0), bytes.NewReader(data))
}
