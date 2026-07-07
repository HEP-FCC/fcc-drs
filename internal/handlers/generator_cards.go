package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"

	"dataset-tracker/internal/middleware"
	"dataset-tracker/internal/models"
)

const maxGeneratorCardSize = 1 << 20  // 1 MB
const maxGeneratorCardsPerUpload = 20

// saveGeneratorCardFromForm reads all "generator_card" files from a parsed
// multipart form and persists them. Silently skips if no files were provided.
func (h *Handler) saveGeneratorCardFromForm(r *http.Request, requestID, userID int) {
	if r.MultipartForm == nil {
		return
	}
	headers := r.MultipartForm.File["generator_card"]
	if len(headers) > maxGeneratorCardsPerUpload {
		headers = headers[:maxGeneratorCardsPerUpload]
	}
	seen := make(map[string]bool)
	for _, header := range headers {
		name := filepath.Base(header.Filename)
		if seen[name] {
			continue
		}
		seen[name] = true
		file, err := header.Open()
		if err != nil {
			continue
		}
		content, err := io.ReadAll(io.LimitReader(file, maxGeneratorCardSize+1))
		file.Close()
		if err != nil || int64(len(content)) > maxGeneratorCardSize || bytes.IndexByte(content, 0) >= 0 {
			continue
		}
		card := &models.GeneratorCard{
			RequestID:  requestID,
			Filename:   name,
			Size:       int64(len(content)),
			Content:    content,
			UploadedBy: userID,
		}
		if _, err := h.generatorCards.Add(card); err != nil {
			slog.Error("save generator card from form", "error", err)
		}
	}
}

func (h *Handler) UploadGeneratorCard(w http.ResponseWriter, r *http.Request) {
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
	user := middleware.GetUser(r)
	if !canEdit(user, req) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(maxGeneratorCardSize*10 + 512); err != nil {
		http.Error(w, "Bad Request", 400)
		return
	}

	headers := r.MultipartForm.File["generator_card"]
	if len(headers) == 0 {
		http.Error(w, "no file provided", 400)
		return
	}
	if len(headers) > maxGeneratorCardsPerUpload {
		http.Error(w, "too many files (max 20 per upload)", 400)
		return
	}

	userID := 0
	if user != nil {
		userID = user.ID
	}

	existing, _ := h.generatorCards.GetByRequestID(id)
	taken := make(map[string]bool, len(existing))
	for _, c := range existing {
		taken[c.Filename] = true
	}

	seen := make(map[string]bool)
	for _, header := range headers {
		name := filepath.Base(header.Filename)
		if taken[name] || seen[name] {
			continue
		}
		seen[name] = true
		file, err := header.Open()
		if err != nil {
			continue
		}
		content, err := io.ReadAll(io.LimitReader(file, maxGeneratorCardSize+1))
		file.Close()
		if err != nil || int64(len(content)) > maxGeneratorCardSize || bytes.IndexByte(content, 0) >= 0 {
			continue
		}
		card := &models.GeneratorCard{
			RequestID:  id,
			Filename:   name,
			Size:       int64(len(content)),
			Content:    content,
			UploadedBy: userID,
		}
		if _, err := h.generatorCards.Add(card); err != nil {
			slog.Error("upload generator card", "error", err)
		}
	}

	cards, _ := h.generatorCards.GetByRequestID(id)
	h.renderPartial(w, r, "generator_cards", PageData{Request: req, GeneratorCards: cards})
}

func (h *Handler) ViewGeneratorCard(w http.ResponseWriter, r *http.Request) {
	cardID, err := strconv.Atoi(r.PathValue("card_id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	card, err := h.generatorCards.GetByID(cardID)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	h.renderPartial(w, r, "generator_card_viewer", PageData{GeneratorCard: card})
}

func (h *Handler) DownloadGeneratorCard(w http.ResponseWriter, r *http.Request) {
	cardID, err := strconv.Atoi(r.PathValue("card_id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	card, err := h.generatorCards.GetByID(cardID)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}

	ct := mime.TypeByExtension(filepath.Ext(card.Filename))
	if ct == "" {
		ct = "application/octet-stream"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, card.Filename))
	w.Header().Set("Content-Length", strconv.FormatInt(card.Size, 10))
	w.Write(card.Content) //nolint:errcheck
}

func (h *Handler) DeleteGeneratorCard(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	cardID, err := strconv.Atoi(r.PathValue("card_id"))
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	card, err := h.generatorCards.GetByID(cardID)
	if err != nil || card.RequestID != id {
		http.Error(w, "Not Found", 404)
		return
	}
	req, err := h.requests.GetByID(id)
	if err != nil {
		http.Error(w, "Not Found", 404)
		return
	}
	user := middleware.GetUser(r)
	ownerCanDelete := user != nil && card.UploadedBy == user.ID && canEdit(user, req)
	if !user.IsCoordinator() && !ownerCanDelete {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.generatorCards.Delete(cardID); err != nil {
		slog.Error("delete generator card", "error", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	cards, _ := h.generatorCards.GetByRequestID(id)
	h.renderPartial(w, r, "generator_cards", PageData{Request: req, GeneratorCards: cards})
}
