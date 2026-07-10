package notifications

import (
	"dataset-tracker/internal/email"
	"dataset-tracker/internal/models"
	"fmt"
	"html"
	"log/slog"
	"strconv"
	"strings"
)

type Notifier struct {
	users    *models.UserStore
	emailCfg email.Config
	baseURL  string
}

func New(users *models.UserStore, cfg email.Config, baseURL string) *Notifier {
	return &Notifier{users: users, emailCfg: cfg, baseURL: baseURL}
}

// OnActivity fires email notifications for a logged activity event.
// It is safe to call from any goroutine; email is sent asynchronously.
func (n *Notifier) OnActivity(req *models.DatasetRequest, update *models.Update) {
	if !n.emailCfg.Enabled() {
		return
	}
	go n.dispatch(req, update)
}

func (n *Notifier) dispatch(req *models.DatasetRequest, update *models.Update) {
	switch update.Type {
	case models.UpdateCreated:
		if req.Status == models.StatusPending {
			n.notifyNewRequest(req, update)
		}
	case models.UpdateStatusChanged:
		n.notifyStatusChange(req, update)
	case models.UpdateComment, models.UpdateInternalNote:
		n.notifyComment(req, update)
	case models.UpdateAssigned:
		n.notifyAssigned(req, update)
	case models.UpdatePriorityChanged:
		n.notifyPriorityChange(req, update)
	}
}

func (n *Notifier) requestInfoHTML(req *models.DatasetRequest) string {
	desc := req.Description
	if len(desc) > 400 {
		desc = desc[:400] + "…"
	}
	groupRow := ""
	if req.AssignedGroupName != "" {
		groupRow = `<tr><td style="padding:4px 12px 4px 0;color:#6b7280;vertical-align:top;white-space:nowrap">Assigned group</td>` +
			`<td style="padding:4px 0">` + html.EscapeString(req.AssignedGroupName) + `</td></tr>`
	}
	linkRow := ""
	if n.baseURL != "" {
		url := n.baseURL + "/requests/" + strconv.Itoa(req.ID)
		linkRow = `<tr><td style="padding:4px 12px 4px 0;color:#6b7280;vertical-align:top;white-space:nowrap">Link</td>` +
			`<td style="padding:4px 0"><a href="` + html.EscapeString(url) + `" style="color:#2563eb">View request →</a></td></tr>`
	}
	return `<table style="border-collapse:collapse;width:100%;margin:16px 0">` +
		`<tr><td style="padding:4px 12px 4px 0;color:#6b7280;vertical-align:top;white-space:nowrap">Title</td>` +
		`<td style="padding:4px 0;font-weight:600">` + html.EscapeString(req.Title) + `</td></tr>` +
		`<tr><td style="padding:4px 12px 4px 0;color:#6b7280;vertical-align:top;white-space:nowrap">ID</td>` +
		`<td style="padding:4px 0">#` + strconv.Itoa(req.ID) + `</td></tr>` +
		`<tr><td style="padding:4px 12px 4px 0;color:#6b7280;vertical-align:top;white-space:nowrap">Requester</td>` +
		`<td style="padding:4px 0">` + html.EscapeString(req.RequesterName) + `</td></tr>` +
		groupRow +
		linkRow +
		`</table>` +
		`<p style="color:#6b7280;margin:16px 0 6px 0;font-size:12px;font-weight:600;text-transform:uppercase;letter-spacing:0.05em">Description</p>` +
		`<p style="background:#f9fafb;border-left:3px solid #e5e7eb;padding:10px 12px;margin:0;color:#374151;white-space:pre-wrap">` + html.EscapeString(desc) + `</p>`
}

func htmlWrap(greeting, inner string) string {
	return `<!DOCTYPE html><html><head><meta charset="UTF-8"></head>` +
		`<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;font-size:14px;color:#374151;max-width:600px;margin:0 auto;padding:24px">` +
		`<p style="margin:0 0 16px 0">` + greeting + `</p>` +
		inner +
		`<hr style="border:none;border-top:1px solid #e5e7eb;margin:24px 0">` +
		`<p style="color:#9ca3af;font-size:12px;margin:0">FCC Dataset Request System</p>` +
		`</body></html>`
}

func (n *Notifier) notifyNewRequest(req *models.DatasetRequest, _ *models.Update) {
	users, err := n.users.GetUsersForNewRequest(req.AssignedGroupID)
	if err != nil {
		slog.Error("notif: get users for new request", "request_id", req.ID, "error", err)
		return
	}
	subject := fmt.Sprintf("[FCC-DRS] New request #%d: %s", req.ID, req.Title)
	inner := `<p style="margin:0 0 4px 0">A new dataset request has been submitted.</p>` +
		n.requestInfoHTML(req)
	n.sendToUsers(users, req.ID, subject, inner)
}

func (n *Notifier) notifyStatusChange(req *models.DatasetRequest, update *models.Update) {
	subject := fmt.Sprintf("[FCC-DRS] Request #%d status updated: %s", req.ID, req.Title)
	inner := `<p style="margin:0 0 4px 0">Status changed: <strong>` + html.EscapeString(update.Body) + `</strong></p>` +
		n.requestInfoHTML(req)
	users, _ := n.users.GetUsersForStatusChange(req.AssignedGroupID)
	sent := n.sendToUsers(users, req.ID, subject, inner)
	n.notifyRequester(req, update.UserID, sent, subject, inner, "status")
}

func (n *Notifier) notifyComment(req *models.DatasetRequest, update *models.Update) {
	subject := fmt.Sprintf("[FCC-DRS] New comment on request #%d: %s", req.ID, req.Title)
	preview := update.Body
	if len(preview) > 400 {
		preview = preview[:400] + "…"
	}
	inner := `<p style="margin:0 0 12px 0">A new comment was added.</p>` +
		`<p style="background:#f9fafb;border-left:3px solid #e5e7eb;padding:10px 12px;margin:0 0 4px 0;color:#374151;white-space:pre-wrap">` + html.EscapeString(preview) + `</p>` +
		n.requestInfoHTML(req)
	users, _ := n.users.GetUsersForComment(req.AssignedGroupID, update.UserID)
	sent := n.sendToUsers(users, req.ID, subject, inner)
	if update.Type == models.UpdateInternalNote {
		return
	}
	n.notifyRequester(req, update.UserID, sent, subject, inner, "comment")
}

func (n *Notifier) notifyAssigned(req *models.DatasetRequest, update *models.Update) {
	var actionHTML string
	switch {
	case req.AssignedGroupName == "":
		actionHTML = "The coordinator group has been unassigned."
	case strings.HasPrefix(update.Body, "Assigned to group:"):
		actionHTML = "Request assigned to coordinator group <strong>" + html.EscapeString(req.AssignedGroupName) + "</strong>."
	default:
		// Reassignment: body is "OldGroup → NewGroup"
		parts := strings.SplitN(update.Body, " → ", 2)
		oldGroup := ""
		if len(parts) == 2 {
			oldGroup = parts[0]
		}
		actionHTML = "Coordinator group changed from <strong>" + html.EscapeString(oldGroup) +
			"</strong> to <strong>" + html.EscapeString(req.AssignedGroupName) + "</strong>."
	}
	subject := fmt.Sprintf("[FCC-DRS] Request #%d group assigned: %s", req.ID, req.Title)
	inner := `<p style="margin:0 0 4px 0">` + actionHTML + `</p>` + n.requestInfoHTML(req)
	users, _ := n.users.GetUsersForComment(req.AssignedGroupID, update.UserID)
	sent := n.sendToUsers(users, req.ID, subject, inner)
	n.notifyRequester(req, update.UserID, sent, subject, inner, "status")
}

func (n *Notifier) notifyPriorityChange(req *models.DatasetRequest, update *models.Update) {
	subject := fmt.Sprintf("[FCC-DRS] Request #%d priority changed: %s", req.ID, req.Title)
	inner := `<p style="margin:0 0 4px 0">Priority changed: <strong>` + html.EscapeString(update.Body) + `</strong></p>` +
		n.requestInfoHTML(req)
	users, _ := n.users.GetUsersForStatusChange(req.AssignedGroupID)
	n.sendToUsers(users, req.ID, subject, inner)
}

// notifyRequester sends to the request's requester if not already sent and their preference allows it.
// notifKind is "status" or "comment" and selects the right preference field.
// authorID is the activity author — requester is skipped if they are the author.
func (n *Notifier) notifyRequester(req *models.DatasetRequest, authorID int, sent map[string]bool, subject, inner, notifKind string) {
	if req.RequesterEmail == "" || sent[req.RequesterEmail] {
		return
	}
	requester := n.lookupUser(req.RequesterUsername)
	if requester != nil && requester.ID == authorID {
		return
	}
	wantsMail := true
	if requester != nil {
		switch notifKind {
		case "status":
			wantsMail = requester.NotifyStatusChanges
		case "comment":
			wantsMail = requester.NotifyComments
		}
	}
	if wantsMail {
		n.sendOne(req.RequesterEmail, req.RequesterUsername, req.ID, subject, inner)
	}
}

func (n *Notifier) sendToUsers(users []*models.User, requestID int, subject, inner string) map[string]bool {
	sent := map[string]bool{}
	for _, u := range users {
		greeting := "Hi @" + html.EscapeString(u.Username) + ","
		body := htmlWrap(greeting, inner)
		if err := n.emailCfg.Send(u.Email, subject, body); err != nil {
			slog.Error("send notification", "request_id", requestID, "user", u.Username, "error", err)
		} else {
			slog.Info("sent notification", "request_id", requestID, "user", u.Username)
			sent[u.Email] = true
		}
	}
	return sent
}

func (n *Notifier) sendOne(to, username string, requestID int, subject, inner string) {
	greeting := "Hi @" + html.EscapeString(username) + ","
	body := htmlWrap(greeting, inner)
	if err := n.emailCfg.Send(to, subject, body); err != nil {
		slog.Error("send notification", "request_id", requestID, "to", to, "error", err)
	} else {
		slog.Info("sent notification", "request_id", requestID, "to", to)
	}
}

func (n *Notifier) lookupUser(username string) *models.User {
	if username == "" {
		return nil
	}
	u, err := n.users.GetByUsername(username)
	if err != nil {
		return nil
	}
	return u
}
