package handlers

import (
	"log/slog"
	"net/http"

	"dataset-tracker/internal/models"
)

// CampaignGroup pairs a campaign with the requests currently assigned to it.
type CampaignGroup struct {
	Campaign *models.Campaign
	Requests []*models.DatasetRequest
}

func (h *Handler) CampaignsView(w http.ResponseWriter, r *http.Request) {
	campaigns, err := h.campaigns.GetAll()
	if err != nil {
		slog.Error("list campaigns", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	var open, closed []CampaignGroup
	for _, c := range campaigns {
		reqs, err := h.requests.GetByCampaign(c.ID)
		if err != nil {
			slog.Error("list campaign requests", "campaign_id", c.ID, "error", err)
			continue
		}
		group := CampaignGroup{Campaign: c, Requests: reqs}
		if c.Status == "closed" {
			closed = append(closed, group)
		} else {
			open = append(open, group)
		}
	}

	unassigned, err := h.requests.GetEligibleUnassigned()
	if err != nil {
		slog.Error("list eligible unassigned requests", "error", err)
	}

	tab := r.URL.Query().Get("tab")
	if tab != "closed" {
		tab = "open"
	}

	h.renderPage(w, r, "campaigns_view", PageData{
		Title:              "Campaigns",
		OpenCampaigns:      open,
		ClosedCampaigns:    closed,
		UnassignedRequests: unassigned,
		CampaignsTab:       tab,
	})
}
