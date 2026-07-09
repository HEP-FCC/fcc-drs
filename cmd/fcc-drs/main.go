package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"dataset-tracker/internal/auth"
	"dataset-tracker/internal/db"
	"dataset-tracker/internal/handlers"
	"dataset-tracker/internal/middleware"
	"dataset-tracker/internal/models"
)

// version is set at build time via -ldflags "-X main.version=vX.Y.Z".
// Falls back to "dev" for local builds without the flag.
var version = "dev"

func main() {
	devMode := os.Getenv("DEV_MODE") == "TRUE"

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	database, err := db.Init()
	if err != nil {
		log.Fatal("init db:", err)
	}
	defer database.Close()

	if devMode {
		slog.Warn("⚠  DEV_MODE enabled — CERN SSO is bypassed, do NOT use in production")
	}

	// OIDC client — optional so the app still starts without credentials configured,
	// but login will return a 503 until env vars are set.
	var oidcClient *auth.Client
	if os.Getenv("OIDC_CLIENT_ID") != "" {
		oidcClient, err = auth.NewClient(context.Background())
		if err != nil {
			log.Fatal("init OIDC:", err)
		}
		slog.Info("CERN SSO OIDC configured", "issuer", auth.CERNIssuer)
	} else {
		slog.Warn("OIDC_CLIENT_ID not set — CERN SSO login will be unavailable")
	}

	userRepo := models.NewUserStore(database.DB, database.DriverName())
	h := handlers.New(database.DB, database.DriverName(), oidcClient, devMode, version)

	authMW := middleware.Auth(userRepo)

	mux := http.NewServeMux()

	// Static assets — public
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Auth — public
	mux.HandleFunc("GET /login", h.ShowLogin)
	mux.HandleFunc("GET /auth/login", h.Login)
	mux.HandleFunc("GET /auth/callback", h.Callback)
	mux.HandleFunc("POST /auth/dev-login", h.DevLogin) // only active when DEV_MODE=true
	mux.HandleFunc("GET /logout", h.Logout)

	// Avatar — public (served by username, cached)
	mux.HandleFunc("GET /users/{username}/avatar", h.ServeAvatar)

	// Protected routes — require authenticated session
	mux.HandleFunc("GET /", middleware.RequireAuth(h.Dashboard))
	mux.HandleFunc("GET /profile", middleware.RequireAuth(h.ShowProfile))
	mux.HandleFunc("POST /profile", middleware.RequireAuth(h.UpdateProfile))
	mux.HandleFunc("POST /profile/avatar/delete", middleware.RequireAuth(h.DeleteAvatar))
	mux.HandleFunc("GET /requests/new", middleware.RequireAuth(h.NewRequestForm))
	mux.HandleFunc("GET /requests", middleware.RequireAuth(h.ListRequests))
	mux.HandleFunc("POST /requests", middleware.RequireAuth(h.CreateRequest))
	mux.HandleFunc("GET /requests/{id}", middleware.RequireAuth(h.GetRequest))
	mux.HandleFunc("GET /requests/{id}/edit", middleware.RequireAuth(h.EditRequestForm))
	mux.HandleFunc("GET /requests/{id}/clone", middleware.RequireAuth(h.GetCloneForm))
	mux.HandleFunc("POST /requests/{id}", middleware.RequireAuth(h.UpdateRequest))
	mux.HandleFunc("GET /api/stats", middleware.RequireAuth(h.GetStats))
	mux.HandleFunc("GET /api/recent", middleware.RequireAuth(h.GetRecent))

	// Admin-only routes
	mux.HandleFunc("GET /admin/users", middleware.RequireAdmin(h.AdminUsers))
	mux.HandleFunc("GET /admin/groups", middleware.RequireAdmin(h.AdminGroups))
	mux.HandleFunc("POST /admin/groups", middleware.RequireAdmin(h.AdminCreateGroup))
	mux.HandleFunc("DELETE /admin/groups/{id}", middleware.RequireAdmin(h.AdminDeleteGroup))
	mux.HandleFunc("POST /admin/groups/{id}/members", middleware.RequireAdmin(h.AdminAddGroupMember))
	mux.HandleFunc("DELETE /admin/groups/{id}/members/{user_id}", middleware.RequireAdmin(h.AdminRemoveGroupMember))
	mux.HandleFunc("POST /admin/users/{id}/role", middleware.RequireAdmin(h.AdminUpdateUserRole))

	// Coordinator-only routes
	mux.HandleFunc("GET /coordinator", middleware.RequireCoordinator(h.CoordinatorView))
	mux.HandleFunc("POST /requests/batch", middleware.RequireCoordinator(h.BatchAction))
	mux.HandleFunc("GET /requests/{id}/section/{section}", middleware.RequireAuth(h.ViewSection))
	mux.HandleFunc("GET /requests/{id}/section/{section}/edit", middleware.RequireAuth(h.EditSection))
	mux.HandleFunc("PATCH /requests/{id}", middleware.RequireAuth(h.PatchRequest))
	mux.HandleFunc("POST /requests/{id}/status", middleware.RequireAuth(h.UpdateStatus))
	mux.HandleFunc("POST /requests/{id}/approval", middleware.RequireCoordinator(h.ApprovalDecision))
	mux.HandleFunc("POST /requests/{id}/assign", middleware.RequireCoordinator(h.AssignRequest))
	mux.HandleFunc("POST /requests/{id}/priority", middleware.RequireCoordinator(h.UpdatePriority))
	mux.HandleFunc("POST /requests/{id}/comments", middleware.RequireAuth(h.AddComment))
	mux.HandleFunc("GET /requests/{id}/comments/{comment_id}", middleware.RequireAuth(h.GetComment))
	mux.HandleFunc("GET /requests/{id}/comments/{comment_id}/edit", middleware.RequireAuth(h.EditCommentForm))
	mux.HandleFunc("PATCH /requests/{id}/comments/{comment_id}", middleware.RequireAuth(h.PatchComment))
	mux.HandleFunc("DELETE /requests/{id}/comments/{comment_id}", middleware.RequireAuth(h.DeleteComment))
	mux.HandleFunc("POST /requests/{id}/relations", middleware.RequireAuth(h.AddRelation))
	mux.HandleFunc("DELETE /requests/{id}/relations/{rel_id}", middleware.RequireAuth(h.RemoveRelation))
	mux.HandleFunc("POST /requests/{id}/generator-cards", middleware.RequireAuth(h.UploadGeneratorCard))
	mux.HandleFunc("GET /requests/{id}/generator-cards/{card_id}/view", middleware.RequireAuth(h.ViewGeneratorCard))
	mux.HandleFunc("GET /requests/{id}/generator-cards/{card_id}/download", middleware.RequireAuth(h.DownloadGeneratorCard))
	mux.HandleFunc("DELETE /requests/{id}/generator-cards/{card_id}", middleware.RequireAuth(h.DeleteGeneratorCard))
	mux.HandleFunc("DELETE /requests/{id}", middleware.RequireAuth(h.DeleteRequest))

	port := os.Getenv("PORT")
	if port == "" {
		port = "5050"
	}
	addr := ":" + port
	slog.Info("server started", "addr", "http://localhost"+addr)
	if err := http.ListenAndServe(addr, authMW(mux)); err != nil {
		log.Fatal(err)
	}
}
