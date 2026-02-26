package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"ocm/internal/api"
	"ocm/internal/dashboard"
	"ocm/internal/pipeline"
	"ocm/internal/storage"
)

func main() {
	log.SetFlags(0)

	var (
		repoPath   = flag.String("repo", ".", "Path to repo/workspace to analyze")
		dbPath     = flag.String("db", "ocm.sqlite", "SQLite DB file path")
		host       = flag.String("host", "127.0.0.1", "HTTP server host")
		port       = flag.Int("port", 8080, "HTTP server port")
		serviceKey = flag.String("service-key", "dir", "Service identity strategy: dir|manifest")
		cvWindow   = flag.Int("cv-window", 30, "Change Volatility lookback window in days")
		printJSON  = flag.Bool("print", false, "Print last run results as JSON")
	)

	flag.Parse()

	absRepo, err := filepath.Abs(*repoPath)
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	st, err := storage.New(db)
	if err != nil {
		log.Fatal(err)
	}
	if err := st.Migrate(context.Background()); err != nil {
		log.Fatal(err)
	}

	pipe := pipeline.New(pipeline.Options{
		RepoPath:           absRepo,
		ServiceKeyStrategy: *serviceKey,
		CVWindow:           time.Duration(*cvWindow) * 24 * time.Hour,
		Now:                time.Now,
		Logger:             log.New(os.Stderr, "", 0),
	})

	res, err := pipe.Run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	if err := st.SaveRun(context.Background(), storage.SaveRunInput{RunAt: res.RunAt, Services: res.Services}); err != nil {
		log.Fatal(err)
	}

	if *printJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
	}

	addr := net.JoinHostPort(*host, fmt.Sprintf("%d", *port))
	apiHandler := api.New(api.Dependencies{Store: st})
	uiHandler := dashboard.New()

	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", apiHandler))
	mux.Handle("/", uiHandler)

	srv := &http.Server{
		Addr:              addr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("ocm: analyzed %d services from %s", len(res.Services), absRepo)
	log.Printf("ocm: db=%s", *dbPath)
	log.Printf("ocm: http://%s", addr)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if strings.EqualFold(r.Method, http.MethodOptions) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
