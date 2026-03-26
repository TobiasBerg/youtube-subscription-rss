package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TobiasBerg/youtube-subscription-rss/config"
	"github.com/TobiasBerg/youtube-subscription-rss/service"
	static "github.com/TobiasBerg/youtube-subscription-rss/static"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/urfave/cli/v3"
)

func StartServerCMD(cfg config.AppConfig) func(ctx context.Context, c *cli.Command) error {
	return func(ctx context.Context, c *cli.Command) error {
		cache := service.NewFeedCache(time.Duration(cfg.RefreshInterval)*time.Minute, cfg)

		go func() {
			if err := cache.Start(); err != nil {
				log.Printf("Cache failed to start: %v\n", err)
				os.Exit(1)
			}
		}()

		r := chi.NewRouter()

		r.Use(middleware.Logger)

		staticFS, err := fs.Sub(static.Files, ".")
		if err != nil {
			return fmt.Errorf("error creating static sub-FS: %w", err)
		}
		fileServer := http.FileServer(http.FS(staticFS))
		r.Handle("/favicon.ico", fileServer)
		r.Handle("/android-chrome-512x512.png", fileServer)

		r.Get("/feed.xml", func(w http.ResponseWriter, r *http.Request) {
			data, ok := cache.Get()
			if !ok {
				http.Error(w, "feed not available yet, please retry shortly", http.StatusServiceUnavailable)
				return
			}

			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			w.Write(data)
		})

		port := "8080"
		if len(cfg.Port) > 0 {
			port = cfg.Port
		}

		server := &http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: r,
		}

		go func() {
			log.Printf("Server starting on port %s", port)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("Server error: %v\n", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Println("Shutting down server...")
		cache.Stop()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
