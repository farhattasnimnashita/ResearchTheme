package main

import (
	"io"
	"net/http"
	"strings"
)

const maxProfilePictureSize = 5 << 20 // 5 MB

func (a *App) uploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	user, err := a.user(r)
	if err != nil {
		fail(w, http.StatusUnauthorized, "Please log in.")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxProfilePictureSize)

	if err := r.ParseMultipartForm(maxProfilePictureSize); err != nil {
		fail(w, http.StatusBadRequest, "Image is too large or invalid.")
		return
	}

	file, _, err := r.FormFile("picture")
	if err != nil {
		fail(w, http.StatusBadRequest, "Please select a picture.")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil || len(data) == 0 || len(data) > maxProfilePictureSize {
		fail(w, http.StatusBadRequest, "Invalid picture size.")
		return
	}

	contentType := http.DetectContentType(data)
	if contentType != "image/jpeg" &&
		contentType != "image/png" &&
		contentType != "image/webp" {
		fail(w, http.StatusBadRequest, "Only JPG, PNG or WEBP images are allowed.")
		return
	}

	_, err = a.db.Exec(
		"UPDATE users SET avatar_data = ?, avatar_type = ? WHERE id = ?",
		data, contentType, user.ID,
	)
	if err != nil {
		fail(w, http.StatusInternalServerError, "Could not save profile picture.")
		return
	}

	reply(w, http.StatusOK, map[string]string{
		"message": "Profile picture uploaded successfully.",
	})
}

func (a *App) getProfilePicture(w http.ResponseWriter, r *http.Request) {
	user, err := a.user(r)
	if err != nil {
		fail(w, http.StatusUnauthorized, "Please log in.")
		return
	}

	var data []byte
	var contentType string

	err = a.db.QueryRow(
		"SELECT avatar_data, avatar_type FROM users WHERE id = ?",
		user.ID,
	).Scan(&data, &contentType)

	if err != nil || len(data) == 0 || strings.TrimSpace(contentType) == "" {
		http.NotFound(w, r)
		return
	}

	if contentType != "image/jpeg" &&
		contentType != "image/png" &&
		contentType != "image/webp" {
		http.Error(w, "Invalid image type.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}