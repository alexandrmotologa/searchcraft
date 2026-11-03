package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RemoteConfig configures an external API embedder.
type RemoteConfig struct {
	Provider string // "openai", "ollama", "voyage", "custom"
	BaseURL  string // e.g. "https://api.openai.com/v1" or "http://localhost:11434/api"
	APIKey   string
	Model    string // e.g. "text-embedding-3-small", "nomic-embed-text"
	Dim      int    // expected output dimension
}

// RemoteEmbedder calls external embedding APIs.
type RemoteEmbedder struct {
	config RemoteConfig
	client *http.Client
}

// NewRemoteEmbedder creates a client for external embedding APIs.
func NewRemoteEmbedder(cfg RemoteConfig) *RemoteEmbedder {
	if cfg.BaseURL == "" {
		switch cfg.Provider {
		case "openai":
			cfg.BaseURL = "https://api.openai.com/v1"
			if cfg.Model == "" {
				cfg.Model = "text-embedding-3-small"
			}
			if cfg.Dim <= 0 {
				cfg.Dim = 1536
			}
		case "ollama":
			cfg.BaseURL = "http://localhost:11434/api"
			if cfg.Model == "" {
				cfg.Model = "nomic-embed-text"
			}
			if cfg.Dim <= 0 {
				cfg.Dim = 768
			}
		case "voyage":
			cfg.BaseURL = "https://api.voyageai.com/v1"
			if cfg.Model == "" {
				cfg.Model = "voyage-3-lite"
			}
			if cfg.Dim <= 0 {
				cfg.Dim = 512
			}
		}
	}

	return &RemoteEmbedder{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (r *RemoteEmbedder) Name() string {
	return r.config.Provider
}

func (r *RemoteEmbedder) Dimension() int {
	return r.config.Dim
}

func (r *RemoteEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	batch, err := r.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(batch) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return batch[0], nil
}

func (r *RemoteEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	if r.config.Provider == "ollama" {
		return r.embedOllama(ctx, texts)
	}

	return r.embedOpenAIFormat(ctx, texts)
}

func (r *RemoteEmbedder) embedOpenAIFormat(ctx context.Context, texts []string) ([][]float32, error) {
	url := fmt.Sprintf("%s/embeddings", r.config.BaseURL)

	payload := map[string]any{
		"model": r.config.Model,
		"input": texts,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if r.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+r.config.APIKey)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("post remote embedding: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read embedding resp: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding api returned %d: %s", resp.StatusCode, string(respBytes))
	}

	var parsed struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal embedding resp: %w", err)
	}

	results := make([][]float32, len(texts))
	for _, item := range parsed.Data {
		if item.Index >= 0 && item.Index < len(results) {
			results[item.Index] = Normalize(item.Embedding)
		}
	}

	return results, nil
}

func (r *RemoteEmbedder) embedOllama(ctx context.Context, texts []string) ([][]float32, error) {
	url := fmt.Sprintf("%s/embeddings", r.config.BaseURL)
	results := make([][]float32, len(texts))

	for i, t := range texts {
		payload := map[string]any{
			"model":  r.config.Model,
			"prompt": t,
		}
		body, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := r.client.Do(req)
		if err != nil {
			return nil, err
		}

		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(respBytes))
		}

		var parsed struct {
			Embedding []float32 `json:"embedding"`
		}
		if err := json.Unmarshal(respBytes, &parsed); err != nil {
			return nil, err
		}
		results[i] = Normalize(parsed.Embedding)
	}

	return results, nil
}
