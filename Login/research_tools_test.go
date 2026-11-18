package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"testing"
	"time"
)

// loginUser returns an authentication cookie, not a User value. Tests that
// need the database user ID should resolve it by the account email.
func researchToolTestUserID(t *testing.T, a *App, email string) int64 {
	t.Helper()
	var id int64
	if err := a.db.QueryRow("SELECT id FROM users WHERE email=?", email).Scan(&id); err != nil {
		t.Fatalf("could not load test user ID for %s: %v", email, err)
	}
	return id
}

func TestResearchToolRecordsAreProjectScopedAndReusable(t *testing.T) {
	a, links := setup(t)
	register(t, a, "tools-student@example.com", "student")
	verifyUser(t, a, links, "tools-student@example.com")
	student := loginUser(t, a, "tools-student@example.com", "student")
	projectID := makeProject(t, a, student)

	payload := map[string]any{
		"project_id": projectID,
		"fields": map[string]string{
			"title":              "Dataset access delay",
			"entry_type":         "Risk",
			"description":        "The selected dataset may not be released on time.",
			"likelihood":         "Medium",
			"impact":             "High",
			"mitigation":         "Identify a permitted alternative dataset.",
			"decision_rationale": "",
			"owner":              "Project owner",
			"status":             "Monitoring",
			"review_date":        "",
		},
	}
	expect(t, call(t, a, "POST", "/api/research-tools/risk-decision-log", payload, student), 201)
	w := call(t, a, "GET", "/api/research-tools/risk-decision-log?project_id="+itoa(projectID), nil, student)
	expect(t, w, 200)
	var records []researchToolRecord
	if err := json.Unmarshal(w.Body.Bytes(), &records); err != nil || len(records) != 1 {
		t.Fatalf("expected one saved risk record, got %s (%v)", w.Body.String(), err)
	}
	if records[0].Title != "Dataset access delay" || records[0].Fields["impact"] != "High" {
		t.Fatalf("record fields were not preserved: %#v", records[0])
	}

	// Literature Review remains the source of bibliographic identity. Citation
	// Manager stores only formatting/verification details for that existing row.
	paper := map[string]any{
		"title": "Bengali NLP Study", "authors": "A. Researcher", "year": 2024,
		"venue": "Research Journal", "doi": "10.1234/example", "paper_url": "https://example.org/paper",
		"objective": "Study Bengali text classification.", "methodology": "Baseline comparison.",
		"dataset": "Public corpus", "key_findings": "The baseline was evaluated.",
		"limitations": "Limited domain coverage.", "research_gap": "More domain testing.", "relevance": "Related work.",
	}
	expect(t, call(t, a, "POST", "/api/literature-papers", paper, student), 201)
	papers := call(t, a, "GET", "/api/literature-papers", nil, student)
	expect(t, papers, 200)
	var paperRows []literaturePaper
	if err := json.Unmarshal(papers.Body.Bytes(), &paperRows); err != nil || len(paperRows) != 1 {
		t.Fatalf("literature paper was not saved: %s (%v)", papers.Body.String(), err)
	}
	citation := map[string]any{
		"project_id": projectID,
		"fields": map[string]string{
			"paper_id": itoa(paperRows[0].ID), "source_type": "Journal article",
			"volume_issue_pages": "Vol. 1, Issue 2, pp. 1-10", "accessed_date": "",
			"citation_style": "IEEE", "bibtex": "@article{verified2024}", "notes": "Supports the related-work discussion.",
		},
	}
	expect(t, call(t, a, "POST", "/api/research-tools/citation-manager", citation, student), 201)
	expect(t, call(t, a, "POST", "/api/research-tools/citation-manager", citation, student), 409)
	citations := call(t, a, "GET", "/api/research-tools/citation-manager?project_id="+itoa(projectID), nil, student)
	expect(t, citations, 200)
	var citationRows []researchToolRecord
	if err := json.Unmarshal(citations.Body.Bytes(), &citationRows); err != nil || len(citationRows) != 1 {
		t.Fatalf("citation metadata was not saved: %s (%v)", citations.Body.String(), err)
	}
	if citationRows[0].Title != "Bengali NLP Study" {
		t.Fatalf("citation manager did not reuse Literature Review title: %#v", citationRows[0])
	}
	sources := call(t, a, "GET", "/api/research-tools/citation-sources?project_id="+itoa(projectID), nil, student)
	expect(t, sources, 200)
	var sourceRows []literaturePaper
	if err := json.Unmarshal(sources.Body.Bytes(), &sourceRows); err != nil || len(sourceRows) != 1 || sourceRows[0].Title != "Bengali NLP Study" {
		t.Fatalf("project citation source list should expose the team's saved paper: %s (%v)", sources.Body.String(), err)
	}
}

func TestResearchToolValidationAndProjectAccess(t *testing.T) {
	a, links := setup(t)
	register(t, a, "tools-owner@example.com", "student")
	verifyUser(t, a, links, "tools-owner@example.com")
	owner := loginUser(t, a, "tools-owner@example.com", "student")
	projectID := makeProject(t, a, owner)

	invalid := map[string]any{
		"project_id": projectID,
		"fields": map[string]string{
			"title": "Unverified dataset", "description": "An example dataset record.",
			"source_url": "javascript:alert(1)", "creator_owner": "", "version": "1",
			"format_size": "CSV", "license": "Unknown", "consent_basis": "",
			"sensitivity": "Public", "access_group": "Team", "retention_date": "",
			"quality_notes": "",
		},
	}
	expect(t, call(t, a, "POST", "/api/research-tools/dataset-registry", invalid, owner), 400)

	register(t, a, "tools-outsider@example.com", "student")
	verifyUser(t, a, links, "tools-outsider@example.com")
	outsider := loginUser(t, a, "tools-outsider@example.com", "student")
	get := call(t, a, "GET", "/api/research-tools/risk-decision-log?project_id="+itoa(projectID), nil, outsider)
	expect(t, get, 403)
}

func TestProjectExportBuildsZipFromSavedPlan(t *testing.T) {
	a, links := setup(t)
	register(t, a, "export-student@example.com", "student")
	verifyUser(t, a, links, "export-student@example.com")
	student := loginUser(t, a, "export-student@example.com", "student")
	studentID := researchToolTestUserID(t, a, "export-student@example.com")
	projectID := makeProject(t, a, student)

	plan := map[string]any{
		"project_id": projectID,
		"fields": map[string]string{
			"title": "Supervisor handover", "export_format": "ZIP metadata bundle",
			"include_proposal": "true", "include_literature": "false", "include_code": "true",
			"include_dataset_metadata": "false", "include_experiments": "false", "include_decisions": "false",
			"recipient": "Supervisor", "handover_date": "", "handover_status": "Preparing",
			"rights_notes": "Check source-code permissions before handover.", "notes": "",
		},
	}
	created := call(t, a, "POST", "/api/research-tools/project-export", plan, student)
	expect(t, created, 201)
	var response struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &response); err != nil || response.ID <= 0 {
		t.Fatalf("missing export plan ID: %s (%v)", created.Body.String(), err)
	}
	w := call(t, a, "GET", "/api/research-tools/export?plan_id="+itoa(response.ID), nil, student)
	expect(t, w, 200)
	if w.Header().Get("Content-Type") != "application/zip" {
		t.Fatalf("expected ZIP content type, got %s", w.Header().Get("Content-Type"))
	}
	zr, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if err != nil {
		t.Fatalf("export is not a valid ZIP: %v", err)
	}
	foundManifest, foundProposal := false, false
	for _, file := range zr.File {
		if file.Name == "manifest.json" {
			foundManifest = true
		}
		if file.Name == "proposal.json" {
			foundProposal = true
		}
		if file.Name == "manifest.json" {
			r, openErr := file.Open()
			if openErr != nil {
				t.Fatal(openErr)
			}
			manifest, readErr := io.ReadAll(r)
			r.Close()
			if readErr != nil || !bytes.Contains(manifest, []byte(`"code_included": false`)) {
				t.Fatalf("manifest should not claim source code was bundled: %s (%v)", manifest, readErr)
			}
		}
	}
	if !foundManifest || !foundProposal {
		t.Fatalf("ZIP missing expected selected files: manifest=%v proposal=%v", foundManifest, foundProposal)
	}

	// When source code has been uploaded to the project's Files workspace, the
	// same saved plan should include it in the ZIP and report that accurately.
	fileID, err := a.db.InsertID(`INSERT INTO workspace_files(project_id,author_id,name,media_type,data,created_at) VALUES(?,?,?,?,?,?)`,
		projectID, studentID, "baseline.py", "text/plain", []byte("print('research')"), time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	w = call(t, a, "GET", "/api/research-tools/export?plan_id="+itoa(response.ID), nil, student)
	expect(t, w, 200)
	zr, err = zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	if err != nil {
		t.Fatalf("export with source code is not a valid ZIP: %v", err)
	}
	foundCode := false
	for _, file := range zr.File {
		if file.Name == "code/"+itoa(fileID)+"-baseline.py" {
			foundCode = true
		}
		if file.Name == "manifest.json" {
			r, openErr := file.Open()
			if openErr != nil {
				t.Fatal(openErr)
			}
			manifest, readErr := io.ReadAll(r)
			r.Close()
			if readErr != nil || !bytes.Contains(manifest, []byte(`"code_included": true`)) {
				t.Fatalf("manifest should report source code included: %s (%v)", manifest, readErr)
			}
		}
	}
	if !foundCode {
		t.Fatalf("ZIP did not include the uploaded source-code file %d", fileID)
	}
}

func TestResearchQuestionAndTaskDependenciesRejectCycles(t *testing.T) {
	a, links := setup(t)
	register(t, a, "cycle-student@example.com", "student")
	verifyUser(t, a, links, "cycle-student@example.com")
	student := loginUser(t, a, "cycle-student@example.com", "student")
	studentID := researchToolTestUserID(t, a, "cycle-student@example.com")
	projectID := makeProject(t, a, student)
	now := time.Now().Unix()

	questionA, err := a.db.InsertID(`INSERT INTO research_tool_records(project_id,feature_key,title,fields_json,created_by,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`,
		projectID, "research-question-map", "Main question", `{"title":"Main question","parent_question":""}`, studentID, now, now)
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.db.InsertID(`INSERT INTO research_tool_records(project_id,feature_key,title,fields_json,created_by,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`,
		projectID, "research-question-map", "Sub question", `{"title":"Sub question","parent_question":"`+itoa(questionA)+`"}`, studentID, now, now)
	if err != nil {
		t.Fatal(err)
	}
	var questionB int64
	if err := a.db.QueryRow("SELECT id FROM research_tool_records WHERE project_id=? AND feature_key='research-question-map' AND title='Sub question'", projectID).Scan(&questionB); err != nil {
		t.Fatal(err)
	}
	if a.researchQuestionParentIsAcyclic(projectID, questionA, questionB) {
		t.Fatal("expected parent-child cycle to be rejected")
	}

	_, err = a.db.InsertID(`INSERT INTO research_tool_records(project_id,feature_key,title,fields_json,created_by,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`,
		projectID, "task-dependencies", "Task A depends on Task B", `{"task_id":"100","depends_on_id":"101"}`, studentID, now, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.validateTaskDependencyGraph(projectID, 0, 101, 100); err == nil {
		t.Fatal("expected a circular task dependency to be rejected")
	}
}
