package handlers

import (
	"log/slog"
	"net/http"
	"os"
)

var helpTabs = map[string]bool{"overview": true, "requester": true, "coordinator": true}

func (h *Handler) Help(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")
	if !helpTabs[tab] {
		tab = "overview"
	}

	content, err := os.ReadFile("docs/help/" + tab + ".md")
	if err != nil {
		slog.Error("read help page", "tab", tab, "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	h.renderPage(w, r, "help", PageData{
		Title:       "Help",
		HelpTab:     tab,
		HelpContent: string(content),
	})
}
