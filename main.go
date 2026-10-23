package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"webmail/internal/api"
	"webmail/internal/store"
	syncmgr "webmail/internal/sync"
)

//go:embed all:web/dist
var uiEmbed embed.FS

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	dataDir := flag.String("data", "./data", "data directory (sqlite db lives here)")
	flag.Parse()

	if err := os.MkdirAll(*dataDir, 0o700); err != nil {
		log.Fatalf("create data dir: %v", err)
	}
	st, err := store.Open(filepath.Join(*dataDir, "webmail.db"))
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer st.DB.Close()

	uiFS, err := fs.Sub(uiEmbed, "web/dist")
	if err != nil {
		log.Fatalf("ui fs: %v", err)
	}

	mgr := syncmgr.NewManager(st)
	mgr.RecomputeSnippets()
	srv := api.New(st, uiFS, mgr)
	log.Printf("webmail listening on %s (data: %s)", *addr, *dataDir)
	log.Fatal(http.ListenAndServe(*addr, srv.Handler()))
}
