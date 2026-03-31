package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	msqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jaykay/vereinstool/backend/config"
	"github.com/jaykay/vereinstool/backend/db/generated"
	"github.com/jaykay/vereinstool/backend/handler"
	"github.com/jaykay/vereinstool/backend/service"
)

//go:embed db/migrations/*.sql
var migrationsFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	// Open database
	db, err := sql.Open("sqlite3", cfg.DBPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	queries := generated.New(db)

	// Services
	mailer := service.NewMailer(cfg)
	authService := service.NewAuth(queries, cfg)

	// Seed admin user
	if cfg.SeedAdminEmail != "" && cfg.SeedAdminPassword != "" {
		if err := authService.SeedAdmin(context.Background(), cfg.SeedAdminEmail, cfg.SeedAdminPassword); err != nil {
			log.Printf("Warning: failed to seed admin: %v", err)
		}
	}

	// Start session cleanup goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go authService.StartSessionCleanup(ctx, 24*time.Hour)

	// Handlers
	authHandler := handler.NewAuth(authService, mailer, cfg)
	usersHandler := handler.NewUsers(queries, authService, mailer, cfg)

	// Router
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(handler.CORS(cfg.AppURL))

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", handler.Health)

		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/logout", authHandler.Logout)
		r.Post("/auth/forgot-password", authHandler.ForgotPassword)
		r.Post("/auth/reset-password", authHandler.ResetPassword)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(handler.AuthRequired(authService))

			r.Get("/auth/me", authHandler.Me)

			// Admin-only routes
			r.Group(func(r chi.Router) {
				r.Use(handler.AdminRequired)
				r.Get("/users", usersHandler.List)
				r.Post("/users/invite", usersHandler.Invite)
				r.Patch("/users/{id}", usersHandler.Update)
			})
		})
	})

	// Serve SvelteKit SPA
	staticContent, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatalf("Failed to create static sub-filesystem: %v", err)
	}
	handler.ServeSPA(r, staticContent)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}

func runMigrations(db *sql.DB) error {
	sourceDriver, err := iofs.New(migrationsFS, "db/migrations")
	if err != nil {
		return fmt.Errorf("creating migration source: %w", err)
	}
	dbDriver, err := msqlite3.WithInstance(db, &msqlite3.Config{})
	if err != nil {
		return fmt.Errorf("creating migration db driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", dbDriver)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}
	log.Println("Migrations applied successfully")
	return nil
}
