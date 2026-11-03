package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexandrmotologa/searchcraft/internal/index"
	"github.com/alexandrmotologa/searchcraft/internal/storage"
)

// Config holds runtime server options.
type Config struct {
	Port      int
	DBPath    string
	Version   string
	AuthToken string
	CORS      bool
}

// Server encapsulates the HTTP API and background engine.
type Server struct {
	cfg       Config
	engine    *index.Engine
	db        *storage.DB
	startTime time.Time
	httpSrv   *http.Server
}

// NewServer initializes the server with database and search engine.
func NewServer(cfg Config, eng *index.Engine, db *storage.DB) *Server {
	if cfg.Port <= 0 {
		cfg.Port = 7700
	}
	if cfg.Version == "" {
		cfg.Version = "1.0.0"
	}

	s := &Server{
		cfg:       cfg,
		engine:    eng,
		db:        db,
		startTime: time.Now(),
	}

	handler := s.routes()
	s.httpSrv = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start runs the HTTP server and handles graceful shutdown on SIGINT/SIGTERM.
func (s *Server) Start() error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("[SearchCraft] Listening on http://localhost:%d (v%s)", s.cfg.Port, s.cfg.Version)
		log.Printf("[SearchCraft] Admin Studio: http://localhost:%d", s.cfg.Port)
		log.Printf("[SearchCraft] Widget Script: http://localhost:%d/widget.js", s.cfg.Port)
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-shutdown:
		log.Println("[SearchCraft] Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.httpSrv.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		log.Println("[SearchCraft] Server stopped cleanly")
	}

	return nil
}

// Handler returns the HTTP handler (useful for testing).
func (s *Server) Handler() http.Handler {
	return s.routes()
}
