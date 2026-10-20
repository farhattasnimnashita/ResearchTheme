package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestStorageWorkspaceLifecycle(t *testing.T) {
	a, owner, teacher, outsider, pid := workspaceFixture(t)
	var gets atomic.Int32
	var deletes atomic.Int32
	var broken atomic.Bool
	var saved []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("apikey") != "sb_secret_test" || r.Header.Get("Authorization") != "" {
			t.Error("incorrect secret headers")
		}
		switch r.Method {
		case "POST":
			if !strings.HasPrefix(r.URL.Path, "/storage/v1/object/research-files/projects/") {
				t.Error("wrong upload route")
			}
			if broken.Load() {
				http.Error(w, "unavailable", 503)
				return
			}
			saved, _ = io.ReadAll(r.Body)
			w.WriteHeader(200)
		case "GET":
			gets.Add(1)
			if !strings.HasPrefix(r.URL.Path, "/storage/v1/object/authenticated/research-files/projects/") {
				t.Error("wrong download route")
			}
			w.Write(saved)
		case "DELETE":
			deletes.Add(1)
			w.Write([]byte(`[]`))
		default:
			t.Error("unexpected method")
		}
	}))
	defer srv.Close()
	a.cfg.StorageMode = "supabase"
	a.cfg.StorageBucket = "research-files"
	a.cfg.StorageKey = "sb_secret_test"
	a.cfg.SupabaseURL = srv.URL
	upload := func(cookie *http.Cookie) *httptest.ResponseRecorder {
		var b bytes.Buffer
		m := multipart.NewWriter(&b)
		f, _ := m.CreateFormFile("file", "report.txt")
		f.Write([]byte("Research results"))
		m.Close()
		r := httptest.NewRequest("POST", "/api/workspace/"+itoa(pid)+"/files", &b)
		r.Header.Set("Content-Type", m.FormDataContentType())
		r.Header.Set("Origin", a.cfg.URL)
		r.Header.Set("X-Requested-With", "ResearchBridge")
		r.AddCookie(cookie)
		w := httptest.NewRecorder()
		a.routes().ServeHTTP(w, r)
		return w
	}
	expect(t, upload(outsider), 403)
	w := upload(owner)
	expect(t, w, 201)
	var result struct {
		ID   int64
		Link string
	}
	json.Unmarshal(w.Body.Bytes(), &result)
	var size, blob int64
	if err := a.db.QueryRow("SELECT s.byte_size,length(f.data) FROM workspace_storage s JOIN workspace_files f ON f.id=s.file_id WHERE f.id=?", result.ID).Scan(&size, &blob); err != nil || size != 16 || blob != 0 {
		t.Fatalf("metadata: %v %d %d", err, size, blob)
	}
	expect(t, call(t, a, "GET", result.Link, nil, outsider), 403)
	expect(t, call(t, a, "GET", result.Link, nil, nil), 401)
	if gets.Load() != 0 {
		t.Fatal("unauthorized storage request")
	}
	w = call(t, a, "GET", result.Link, nil, teacher)
	expect(t, w, 200)
	if w.Body.String() != "Research results" {
		t.Fatal("download mismatch")
	}
	w = call(t, a, "GET", "/api/workspace/"+itoa(pid)+"/files", nil, owner)
	expect(t, w, 200)
	if !strings.Contains(w.Body.String(), `"size":16`) {
		t.Fatal("size missing")
	}
	broken.Store(true)
	expect(t, upload(owner), 502)
	var count int
	a.db.QueryRow("SELECT COUNT(*) FROM workspace_files WHERE project_id=?", pid).Scan(&count)
	if count != 1 || deletes.Load() != 1 {
		t.Fatalf("failure cleanup: %d %d", count, deletes.Load())
	}
	broken.Store(false)
	a.db.Exec("UPDATE workspace_storage SET byte_size=? WHERE file_id=?", 64<<20, result.ID)
	expect(t, upload(owner), 409)
}

func TestStorageConfiguration(t *testing.T) {
	c := Config{StorageMode: "supabase", StorageBucket: "research-files", StorageKey: "sb_secret_test", SupabaseURL: "https://project.supabase.co"}
	if err := validateStorageConfig(c); err != nil {
		t.Fatal(err)
	}
	c.StorageKey = "sb_publishable_example"
	if validateStorageConfig(c) == nil {
		t.Fatal("accepted public credential")
	}
	c.StorageKey = "sb_secret_test"
	c.SupabaseURL = "http://project.supabase.co"
	if validateStorageConfig(c) == nil {
		t.Fatal("accepted insecure origin")
	}
	if _, err := storageObjectPath("research-files", "projects/../secret"); err == nil {
		t.Fatal("accepted traversal")
	}
}
