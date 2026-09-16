package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"dataset-tracker/internal/middleware"
	"dataset-tracker/internal/models"
)

func productionIDBody(verb string, label string, prodID int) string {
	body := "Production ID " + verb + ": " + strconv.Itoa(prodID)
	if label != "" {
		body += " (" + label + ")"
	}
	return body
}

func (h *Handler) AddProductionID(w http.ResponseWriter, r *http.Request) {
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

	prodID, err := strconv.Atoi(strings.TrimSpace(r.FormValue("production_id")))
	if err != nil || prodID <= 0 {
		http.Error(w, "production ID must be a positive number", 400)
		return
	}
	label := strings.TrimSpace(r.FormValue("label"))

	if err := h.productionIDs.Add(id, label, prodID, user.ID); err != nil {
		slog.Error("add production id", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.updates.Add(id, user.ID, models.UpdateProductionID, productionIDBody("added", label, prodID))

	prodIDs, _ := h.productionIDs.GetByRequestID(id)
	h.renderPartial(w, r, "production_ids", PageData{Request: req, ProductionIDs: prodIDs})
}

func (h *Handler) DeleteProductionID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	prodRowID, err := strconv.Atoi(r.PathValue("prod_id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	entry, err := h.productionIDs.GetByID(prodRowID)
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
	if err := h.productionIDs.Delete(prodRowID); err != nil {
		slog.Error("delete production id", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.updates.Add(id, user.ID, models.UpdateProductionID, productionIDBody("removed", entry.Label, entry.ProductionID))

	prodIDs, _ := h.productionIDs.GetByRequestID(id)
	h.renderPartial(w, r, "production_ids", PageData{Request: req, ProductionIDs: prodIDs})
}
