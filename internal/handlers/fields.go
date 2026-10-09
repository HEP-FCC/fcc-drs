package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"fcc-drs/internal/middleware"
	"fcc-drs/internal/models"
)

// requestField describes a request field that can be edited on its own in the edit modal.
type requestField struct {
	Key         string
	Label       string
	Input       string // "text", "textarea", "select" or "date"
	Placeholder string
	Required    bool
	Options     []models.Option
}

var requestFields = map[string]requestField{
	"title":           {Key: "title", Label: "Title", Input: "text", Required: true, Placeholder: "e.g. Dijet events at √s=100 TeV for calorimeter studies"},
	"description":     {Key: "description", Label: "Description", Input: "textarea", Required: true, Placeholder: "Physics process, energy range, detector concept, selection criteria…"},
	"group":           {Key: "group", Label: "Group / Team", Input: "select", Required: true},
	"use_case":        {Key: "use_case", Label: "Use Case", Input: "select", Required: true, Options: models.UseCaseLabels},
	"dataset_type":    {Key: "dataset_type", Label: "Final Processing Stage", Input: "select", Required: true, Options: models.DatasetTypeLabels},
	"format":          {Key: "format", Label: "Format", Input: "select", Required: true, Options: models.FormatLabels},
	"detector":        {Key: "detector", Label: "Detector(s)", Input: "text", Required: true, Placeholder: "e.g. IDEA, ALLEGRO, ALFA — comma-separated if more than one"},
	"statistics":      {Key: "statistics", Label: "Statistics", Input: "text", Required: true, Placeholder: "e.g. 10M events"},
	"estimated_size":  {Key: "estimated_size", Label: "Estimated Size", Input: "text", Required: true, Placeholder: "e.g. 500 GB"},
	"target_campaign": {Key: "target_campaign", Label: "Target Campaign", Input: "text", Placeholder: "e.g. Spring2026"},
	"key4hep_stack":   {Key: "key4hep_stack", Label: "Key4hep Stack", Input: "text", Placeholder: "e.g. key4hep-2024-03-10"},
	"due_date":        {Key: "due_date", Label: "Due Date", Input: "date"},
	"notes":           {Key: "notes", Label: "Additional Notes", Input: "textarea", Placeholder: "Generator settings, beam energy, pile-up conditions, special requirements…"},
	"tags":            {Key: "tags", Label: "Tags", Input: "text", Placeholder: "e.g. fcc-hh, fcc-ee, higgs, top, ewk, bsm, flavour, llp"},
}

// FieldEdit is the data for the edit modal of a single field.
type FieldEdit struct {
	requestField
	Value string
}

// DetailItem is the data for one field in the request summary view.
type DetailItem struct {
	RequestID int
	Key       string
	Label     string
	Value     string
	Required  bool
	Missing   bool
	CanEdit   bool
}

// CanEdit reports whether the current user may edit the request being shown.
func (d PageData) CanEdit() bool {
	return d.Request != nil && canEdit(d.CurrentUser, d.Request)
}

// detailItem builds the summary-view data for a field; used from templates.
func detailItem(d PageData, key, value string) DetailItem {
	f := requestFields[key]
	return DetailItem{
		RequestID: d.Request.ID,
		Key:       key,
		Label:     f.Label,
		Value:     value,
		Required:  f.Required,
		Missing:   d.Request.IsMissing(key),
		CanEdit:   d.CanEdit(),
	}
}

// canEditGroup: requesters pick the group only while their request is a draft;
// afterwards coordinators reassign it through the assignment control.
func canEditGroup(user *models.User, req *models.DatasetRequest) bool {
	if user == nil {
		return false
	}
	return user.IsCoordinator() || (req.CreatedBy == user.ID && req.Status == models.StatusDraft)
}

func fieldValue(req *models.DatasetRequest, key string) string {
	switch key {
	case "title":
		return req.Title
	case "description":
		return req.Description
	case "group":
		if req.AssignedGroupID == 0 {
			return ""
		}
		return strconv.Itoa(req.AssignedGroupID)
	case "use_case":
		return req.UseCase
	case "dataset_type":
		return req.DatasetType
	case "format":
		return req.Format
	case "detector":
		return req.Detector
	case "statistics":
		return req.Statistics
	case "estimated_size":
		return req.EstimatedSize
	case "target_campaign":
		return req.TargetCampaign
	case "key4hep_stack":
		return req.Key4hepStack
	case "due_date":
		return req.DueDate
	case "notes":
		return req.Notes
	case "tags":
		return req.Tags
	}
	return ""
}

func setFieldValue(req *models.DatasetRequest, key, value string) {
	switch key {
	case "title":
		req.Title = value
	case "description":
		req.Description = value
	case "use_case":
		req.UseCase = value
	case "dataset_type":
		req.DatasetType = value
	case "format":
		req.Format = value
	case "detector":
		req.Detector = value
	case "statistics":
		req.Statistics = value
	case "estimated_size":
		req.EstimatedSize = value
	case "target_campaign":
		req.TargetCampaign = value
	case "key4hep_stack":
		req.Key4hepStack = value
	case "due_date":
		req.DueDate = value
	case "notes":
		req.Notes = value
	case "tags":
		req.Tags = value
	}
}

// EditField renders the edit-modal form for a single request field.
func (h *Handler) EditField(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	f, ok := requestFields[r.PathValue("field")]
	if !ok {
		http.Error(w, "Not Found", 404)
		return
	}
	req, err := h.requests.GetByID(id)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	user := middleware.GetUser(r)
	if !canEdit(user, req) || (f.Key == "group" && !canEditGroup(user, req)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if f.Key == "group" {
		groups, _ := h.groups.GetAll()
		for _, g := range groups {
			f.Options = append(f.Options, models.Option{Value: strconv.Itoa(g.ID), Label: g.Name})
		}
	}
	h.renderPartial(w, r, "field_edit", PageData{Request: req, Field: &FieldEdit{requestField: f, Value: fieldValue(req, f.Key)}})
}

// PatchRequest updates a single request field. Required fields are not
// enforced here — only when the request is submitted for review.
func (h *Handler) PatchRequest(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	existing, err := h.requests.GetByID(id)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	user := middleware.GetUser(r)
	if !canEdit(user, existing) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}
	f, ok := requestFields[r.FormValue("_field")]
	if !ok {
		http.Error(w, "unknown field", 400)
		return
	}
	value := strings.TrimSpace(r.FormValue("value"))

	switch f.Key {
	case "group":
		if !canEditGroup(user, existing) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		groupID, _ := strconv.Atoi(value)
		if groupID == existing.AssignedGroupID {
			break
		}
		if err := h.requests.AssignGroup(id, groupID); err != nil {
			slog.Error("patch request group", "error", err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		if fresh, err := h.requests.GetByID(id); err == nil {
			body := "Group unassigned"
			if fresh.AssignedGroupName != "" {
				body = "Assigned to group: " + fresh.AssignedGroupName
			}
			h.updates.Add(id, user.ID, models.UpdateAssigned, body)
		}
	case "title":
		if value == "" {
			http.Error(w, "title is required", 400)
			return
		}
		fallthrough
	default:
		setFieldValue(existing, f.Key, value)
		if err := h.requests.Update(existing); err != nil {
			slog.Error("patch request", "field", f.Key, "error", err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		if f.Key == "description" || f.Key == "notes" {
			h.relations.CreateMentions(id, user.ID, existing.Description, existing.Notes)
		}
	}

	req, err := h.requests.GetByID(id)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	h.renderPartial(w, r, "request_detail", h.requestDetailData(req))
}

// requestDetailData loads everything the request_detail partial shows.
func (h *Handler) requestDetailData(req *models.DatasetRequest) PageData {
	updates, _ := h.updates.GetByRequestID(req.ID)
	groups, _ := h.groups.GetAll()
	campaigns, _ := h.campaigns.GetAssignable(req.CampaignID)
	relations, _ := h.relations.GetByRequestID(req.ID)
	cards, _ := h.generatorCards.GetByRequestID(req.ID)
	prodIDs, _ := h.productionIDs.GetByRequestID(req.ID)
	return PageData{
		Request:        req,
		Updates:        updates,
		Groups:         groups,
		AssignedGroup:  assignedGroupFrom(req, groups),
		Campaigns:      campaigns,
		Relations:      relations,
		GeneratorCards: cards,
		ProductionIDs:  prodIDs,
	}
}
