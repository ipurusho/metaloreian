package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/imman/metaloreian/internal/api"
	"github.com/imman/metaloreian/internal/config"
	"github.com/imman/metaloreian/internal/matcher"
	"github.com/imman/metaloreian/internal/scraper"
	"github.com/imman/metaloreian/internal/store"
)

func main() {
	cfg := config.Load()

	// Database is optional — scrape-only mode if unavailable
	var s *store.Store
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err == nil {
		if err := db.Ping(); err != nil {
			log.Printf("database unavailable, running in scrape-only mode: %v", err)
		} else {
			log.Println("connected to database")
			s = store.New(db)
			defer db.Close()
		}
	} else {
		log.Printf("database unavailable, running in scrape-only mode: %v", err)
	}

	sc := scraper.NewClient(cfg.FlareSolverrURL)
	m := matcher.New(s, sc)

	router := api.NewRouter(cfg, s, m, m)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
