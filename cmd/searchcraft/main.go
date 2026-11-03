package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/alexandrmotologa/searchcraft/internal/embedding"
	"github.com/alexandrmotologa/searchcraft/internal/index"
	"github.com/alexandrmotologa/searchcraft/internal/server"
	"github.com/alexandrmotologa/searchcraft/internal/storage"
)

var (
	Version = "1.0.0"
	Commit  = "head"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "serve":
		cmdServe(os.Args[2:])
	case "import":
		cmdImport(os.Args[2:])
	case "benchmark":
		cmdBenchmark(os.Args[2:])
	case "version":
		fmt.Printf("SearchCraft v%s (commit: %s)\n", Version, Commit)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`SearchCraft: Single-Binary Hybrid Search Engine (< 40MB RAM)

Usage:
  searchcraft <command> [arguments]

Commands:
  serve       Start the HTTP server, REST API, widget server, and admin studio
  import      Bulk ingest a JSON document file into an index
  benchmark   Run performance benchmark against an index
  version     Show application version

Run 'searchcraft <command> -h' for more details on a specific command.`)
}

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 7700, "HTTP listen port")
	dbPath := fs.String("db", "./searchcraft.db", "SQLite database file path")
	authToken := fs.String("auth-token", "", "Optional bearer token for API security")
	provider := fs.String("embedding-provider", "fast_local", "Embedding provider (fast_local, openai, ollama, onnx)")
	onnxModel := fs.String("onnx-model", "", "Path to local ONNX model file (for onnx provider)")
	openAIKey := fs.String("openai-key", os.Getenv("OPENAI_API_KEY"), "OpenAI API key (for openai provider)")
	_ = fs.Parse(args)

	// Initialize SQLite Database
	db, err := storage.Open(*dbPath)
	if err != nil {
		log.Fatalf("Fatal: failed to open SQLite database at %s: %v", *dbPath, err)
	}
	defer db.Close()

	// Select Embedder
	var emb embedding.Embedder
	switch strings.ToLower(*provider) {
	case "openai":
		emb = embedding.NewRemoteEmbedder(embedding.RemoteConfig{
			Provider: "openai",
			APIKey:   *openAIKey,
		})
	case "ollama":
		emb = embedding.NewRemoteEmbedder(embedding.RemoteConfig{
			Provider: "ollama",
		})
	case "onnx":
		onnxEmb, err := embedding.NewONNXLocalEmbedder(*onnxModel)
		if err != nil {
			log.Printf("Warning: failed to initialize ONNX model, falling back to fast_local: %v", err)
			emb = embedding.NewFastLocalEmbedder(256)
		} else {
			emb = onnxEmb
		}
	default:
		emb = embedding.NewFastLocalEmbedder(256)
	}

	// Initialize Search Engine
	eng, err := index.NewEngine(db, emb)
	if err != nil {
		log.Fatalf("Fatal: failed to initialize search engine: %v", err)
	}

	// Seed default products index if empty
	indexes, _ := eng.ListIndexes()
	if len(indexes) == 0 {
		log.Println("[SearchCraft] No existing indexes found. Seeding initial 'products' index...")
		_ = eng.CreateIndex("products", map[string]float64{
			"title":       3.0,
			"description": 1.0,
			"brand":       1.5,
			"tags":        2.0,
		}, "hybrid", emb.Name())

		seedDocs := getInitialSeedDocs()
		_, _ = eng.IndexDocuments(context.Background(), "products", seedDocs)
		log.Printf("[SearchCraft] Successfully seeded %d documents into 'products' index.", len(seedDocs))
	}

	// Print Startup Banner
	fmt.Println("===============================================================")
	fmt.Printf("  SearchCraft v%s\n", Version)
	fmt.Println("  Lightweight Hybrid Instant Search Engine (< 40MB RAM)")
	fmt.Println("===============================================================")
	fmt.Printf("  HTTP API & Studio : http://localhost:%d\n", *port)
	fmt.Printf("  Widget Script     : http://localhost:%d/widget.js\n", *port)
	fmt.Printf("  Database Path     : %s\n", *dbPath)
	fmt.Printf("  Embedding Provider: %s (%d dimensions)\n", emb.Name(), emb.Dimension())
	if *authToken != "" {
		fmt.Println("  Authentication    : Enabled (Bearer token required for mutations)")
	} else {
		fmt.Println("  Authentication    : Disabled (Open API mode)")
	}
	fmt.Println("===============================================================")

	srv := server.NewServer(server.Config{
		Port:      *port,
		DBPath:    *dbPath,
		Version:   Version,
		AuthToken: *authToken,
		CORS:      true,
	}, eng, db)

	if err := srv.Start(); err != nil {
		log.Fatalf("Fatal server error: %v", err)
	}
}

func cmdImport(args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	dbPath := fs.String("db", "./searchcraft.db", "SQLite database file path")
	indexName := fs.String("index", "products", "Target index name")
	filePath := fs.String("file", "", "Path to JSON file to import")
	_ = fs.Parse(args)

	if *filePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --file argument is required")
		fs.Usage()
		os.Exit(1)
	}

	fileBytes, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("Fatal: failed to read file %s: %v", *filePath, err)
	}

	var rawDocs []map[string]any
	if err := json.Unmarshal(fileBytes, &rawDocs); err != nil {
		log.Fatalf("Fatal: invalid JSON format: %v", err)
	}

	db, err := storage.Open(*dbPath)
	if err != nil {
		log.Fatalf("Fatal: open db: %v", err)
	}
	defer db.Close()

	emb := embedding.NewFastLocalEmbedder(256)
	eng, err := index.NewEngine(db, emb)
	if err != nil {
		log.Fatalf("Fatal: new engine: %v", err)
	}

	// Ensure index exists
	if _, err := eng.GetIndex(*indexName); err != nil {
		_ = eng.CreateIndex(*indexName, nil, "hybrid", "fast_local")
	}

	docs := make([]storage.Document, len(rawDocs))
	for i, item := range rawDocs {
		id := fmt.Sprintf("%v", item["id"])
		if id == "" || id == "<nil>" {
			id = fmt.Sprintf("doc_%d_%d", time.Now().UnixNano(), i)
			item["id"] = id
		}
		docs[i] = storage.Document{
			ID:   id,
			Data: item,
		}
	}

	start := time.Now()
	indexed, err := eng.IndexDocuments(context.Background(), *indexName, docs)
	if err != nil {
		log.Fatalf("Fatal: indexing failed: %v", err)
	}

	duration := time.Since(start)
	rate := float64(indexed) / duration.Seconds()
	fmt.Printf("Successfully indexed %d documents into '%s' in %v (%.1f docs/sec)\n", indexed, *indexName, duration, rate)
}

func cmdBenchmark(args []string) {
	fs := flag.NewFlagSet("benchmark", flag.ExitOnError)
	dbPath := fs.String("db", "./searchcraft.db", "SQLite database file path")
	indexName := fs.String("index", "products", "Target index name")
	numQueries := fs.Int("queries", 200, "Number of queries to run")
	_ = fs.Parse(args)

	db, err := storage.Open(*dbPath)
	if err != nil {
		log.Fatalf("Fatal: open db: %v", err)
	}
	defer db.Close()

	emb := embedding.NewFastLocalEmbedder(256)
	eng, err := index.NewEngine(db, emb)
	if err != nil {
		log.Fatalf("Fatal: new engine: %v", err)
	}

	testQueries := []string{
		"wireless mechanical keyboard",
		"ergonomic mouse bluetooth",
		"custom aluminum switches",
		"optical gaming sensor",
		"rgb backlit quiet",
		"sqlite fulltext search",
		"reciprocal rank fusion",
		"damerau levenshtein typo",
	}

	fmt.Printf("Running %d hybrid search queries against index '%s'...\n", *numQueries, *indexName)

	ctx := context.Background()
	var latencies []float64
	startTotal := time.Now()

	for i := 0; i < *numQueries; i++ {
		q := testQueries[rand.Intn(len(testQueries))]
		qStart := time.Now()
		res, err := eng.Search(ctx, *indexName, storage.SearchRequest{
			Query: q,
			Mode:  "hybrid",
			Limit: 10,
		})
		if err != nil {
			log.Fatalf("Query failed: %v", err)
		}
		tookMs := float64(time.Since(qStart).Microseconds()) / 1000.0
		latencies = append(latencies, tookMs)
		_ = res
	}

	totalDuration := time.Since(startTotal)
	qps := float64(*numQueries) / totalDuration.Seconds()

	var sum float64
	for _, l := range latencies {
		sum += l
	}
	avg := sum / float64(len(latencies))

	fmt.Println("\n--- Benchmark Results ---")
	fmt.Printf("Total Queries : %d\n", *numQueries)
	fmt.Printf("Total Time    : %v\n", totalDuration)
	fmt.Printf("Throughput    : %.1f queries/sec (QPS)\n", qps)
	fmt.Printf("Average Latency: %.2f ms\n", avg)
	fmt.Println("-------------------------")
}

func getInitialSeedDocs() []storage.Document {
	sample := []map[string]any{
		{
			"id":          "prod_1",
			"title":       "Keychron Q1 Pro Wireless Custom Mechanical Keyboard",
			"description": "Compact 75% layout, CNC aluminum body, hot-swappable switches, wireless Bluetooth 5.1 and Type-C connectivity.",
			"category":    "Keyboards",
			"brand":       "Keychron",
			"price":       199.0,
			"rating":      4.9,
			"in_stock":    true,
			"tags":        []string{"wireless", "aluminum", "mechanical", "custom"},
		},
		{
			"id":          "prod_2",
			"title":       "Logitech MX Master 3S Wireless Performance Mouse",
			"description": "Quiet clicks, 8K DPI any-surface tracking, MagSpeed electromagnetic scrolling wheel with ergonomic thumb rest.",
			"category":    "Mice",
			"brand":       "Logitech",
			"price":       99.99,
			"rating":      4.8,
			"in_stock":    true,
			"tags":        []string{"wireless", "ergonomic", "productivity", "office"},
		},
		{
			"id":          "prod_3",
			"title":       "Glorious GMMK Pro 75% Gasket-Mounted Mechanical Keyboard",
			"description": "Rotary encoder knob, fully modular switch sockets, vibrant RGB backlighting, black anodized case.",
			"category":    "Keyboards",
			"brand":       "Glorious",
			"price":       169.99,
			"rating":      4.7,
			"in_stock":    true,
			"tags":        []string{"mechanical", "rgb", "custom", "gasket"},
		},
		{
			"id":          "prod_4",
			"title":       "Razer DeathAdder V3 Pro Ultra-Lightweight Wireless Mouse",
			"description": "63g ultra-lightweight ergonomic esports mouse, Focus Pro 30K optical sensor, optical switches Gen-3.",
			"category":    "Mice",
			"brand":       "Razer",
			"price":       149.99,
			"rating":      4.6,
			"in_stock":    true,
			"tags":        []string{"gaming", "wireless", "ultralight", "fast"},
		},
		{
			"id":          "doc_sqlite_fts5",
			"title":       "Full-Text Search with SQLite FTS5 in Go",
			"description": "Guide to configuring virtual tables, BM25 ranking, trigram tokenizers, and porter stemmers in embedded SQLite applications.",
			"category":    "Documentation",
			"brand":       "SearchCraft",
			"price":       0.0,
			"rating":      5.0,
			"in_stock":    true,
			"tags":        []string{"sqlite", "fts5", "golang", "bm25", "database"},
		},
		{
			"id":          "doc_rrf_ranking",
			"title":       "Reciprocal Rank Fusion (RRF) for Hybrid Search",
			"description": "Mathematical explanation of Reciprocal Rank Fusion, smoothing parameter k=60, and merging disparate score distributions.",
			"category":    "Documentation",
			"brand":       "SearchCraft",
			"price":       0.0,
			"rating":      5.0,
			"in_stock":    true,
			"tags":        []string{"rrf", "hybrid", "vector", "bm25", "search"},
		},
	}

	docs := make([]storage.Document, len(sample))
	for i, item := range sample {
		docs[i] = storage.Document{
			ID:   item["id"].(string),
			Data: item,
		}
	}
	return docs
}
