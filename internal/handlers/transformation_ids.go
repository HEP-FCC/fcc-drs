package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"fcc-drs/internal/middleware"
	"fcc-drs/internal/models"
)

func transformationIDBody(verb string, label string, transID int) string {
	body := "Transformation ID " + verb + ": " + strconv.Itoa(transID)
	if label != "" {
		body += " (" + label + ")"
	}
	return body
}

func (h *Handler) AddTransformationID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	req, err := h.requests.GetByID(id)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	if req.CampaignID == 0 {
		http.Error(w, "request must be assigned to a campaign first", http.StatusBadRequest)
		return
	}
	user := middleware.GetUser(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	transID, err := strconv.Atoi(strings.TrimSpace(r.FormValue("transformation_id")))
	if err != nil || transID <= 0 {
		http.Error(w, "transformation ID must be a positive number", 400)
		return
	}
	label := strings.TrimSpace(r.FormValue("label"))

	if err := h.transformationIDs.Add(id, label, transID, user.ID); err != nil {
		slog.Error("add transformation id", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.updates.Add(id, user.ID, models.UpdateTransformationID, transformationIDBody("added", label, transID))

	transIDs, _ := h.transformationIDs.GetByRequestID(id)
	h.renderPartial(w, r, "transformation_ids", PageData{Request: req, TransformationIDs: transIDs})
}

func (h *Handler) DeleteTransformationID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	transRowID, err := strconv.Atoi(r.PathValue("trans_id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	entry, err := h.transformationIDs.GetByID(transRowID)
	if err != nil || entry.RequestID != id {
		http.Error(w, "Not Found", 404)
		return
	}
	req, err := h.requests.GetByID(id)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	if req.CampaignID == 0 {
		http.Error(w, "request must be assigned to a campaign first", http.StatusBadRequest)
		return
	}
	user := middleware.GetUser(r)
	if err := h.transformationIDs.Delete(transRowID); err != nil {
		slog.Error("delete transformation id", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.updates.Add(id, user.ID, models.UpdateTransformationID, transformationIDBody("removed", entry.Label, entry.TransformationID))

	transIDs, _ := h.transformationIDs.GetByRequestID(id)
	h.renderPartial(w, r, "transformation_ids", PageData{Request: req, TransformationIDs: transIDs})
}
