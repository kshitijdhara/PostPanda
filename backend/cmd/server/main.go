package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/kshitijdhara/blog/internal/config"
	"github.com/kshitijdhara/blog/internal/handlers"
	"github.com/kshitijdhara/blog/internal/logger"
	mw "github.com/kshitijdhara/blog/internal/middleware"
	"github.com/kshitijdhara/blog/internal/repositories"
	"github.com/kshitijdhara/blog/internal/services"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.Env)
	slog.SetDefault(log)

	// Connect to PostgreSQL
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Run migrations
	if err := runMigrations(db, log); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Repositories
	userRepo := repositories.NewUserRepository(db)
	postRepo := repositories.NewPostRepository(db)
	commentRepo := repositories.NewCommentRepository(db)

	// Services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	postService := services.NewPostService(postRepo)
	commentService := services.NewCommentService(commentRepo, postRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	postHandler := handlers.NewPostHandler(postService, authService)
	commentHandler := handlers.NewCommentHandler(commentService, authService)
	userHandler := handlers.NewUserHandler(authService, postService, commentService)

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(mw.Recovery(log))
	r.Use(mw.Logger(log))

	// Routes
	r.Get("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Mount("/api/v1/auth", authHandler.Routes())
	r.Route("/api/v1/posts", func(r chi.Router) {
		// Post CRUD routes
		r.With(mw.OptionalAuth(authService)).Get("/", postHandler.List)
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(authService))
			r.Get("/drafts/mine", postHandler.ListDrafts)
			r.Post("/", postHandler.Create)
		})
		r.Route("/{slug}", func(r chi.Router) {
			r.With(mw.OptionalAuth(authService)).Get("/", postHandler.GetBySlug)
			r.With(mw.Auth(authService)).Put("/", postHandler.Update)
			r.With(mw.Auth(authService)).Delete("/", postHandler.Delete)
			r.With(mw.Auth(authService)).Post("/like", postHandler.ToggleLike)
			r.With(mw.Auth(authService)).Post("/bookmark", postHandler.ToggleBookmark)
			// Comment routes nested under post
			r.With(mw.OptionalAuth(authService)).Get("/comments", commentHandler.ListByPost)
			r.With(mw.Auth(authService)).Post("/comments", commentHandler.Create)
		})
	})
	r.With(mw.Auth(authService)).Delete("/api/v1/comments/{id}", commentHandler.Delete)
	r.With(mw.Auth(authService)).Post("/api/v1/comments/{id}/vote", commentHandler.Vote)
	r.Mount("/api/v1/users", userHandler.Routes(authService))

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", "error", err)
	}
	log.Info("server stopped")
}

func runMigrations(db *sqlx.DB, log *slog.Logger) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("creating migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("creating migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}

	log.Info("migrations completed")
	return nil
}
