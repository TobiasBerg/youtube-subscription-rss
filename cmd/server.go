package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/TobiasBerg/youtube-subscription-rss/config"
	"github.com/TobiasBerg/youtube-subscription-rss/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/urfave/cli/v3"
)

func StartServerCMD(cfg config.AppConfig) func(ctx context.Context, c *cli.Command) error {
	return func(ctx context.Context, c *cli.Command) error {
		cache := service.NewFeedCache(15 * time.Minute)

		r := chi.NewRouter()

		r.Use(middleware.Logger)

		r.Get("/feed.xml", func(w http.ResponseWriter, r *http.Request) {
			if data, ok := cache.Get(); ok {
				log.Println("Serving feed from cache")
				w.Header().Set("Content-Type", "application/xml; charset=utf-8")
				w.Write(data)
				return
			}

			log.Println("Cache miss — regenerating feed")
			data, err := service.GenerateFeed(r.Context(), cfg)
			if err != nil {
				http.Error(w, fmt.Sprintf("error generating feed: %v", err), http.StatusInternalServerError)
				return
			}

			cache.Set(data)

			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			w.Write(data)
		})

		port := "8080"
		if len(cfg.Port) > 0 {
			port = cfg.Port
		}

		http.ListenAndServe(fmt.Sprintf(":%s", port), r)
		return nil
	}
}
