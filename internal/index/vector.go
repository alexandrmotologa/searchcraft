package index

import (
	"sort"
	"sync"

	"github.com/alexandrmotologa/searchcraft/internal/embedding"
)

// VectorMatch represents a document matched via vector similarity.
type VectorMatch struct {
	DocID       string
	CosineScore float64 // in range [-1.0, 1.0], normalized vectors typically [0.0, 1.0]
}

// VectorIndex maintains an in-memory cache of normalized document vectors for rapid cosine search.
type VectorIndex struct {
	mu      sync.RWMutex
	vectors map[string][]float32
	dim     int
}

// NewVectorIndex creates an empty vector index cache.
func NewVectorIndex(dim int) *VectorIndex {
	return &VectorIndex{
		vectors: make(map[string][]float32),
		dim:     dim,
	}
}

// SetBatch stores or updates a collection of document vectors.
func (vi *VectorIndex) SetBatch(vectors map[string][]float32) {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	for id, vec := range vectors {
		vi.vectors[id] = vec
	}
}

// Delete removes a document vector.
func (vi *VectorIndex) Delete(docID string) {
	vi.mu.Lock()
	defer vi.mu.Unlock()

	delete(vi.vectors, docID)
}

// Count returns the number of indexed vectors.
func (vi *VectorIndex) Count() int {
	vi.mu.RLock()
	defer vi.mu.RUnlock()

	return len(vi.vectors)
}

// Search calculates cosine similarity against all stored vectors and returns the top-K highest scoring.
func (vi *VectorIndex) Search(queryVec []float32, limit int) []VectorMatch {
	if len(queryVec) == 0 || limit <= 0 {
		return nil
	}

	vi.mu.RLock()
	defer vi.mu.RUnlock()

	if len(vi.vectors) == 0 {
		return nil
	}

	matches := make([]VectorMatch, 0, len(vi.vectors))
	for docID, docVec := range vi.vectors {
		// Because both queryVec and docVec are L2 normalized, dot product equals cosine similarity
		sim := float64(embedding.DotProduct(queryVec, docVec))
		// Filter out negative or zero similarity
		if sim > 0 {
			matches = append(matches, VectorMatch{
				DocID:       docID,
				CosineScore: sim,
			})
		}
	}

	// Sort descending by score
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].CosineScore > matches[j].CosineScore
	})

	if len(matches) > limit {
		matches = matches[:limit]
	}

	return matches
}
