package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"dataset-tracker/internal/middleware"
)

// avatarPalette holds background and foreground hex colours for 8 hues.
var avatarPalette = [8][2]string{
	{"#e0e7ff", "#4338ca"}, // indigo
	{"#ffe4e6", "#be123c"}, // rose
	{"#fef3c7", "#b45309"}, // amber
	{"#d1fae5", "#065f46"}, // emerald
	{"#e0f2fe", "#0369a1"}, // sky
	{"#ede9fe", "#6d28d9"}, // violet
	{"#fce7f3", "#9d174d"}, // pink
	{"#ccfbf1", "#0f766e"}, // teal
}

func avatarColorIndex(name string) int {
	var sum int
	for _, r := range name {
		sum += int(r)
	}
	return sum % len(avatarPalette)
}

func initialsOf(name string) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) == 0 {
		return "?"
	}
	if len(runes) == 1 {
		return string(runes[0])
	}
	return string(runes[0:2])
}

func initialsSVG(name string) []byte {
	idx := avatarColorIndex(name)
	bg, fg := avatarPalette[idx][0], avatarPalette[idx][1]
	text := initialsOf(name)
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 40 40" width="40" height="40">`+
		`<rect width="40" height="40" fill="%s"/>`+
		`<text x="20" y="20" dy=".35em" text-anchor="middle"`+
		` font-family="system-ui,sans-serif" font-size="14" font-weight="700" fill="%s">%s</text>`+
		`</svg>`, bg, fg, text)
	return []byte(svg)
}

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
	if err == nil && len(data) > 0 {
		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(data)
		return
	}
	user, err := h.users.GetByUsername(username)
	name := username
	if err == nil && user != nil {
		name = user.DisplayedName()
	}
	svg := initialsSVG(name)
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(svg)
}
