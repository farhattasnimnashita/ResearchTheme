package main

import (
	"net/http"
	"strings"
	"time"
)

type literaturePaper struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Authors     string `json:"authors"`
	Year        int    `json:"year"`
	Venue       string `json:"venue"`
	DOI         string `json:"doi"`
	PaperURL    string `json:"paper_url"`
	Objective   string `json:"objective"`
	Methodology string `json:"methodology"`
	Dataset     string `json:"dataset"`
	KeyFindings string `json:"key_findings"`
	Limitations string `json:"limitations"`
	ResearchGap string `json:"research_gap"`
	Relevance   string `json:"relevance"`
	CreatedAt   int64  `json:"created_at"`
}

func (a *App) literaturePapers(w http.ResponseWriter, r *http.Request) {
	// User login koreche ki na check kora.
	u, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}

	// Ei feature shudhu student use korte parbe.
	if u.Role != "student" {
		fail(w, 403, "Only students can manage literature review papers.")
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Logged-in student-er paper-gulo load kora.
		rows, err := a.db.Query(`
			SELECT id, title, authors, publication_year,
			       venue, doi, paper_url, objective, methodology,
			       dataset, key_findings, limitations, research_gap,
			       relevance, created_at
			FROM literature_papers
			WHERE student_id = ?
			ORDER BY created_at DESC, id DESC
		`, u.ID)

		if err != nil {
			fail(w, 500, "Could not load research papers.")
			return
		}
		defer rows.Close()

		papers := make([]literaturePaper, 0)

		for rows.Next() {
			var p literaturePaper

			err := rows.Scan(
				&p.ID,
				&p.Title,
				&p.Authors,
				&p.Year,
				&p.Venue,
				&p.DOI,
				&p.PaperURL,
				&p.Objective,
				&p.Methodology,
				&p.Dataset,
				&p.KeyFindings,
				&p.Limitations,
				&p.ResearchGap,
				&p.Relevance,
				&p.CreatedAt,
			)

			if err != nil {
				fail(w, 500, "Could not read research papers.")
				return
			}

			papers = append(papers, p)
		}

		if err := rows.Err(); err != nil {
			fail(w, 500, "Could not load research papers.")
			return
		}

		reply(w, 200, papers)

	case http.MethodPost:
		// Form theke asha data.
		var in struct {
			Title       string `json:"title"`
			Authors     string `json:"authors"`
			Year        int    `json:"year"`
			Venue       string `json:"venue"`
			DOI         string `json:"doi"`
			PaperURL    string `json:"paper_url"`
			Objective   string `json:"objective"`
			Methodology string `json:"methodology"`
			Dataset     string `json:"dataset"`
			KeyFindings string `json:"key_findings"`
			Limitations string `json:"limitations"`
			ResearchGap string `json:"research_gap"`
			Relevance   string `json:"relevance"`
		}

		if !decode(w, r, &in) {
			return
		}

		// Extra space remove kora.
		in.Title = strings.TrimSpace(in.Title)
		in.Authors = strings.TrimSpace(in.Authors)
		in.Venue = strings.TrimSpace(in.Venue)
		in.DOI = strings.TrimSpace(in.DOI)
		in.PaperURL = strings.TrimSpace(in.PaperURL)
		in.Objective = strings.TrimSpace(in.Objective)
		in.Methodology = strings.TrimSpace(in.Methodology)
		in.Dataset = strings.TrimSpace(in.Dataset)
		in.KeyFindings = strings.TrimSpace(in.KeyFindings)
		in.Limitations = strings.TrimSpace(in.Limitations)
		in.ResearchGap = strings.TrimSpace(in.ResearchGap)
		in.Relevance = strings.TrimSpace(in.Relevance)

		// Required field validation.
		if in.Title == "" ||
			in.Authors == "" ||
			in.Objective == "" ||
			in.Methodology == "" ||
			in.KeyFindings == "" {
			fail(w, 400, "Please fill in all required fields.")
			return
		}

		// Field length ebong publication year check.
		if len(in.Title) > 500 ||
			len(in.Authors) > 500 ||
			(in.Year != 0 && (in.Year < 1900 || in.Year > 2100)) ||
			len(in.Venue) > 500 ||
			len(in.DOI) > 300 ||
			len(in.PaperURL) > 2000 ||
			len(in.Objective) > 5000 ||
			len(in.Methodology) > 5000 ||
			len(in.Dataset) > 3000 ||
			len(in.KeyFindings) > 5000 ||
			len(in.Limitations) > 3000 ||
			len(in.ResearchGap) > 3000 ||
			len(in.Relevance) > 3000 {
			fail(w, 400, "One or more fields are too long or invalid.")
			return
		}

		// Paper database-e save kora.
		id, err := a.db.InsertID(`
			INSERT INTO literature_papers (
				student_id, title, authors, publication_year,
				venue, doi, paper_url, objective, methodology,
				dataset, key_findings, limitations, research_gap,
				relevance, created_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
			u.ID,
			in.Title,
			in.Authors,
			in.Year,
			in.Venue,
			in.DOI,
			in.PaperURL,
			in.Objective,
			in.Methodology,
			in.Dataset,
			in.KeyFindings,
			in.Limitations,
			in.ResearchGap,
			in.Relevance,
			time.Now().Unix(),
		)

		if err != nil {
			fail(w, 500, "Could not save the research paper.")
			return
		}

		reply(w, 201, map[string]any{
			"id":      id,
			"message": "Research paper saved successfully.",
		})

	default:
		w.Header().Set("Allow", "GET, POST")
		fail(w, 405, "Method not allowed.")
	}
}
