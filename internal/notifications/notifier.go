package notifications

import (
	"dataset-tracker/internal/email"
	"dataset-tracker/internal/models"
	"fmt"
	"html"
	"log/slog"
	"strconv"
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
	group := req.AssignedGroupName
	if group == "" {
		group = "Unassigned"
	}
	desc := req.Description
	if len(desc) > 400 {
		desc = desc[:400] + "…"
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
		`<tr><td style="padding:4px 12px 4px 0;color:#6b7280;vertical-align:top;white-space:nowrap">Assigned group</td>` +
		`<td style="padding:4px 0">` + html.EscapeString(group) + `</td></tr>` +
		linkRow +
		`</table>` +
		`<p style="color:#6b7280;margin:16px 0 6px 0;font-size:12px;font-weight:600;text-transform:uppercase;letter-spacing:0.05em">Description</p>` +
		`<p style="background:#f9fafb;border-left:3px solid #e5e7eb;padding:10px 12px;margin:0;color:#374151;white-space:pre-wrap">` + html.EscapeString(desc) + `</p>`
}

func htmlWrap(inner string) string {
	return `<!DOCTYPE html><html><head><meta charset="UTF-8"></head>` +
		`<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;font-size:14px;color:#374151;max-width:600px;margin:0 auto;padding:24px">` +
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
	body := htmlWrap(
		`<p style="margin:0 0 4px 0">A new dataset request has been submitted.</p>` +
			n.requestInfoHTML(req),
	)
	n.sendToUsers(users, req.ID, subject, body)
}

func (n *Notifier) notifyStatusChange(req *models.DatasetRequest, update *models.Update) {
	subject := fmt.Sprintf("[FCC-DRS] Request #%d status updated: %s", req.ID, req.Title)
	body := htmlWrap(
		`<p style="margin:0 0 4px 0">Status changed: <strong>` + html.EscapeString(update.Body) + `</strong></p>` +
			n.requestInfoHTML(req),
	)
	users, _ := n.users.GetUsersForStatusChange(req.AssignedGroupID)
	sent := n.sendToUsers(users, req.ID, subject, body)
	n.notifyRequester(req, update.UserID, sent, subject, body, "status")
}

func (n *Notifier) notifyComment(req *models.DatasetRequest, update *models.Update) {
	subject := fmt.Sprintf("[FCC-DRS] New comment on request #%d: %s", req.ID, req.Title)
	preview := update.Body
	if len(preview) > 400 {
		preview = preview[:400] + "…"
	}
	body := htmlWrap(
		`<p style="margin:0 0 12px 0">A new comment was added.</p>` +
			`<p style="background:#f9fafb;border-left:3px solid #e5e7eb;padding:10px 12px;margin:0 0 4px 0;color:#374151;white-space:pre-wrap">` + html.EscapeString(preview) + `</p>` +
			n.requestInfoHTML(req),
	)
	users, _ := n.users.GetUsersForComment(req.AssignedGroupID, update.UserID)
	sent := n.sendToUsers(users, req.ID, subject, body)
	if update.Type == models.UpdateInternalNote {
		return
	}
	n.notifyRequester(req, update.UserID, sent, subject, body, "comment")
}

func (n *Notifier) notifyAssigned(req *models.DatasetRequest, update *models.Update) {
	var actionHTML string
	if req.AssignedGroupName != "" {
		actionHTML = "Coordinator group changed to <strong>" + html.EscapeString(req.AssignedGroupName) + "</strong>"
	} else {
		actionHTML = "Coordinator group unassigned"
	}
	subject := fmt.Sprintf("[FCC-DRS] Request #%d group assigned: %s", req.ID, req.Title)
	body := htmlWrap(
		`<p style="margin:0 0 4px 0">` + actionHTML + `</p>` +
			n.requestInfoHTML(req),
	)
	users, _ := n.users.GetUsersForComment(req.AssignedGroupID, update.UserID)
	sent := n.sendToUsers(users, req.ID, subject, body)
	n.notifyRequester(req, update.UserID, sent, subject, body, "status")
}

func (n *Notifier) notifyPriorityChange(req *models.DatasetRequest, update *models.Update) {
	subject := fmt.Sprintf("[FCC-DRS] Request #%d priority changed: %s", req.ID, req.Title)
	body := htmlWrap(
		`<p style="margin:0 0 4px 0">Priority changed: <strong>` + html.EscapeString(update.Body) + `</strong></p>` +
			n.requestInfoHTML(req),
	)
	users, _ := n.users.GetUsersForStatusChange(req.AssignedGroupID)
	n.sendToUsers(users, req.ID, subject, body)
}

// notifyRequester sends to the request's requester if not already sent and their preference allows it.
// notifKind is "status" or "comment" and selects the right preference field.
// authorID is the activity author — requester is skipped if they are the author.
func (n *Notifier) notifyRequester(req *models.DatasetRequest, authorID int, sent map[string]bool, subject, body, notifKind string) {
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
		n.sendOne(req.RequesterEmail, req.ID, subject, body)
	}
}

func (n *Notifier) sendToUsers(users []*models.User, requestID int, subject, body string) map[string]bool {
	sent := map[string]bool{}
	for _, u := range users {
		if err := n.emailCfg.Send(u.Email, subject, body); err != nil {
			slog.Error("send notification", "request_id", requestID, "user", u.Username, "error", err)
		} else {
			slog.Info("sent notification", "request_id", requestID, "user", u.Username)
			sent[u.Email] = true
		}
	}
	return sent
}

func (n *Notifier) sendOne(to string, requestID int, subject, body string) {
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
