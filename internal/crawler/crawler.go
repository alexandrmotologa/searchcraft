package crawler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alexandrmotologa/searchcraft/internal/index"
	"github.com/alexandrmotologa/searchcraft/internal/storage"
)

var (
	scriptRegex = regexp.MustCompile(`(?is)<script.*?>.*?</script>`)
	styleRegex  = regexp.MustCompile(`(?is)<style.*?>.*?</style>`)
	navRegex    = regexp.MustCompile(`(?is)<nav.*?>.*?</nav>`)
	footerRegex = regexp.MustCompile(`(?is)<footer.*?>.*?</footer>`)
	tagRegex    = regexp.MustCompile(`(?s)<.*?>`)
	spaceRegex  = regexp.MustCompile(`\s+`)

	titleRegex = regexp.MustCompile(`(?i)<title>(.*?)</title>`)
	metaRegex  = regexp.MustCompile(`(?i)<meta\s+name=["']description["']\s+content=["'](.*?)["']`)
	h1Regex    = regexp.MustCompile(`(?i)<h1.*?>(.*?)</h1>`)
)

// URLSet represents standard XML sitemap format.
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

// SitemapIndex represents a sitemap index file containing child sitemaps.
type SitemapIndex struct {
	XMLName  xml.Name `xml:"sitemapindex"`
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

// Crawler manages web page fetching and HTML content extraction.
type Crawler struct {
	client *http.Client
}

// NewCrawler initializes a crawler with sensible timeouts and redirects.
func NewCrawler() *Crawler {
	return &Crawler{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Crawl executes a web crawl job starting from a sitemap or base URL.
func (c *Crawler) Crawl(ctx context.Context, eng *index.Engine, indexName string, req storage.CrawlRequest) (*storage.CrawlResult, error) {
	start := time.Now()
	maxPages := req.MaxPages
	if maxPages <= 0 {
		maxPages = 50
	}
	if maxPages > 200 {
		maxPages = 200
	}

	targetURL := strings.TrimSpace(req.URL)
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}

	// 1. Discover target URLs (either sitemap.xml or direct links)
	discoveredURLs, err := c.discoverURLs(ctx, targetURL, maxPages)
	if err != nil {
		return nil, fmt.Errorf("discover urls from %s: %w", targetURL, err)
	}

	if len(discoveredURLs) == 0 {
		discoveredURLs = []string{targetURL}
	}

	// 2. Concurrently fetch and parse HTML pages (5 concurrent workers)
	type parseResult struct {
		doc *storage.Document
		err error
	}

	resultsChan := make(chan parseResult, len(discoveredURLs))
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup

	for _, pageURL := range discoveredURLs {
		wg.Add(1)
		go func(pURL string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				resultsChan <- parseResult{err: ctx.Err()}
				return
			}

			doc, err := c.fetchAndParsePage(ctx, pURL)
			resultsChan <- parseResult{doc: doc, err: err}
		}(pageURL)
	}

	wg.Wait()
	close(resultsChan)

	var docs []storage.Document
	for res := range resultsChan {
		if res.err == nil && res.doc != nil {
			docs = append(docs, *res.doc)
		}
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no content could be extracted from %s", targetURL)
	}

	// 3. Index extracted documents
	indexed, err := eng.IndexDocuments(ctx, indexName, docs)
	if err != nil {
		return nil, fmt.Errorf("index crawled docs: %w", err)
	}

	duration := time.Since(start)
	tookMs := float64(duration.Microseconds()) / 1000.0

	return &storage.CrawlResult{
		URL:          targetURL,
		PagesCrawled: len(discoveredURLs),
		DocsIndexed:  indexed,
		TookMs:       tookMs,
	}, nil
}

func (c *Crawler) discoverURLs(ctx context.Context, base string, maxPages int) ([]string, error) {
	sitemapURL := base
	if !strings.HasSuffix(sitemapURL, ".xml") {
		// Attempt checking common sitemap location
		parsed, err := url.Parse(base)
		if err == nil {
			sitemapURL = fmt.Sprintf("%s://%s/sitemap.xml", parsed.Scheme, parsed.Host)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sitemapURL, nil)
	if err != nil {
		return []string{base}, nil
	}
	req.Header.Set("User-Agent", "SearchCraft-Crawler/1.0")

	resp, err := c.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return []string{base}, nil
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []string{base}, nil
	}

	// Try urlset
	var urlset URLSet
	if err := xml.Unmarshal(bodyBytes, &urlset); err == nil && len(urlset.URLs) > 0 {
		var list []string
		for _, u := range urlset.URLs {
			loc := strings.TrimSpace(u.Loc)
			if loc != "" {
				list = append(list, loc)
				if len(list) >= maxPages {
					break
				}
			}
		}
		return list, nil
	}

	// Try sitemapindex
	var sIndex SitemapIndex
	if err := xml.Unmarshal(bodyBytes, &sIndex); err == nil && len(sIndex.Sitemaps) > 0 {
		var list []string
		for _, sm := range sIndex.Sitemaps {
			subURLs, _ := c.discoverURLs(ctx, sm.Loc, maxPages-len(list))
			list = append(list, subURLs...)
			if len(list) >= maxPages {
				break
			}
		}
		return list, nil
	}

	return []string{base}, nil
}

func (c *Crawler) fetchAndParsePage(ctx context.Context, pageURL string) (*storage.Document, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "SearchCraft-Crawler/1.0 (Hybrid Indexing Engine)")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024)) // limit to 2MB
	if err != nil {
		return nil, err
	}

	htmlContent := string(bodyBytes)

	// Extract Title
	title := ""
	if m := titleRegex.FindStringSubmatch(htmlContent); len(m) > 1 {
		title = strings.TrimSpace(cleanHTMLText(m[1]))
	}
	if title == "" {
		if m := h1Regex.FindStringSubmatch(htmlContent); len(m) > 1 {
			title = strings.TrimSpace(cleanHTMLText(m[1]))
		}
	}
	if title == "" {
		title = pageURL
	}

	// Extract Meta Description
	desc := ""
	if m := metaRegex.FindStringSubmatch(htmlContent); len(m) > 1 {
		desc = strings.TrimSpace(cleanHTMLText(m[1]))
	}

	// Extract and clean Body Text
	cleaned := scriptRegex.ReplaceAllString(htmlContent, "")
	cleaned = styleRegex.ReplaceAllString(cleaned, "")
	cleaned = navRegex.ReplaceAllString(cleaned, "")
	cleaned = footerRegex.ReplaceAllString(cleaned, "")
	cleaned = tagRegex.ReplaceAllString(cleaned, " ")
	cleaned = cleanHTMLText(cleaned)

	if desc == "" && len(cleaned) > 20 {
		// Use first 200 characters of text
		limitLen := 200
		if len(cleaned) < limitLen {
			limitLen = len(cleaned)
		}
		desc = strings.TrimSpace(cleaned[:limitLen]) + "..."
	}

	// Category from URL path
	category := "General"
	if parsed, err := url.Parse(pageURL); err == nil {
		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) > 0 && parts[0] != "" {
			category = strings.Title(parts[0])
		}
	}

	// Document ID from URL hash
	h := sha256.Sum256([]byte(pageURL))
	docID := "page_" + hex.EncodeToString(h[:8])

	return &storage.Document{
		ID: docID,
		Data: map[string]any{
			"id":          docID,
			"title":       title,
			"description": desc,
			"content":     cleaned,
			"url":         pageURL,
			"category":    category,
		},
	}, nil
}

func cleanHTMLText(s string) string {
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	return spaceRegex.ReplaceAllString(strings.TrimSpace(s), " ")
}
