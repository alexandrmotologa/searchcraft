package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/searchcraft/internal/embedding"
	"github.com/alexandrmotologa/searchcraft/internal/index"
	"github.com/alexandrmotologa/searchcraft/internal/storage"
)

func TestCrawlerSitemapAndHTML(t *testing.T) {
	// Mock server serving sitemap and html pages
	var serverURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sitemap.xml":
			w.Header().Set("Content-Type", "application/xml")
			fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>%s/docs/getting-started</loc></url>
  <url><loc>%s/docs/architecture</loc></url>
</urlset>`, serverURL, serverURL)
		case "/docs/getting-started":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
  <title>Getting Started with SearchCraft</title>
  <meta name="description" content="Quickstart tutorial for instant hybrid search." />
</head>
<body>
  <nav>Menu items</nav>
  <h1>Getting Started</h1>
  <p>SearchCraft runs as a single Go binary with under 40MB RAM.</p>
  <footer>Footer copyright</footer>
</body>
</html>`)
		case "/docs/architecture":
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
  <title>SearchCraft Architecture</title>
</head>
<body>
  <h1>Engine Architecture</h1>
  <p>Combines SQLite FTS5 BM25 with dense vector cosine similarity via Reciprocal Rank Fusion.</p>
</body>
</html>`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	serverURL = ts.URL

	tmpDir, err := os.MkdirTemp("", "sc_crawl_test_*")
	if err != nil {
		t.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "crawl.db")
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	emb := embedding.NewFastLocalEmbedder(256)
	eng, err := index.NewEngine(db, emb)
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}

	_ = eng.CreateIndex("webdocs", nil, "hybrid", "fast_local")

	c := NewCrawler()
	res, err := c.Crawl(context.Background(), eng, "webdocs", storage.CrawlRequest{
		URL:      ts.URL + "/sitemap.xml",
		MaxPages: 10,
	})
	if err != nil {
		t.Fatalf("crawl failed: %v", err)
	}

	if res.DocsIndexed != 2 {
		t.Errorf("expected 2 docs indexed, got %d", res.DocsIndexed)
	}

	// Verify we can search the crawled docs
	searchRes, err := eng.Search(context.Background(), "webdocs", storage.SearchRequest{
		Query: "instant hybrid search",
		Mode:  "hybrid",
	})
	if err != nil || len(searchRes.Hits) == 0 {
		t.Fatalf("expected search hits for crawled docs: %v", err)
	}
	if searchRes.Hits[0].Document["title"] != "Getting Started with SearchCraft" {
		t.Errorf("unexpected top hit title: %v", searchRes.Hits[0].Document["title"])
	}
}
