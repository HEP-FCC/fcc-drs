package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"dataset-tracker/internal/middleware"
)

const maxAvatarSize = 2 << 20 // 2 MB

func (h *Handler) ShowProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	h.renderPage(w, r, "profile", PageData{Title: "Profile", CurrentUser: user})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)

	if err := r.ParseMultipartForm(maxAvatarSize + 512); err != nil {
		if err != http.ErrNotMultipart {
			http.Error(w, "Bad Request", 400)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", 400)
			return
		}
	}

	// Preferred name
	name := strings.TrimSpace(r.FormValue("preferred_name"))
	if err := h.users.UpdatePreferredName(user.ID, name); err != nil {
		slog.Error("update preferred name", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Avatar upload (optional)
	if r.MultipartForm != nil {
		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			if header.Size > maxAvatarSize {
				h.renderPage(w, r, "profile", PageData{
					Title: "Profile", CurrentUser: user,
					Error: "Avatar must be under 2 MB.",
				})
				return
			}
			data := make([]byte, header.Size)
			if _, err := file.Read(data); err != nil {
				http.Error(w, "Internal Server Error", 500)
				return
			}
			mime := http.DetectContentType(data)
			if !strings.HasPrefix(mime, "image/") {
				h.renderPage(w, r, "profile", PageData{
					Title: "Profile", CurrentUser: user,
					Error: "Only image files are accepted.",
				})
				return
			}
			if err := h.users.UpdateAvatar(user.ID, data, mime); err != nil {
				slog.Error("update avatar", "error", err)
				http.Error(w, "Internal Server Error", 500)
				return
			}
		}
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *Handler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if err := h.users.DeleteAvatar(user.ID); err != nil {
		slog.Error("delete avatar", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *Handler) ServeAvatar(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	data, mime, err := h.users.GetAvatar(username)
	if err != nil || len(data) == 0 {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}
