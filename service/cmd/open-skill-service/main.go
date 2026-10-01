// Command open-skill-service runs the Open-Skill ranking API.
//
// It opens (creating) the SQLite database, applies the idempotent schema, and
// serves the HTTP API. The database starts empty; demo content is only loaded
// when -seed (or OPEN_SKILL_SEED=1) is set.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"open-skill/service/internal/api"
	"open-skill/service/internal/seed"
	"open-skill/service/internal/store"
)

func main() {
	var (
		dbPath   = flag.String("db", "open-skill.db", "path to the SQLite database (ignored when OPEN_SKILL_DB_URL is set)")
		addr     = flag.String("addr", ":8080", "http listen address")
		withSeed = flag.Bool("seed", os.Getenv("OPEN_SKILL_SEED") == "1", "load demo content into an empty database")
	)
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		st     *store.Store
		err    error
		dbDesc = *dbPath
	)
	if dbURL := os.Getenv("OPEN_SKILL_DB_URL"); dbURL != "" {
		dbDesc = "remote libsql"
		st, err = store.OpenRemote(dbURL, os.Getenv("OPEN_SKILL_DB_TOKEN"))
	} else {
		st, err = store.Open(*dbPath)
	}
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.Close()

	if *withSeed {
		if err := seed.Seed(st.DB()); err != nil {
			log.Fatalf("seed: %v", err)
		}
		log.Print("demo content seeded")
	}

	srv := &http.Server{
		Addr:    *addr,
		Handler: api.New(st).Handler(),
	}

	go func() {
		log.Printf("open-skill service listening on %s (db=%s)", *addr, dbDesc)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Print("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
