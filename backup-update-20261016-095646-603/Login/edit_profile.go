package main

import (
	"net/http"
	"strings"
	"unicode/utf8"
)

func (a *App) editStudentProfile(w http.ResponseWriter, r *http.Request) {
	user, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}

	if user.Role != "student" {
		fail(w, 403, "Student account required.")
		return
	}

	var input struct {
		Name         string `json:"name"`
		Institution  string `json:"institution"`
		Department   string `json:"department"`
		AcademicYear int    `json:"academic_year"`
		Semester     int    `json:"semester"`
	}

	if !decode(w, r, &input) {
		return
	}

	input.Name = strings.TrimSpace(input.Name)

	if !utf8.ValidString(input.Name) ||
		utf8.RuneCountInString(input.Name) < 2 ||
		utf8.RuneCountInString(input.Name) > 100 {
		fail(w, 400, "Name must contain 2–100 characters.")
		return
	}

	profile := credentials{
		Role:         "student",
		Institution:  input.Institution,
		Department:   input.Department,
		AcademicYear: input.AcademicYear,
		Semester:     input.Semester,
	}

	if err := validateProfile(&profile); err != nil {
		fail(w, 400, err.Error())
		return
	}

	_, err = a.db.Exec(`
		UPDATE users
		SET name = ?, institution = ?, department = ?,
		    academic_year = ?, semester = ?
		WHERE id = ? AND role = 'student'
	`,
		input.Name,
		profile.Institution,
		profile.Department,
		profile.AcademicYear,
		profile.Semester,
		user.ID,
	)

	if err != nil {
		fail(w, 500, "Could not save profile.")
		return
	}

	reply(w, 200, map[string]string{
		"message": "Profile saved successfully.",
	})
}