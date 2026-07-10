package notifications

import (
	"dataset-tracker/internal/email"
	"dataset-tracker/internal/models"
	"fmt"
	"log/slog"
)

type Notifier struct {
	users    *models.UserStore
	emailCfg email.Config
}

func New(users *models.UserStore, cfg email.Config) *Notifier {
	return &Notifier{users: users, emailCfg: cfg}
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
		// Only notify for submitted (non-draft) requests.
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

func (n *Notifier) notifyNewRequest(req *models.DatasetRequest, _ *models.Update) {
	users, err := n.users.GetUsersForNewRequest(req.AssignedGroupID)
	if err != nil {
		slog.Error("notif: get users for new request", "request_id", req.ID, "error", err)
		return
	}
	desc := req.Description
	if len(desc) > 500 {
		desc = desc[:500] + "…"
	}
	subject := fmt.Sprintf("[FCC-DRS] New request #%d: %s", req.ID, req.Title)
	body := fmt.Sprintf(
		"A new dataset request has been submitted.\n\nTitle: %s\nID: %d\nRequester: %s\n\nDescription:\n%s\n\nFCC Dataset Request System",
		req.Title, req.ID, req.RequesterName, desc,
	)
	n.sendToUsers(users, req.ID, subject, body)
}

func (n *Notifier) notifyStatusChange(req *models.DatasetRequest, update *models.Update) {
	subject := fmt.Sprintf("[FCC-DRS] Request #%d status updated: %s", req.ID, req.Title)
	body := fmt.Sprintf(
		"Dataset request \"%s\" (ID: %d) has been updated.\n\nChange: %s\n\nFCC Dataset Request System",
		req.Title, req.ID, update.Body,
	)
	users, _ := n.users.GetUsersForStatusChange(req.AssignedGroupID)
	sent := n.sendToUsers(users, req.ID, subject, body)
	n.notifyRequester(req, update.UserID, sent, subject, body, "status")
}

func (n *Notifier) notifyComment(req *models.DatasetRequest, update *models.Update) {
	subject := fmt.Sprintf("[FCC-DRS] New comment on request #%d: %s", req.ID, req.Title)
	preview := update.Body
	if len(preview) > 500 {
		preview = preview[:500] + "…"
	}
	body := fmt.Sprintf(
		"A new comment was added to request \"%s\" (ID: %d).\n\n%s\n\nFCC Dataset Request System",
		req.Title, req.ID, preview,
	)
	users, _ := n.users.GetUsersForComment(req.AssignedGroupID, update.UserID)
	sent := n.sendToUsers(users, req.ID, subject, body)
	if update.Type == models.UpdateInternalNote {
		return
	}
	n.notifyRequester(req, update.UserID, sent, subject, body, "comment")
}

func (n *Notifier) notifyAssigned(req *models.DatasetRequest, update *models.Update) {
	subject := fmt.Sprintf("[FCC-DRS] Request #%d group assigned: %s", req.ID, req.Title)
	body := fmt.Sprintf(
		"Dataset request \"%s\" (ID: %d) has been assigned.\n\n%s\n\nFCC Dataset Request System",
		req.Title, req.ID, update.Body,
	)
	users, _ := n.users.GetUsersForStatusChange(req.AssignedGroupID)
	sent := n.sendToUsers(users, req.ID, subject, body)
	n.notifyRequester(req, update.UserID, sent, subject, body, "status")
}

func (n *Notifier) notifyPriorityChange(req *models.DatasetRequest, update *models.Update) {
	subject := fmt.Sprintf("[FCC-DRS] Request #%d priority changed: %s", req.ID, req.Title)
	body := fmt.Sprintf(
		"Priority of request \"%s\" (ID: %d) has been changed.\n\n%s\n\nFCC Dataset Request System",
		req.Title, req.ID, update.Body,
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

// sendToUsers emails a list of users and returns the set of addresses successfully sent.
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
