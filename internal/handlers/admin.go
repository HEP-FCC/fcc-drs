package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"dataset-tracker/internal/middleware"
	"dataset-tracker/internal/models"
)

// adminUsersPageData fetches users (with group memberships) and all groups.
func (h *Handler) adminUsersPageData() (PageData, error) {
	users, err := h.users.GetAll()
	if err != nil {
		return PageData{}, err
	}
	groups, err := h.groups.GetAll()
	if err != nil {
		return PageData{}, err
	}
	memberships, err := h.groups.GetMembershipsByUser()
	if err != nil {
		return PageData{}, err
	}
	for _, u := range users {
		u.GroupMemberships = memberships[u.ID]
	}
	return PageData{Title: "User Management", Users: users, Groups: groups}, nil
}


func (h *Handler) adminGroupsPageData() (PageData, error) {
	groups, err := h.groups.GetAll()
	if err != nil {
		return PageData{}, err
	}
	coordinators, err := h.users.GetCoordinators()
	if err != nil {
		return PageData{}, err
	}
	return PageData{Title: "Coordinator Groups", Groups: groups, Coordinators: coordinators}, nil
}

func (h *Handler) AdminGroups(w http.ResponseWriter, r *http.Request) {
	data, err := h.adminGroupsPageData()
	if err != nil {
		slog.Error("admin list groups", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.renderPage(w, r, "admin_groups", data)
}

func (h *Handler) AdminCreateGroup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Error(w, "group name is required", 400)
		return
	}
	description := strings.TrimSpace(r.FormValue("description"))
	if err := h.groups.Create(name, description); err != nil {
		slog.Error("create group", "name", name, "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	data, err := h.adminGroupsPageData()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.renderPartial(w, r, "groups_grid", data)
}

func (h *Handler) AdminDeleteGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	if err := h.groups.Delete(id); err != nil {
		slog.Error("delete group", "id", id, "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	data, err := h.adminGroupsPageData()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.renderPartial(w, r, "groups_grid", data)
}

func (h *Handler) AdminAddGroupMember(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}
	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil || userID <= 0 {
		http.Error(w, "invalid user_id", 400)
		return
	}
	if err := h.groups.AddMember(groupID, userID); err != nil {
		slog.Error("add group member", "group_id", groupID, "user_id", userID, "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	group, err := h.groups.GetByID(groupID)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	coordinators, _ := h.users.GetCoordinators()
	h.renderPartial(w, r, "group_card", PageData{Group: group, Coordinators: coordinators})
}

func (h *Handler) AdminRemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	userID, err := strconv.Atoi(r.PathValue("user_id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	currentUser := middleware.GetUser(r)
	if currentUser != nil && currentUser.ID == userID {
		http.Error(w, "cannot remove yourself from a group", http.StatusBadRequest)
		return
	}
	if err := h.groups.RemoveMember(groupID, userID); err != nil {
		slog.Error("remove group member", "group_id", groupID, "user_id", userID, "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	group, err := h.groups.GetByID(groupID)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	coordinators, _ := h.users.GetCoordinators()
	h.renderPartial(w, r, "group_card", PageData{Group: group, Coordinators: coordinators})
}

func (h *Handler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	data, err := h.adminUsersPageData()
	if err != nil {
		slog.Error("admin list users", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.renderPage(w, r, "admin_users", data)
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

	data, err := h.adminUsersPageData()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.renderPartial(w, r, "user_table", data)
}
