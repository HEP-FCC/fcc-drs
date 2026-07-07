package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"dataset-tracker/internal/middleware"
	"dataset-tracker/internal/models"
)

func (h *Handler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.GetAll()
	if err != nil {
		slog.Error("admin list users", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.renderPage(w, r, "admin_users", PageData{Title: "User Management", Users: users})
}

func (h *Handler) AdminUpdateUserRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	role := models.Role(r.FormValue("role"))
	if role != models.RoleRequester && role != models.RoleCoordinator && role != models.RoleAdmin {
		http.Error(w, "invalid role", 400)
		return
	}

	// Prevent an admin from demoting themselves.
	currentUser := middleware.GetUser(r)
	if currentUser != nil && currentUser.ID == id && role != models.RoleAdmin {
		http.Error(w, "cannot change your own role", http.StatusBadRequest)
		return
	}

	if err := h.users.UpdateRole(id, role); err != nil {
		slog.Error("admin update role", "user_id", id, "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	slog.Info("role updated", "user_id", id, "role", role, "by", currentUser.Username)

	users, err := h.users.GetAll()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.renderPartial(w, r, "user_table", PageData{Users: users})
}
