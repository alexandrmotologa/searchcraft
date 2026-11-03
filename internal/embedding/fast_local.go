package embedding

import (
	"context"
	"hash/fnv"
	"math"
	"strings"
	"unicode"
)

const DefaultFastLocalDim = 256

// FastLocalEmbedder implements a deterministic subword and character n-gram feature hashing vectorizer.
// It requires zero external model downloads or API keys, runs in under 10 microseconds, and produces
// 256-dimensional normalized vectors with strong morphological and lexical similarity properties.
type FastLocalEmbedder struct {
	dimension int
}

// NewFastLocalEmbedder creates an embedder with the specified dimension (default 256).
func NewFastLocalEmbedder(dim ...int) *FastLocalEmbedder {
	d := DefaultFastLocalDim
	if len(dim) > 0 && dim[0] > 0 {
		d = dim[0]
	}
	return &FastLocalEmbedder{dimension: d}
}

func (e *FastLocalEmbedder) Name() string {
	return "fast_local"
}

func (e *FastLocalEmbedder) Dimension() int {
	return e.dimension
}

func (e *FastLocalEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	return e.embedText(text), nil
}

func (e *FastLocalEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		results[i] = e.embedText(text)
	}
	return results, nil
}

func (e *FastLocalEmbedder) embedText(text string) []float32 {
	vector := make([]float32, e.dimension)
	cleaned := strings.ToLower(strings.TrimSpace(text))
	if cleaned == "" {
		return vector
	}

	// Tokenize words
	words := strings.FieldsFunc(cleaned, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	counts := make(map[uint32]float32)

	for _, w := range words {
		if len(w) == 0 {
			continue
		}

		// Full word feature
		hWord := hashString(w)
		counts[hWord] += 2.0

		// Character n-grams (3-grams and 4-grams)
		runes := []rune(w)
		bounded := "^" + string(runes) + "$"
		bRunes := []rune(bounded)

		// 3-grams
		for i := 0; i+3 <= len(bRunes); i++ {
			gram := string(bRunes[i : i+3])
			counts[hashString(gram)] += 1.0
		}

		// 4-grams
		for i := 0; i+4 <= len(bRunes); i++ {
			gram := string(bRunes[i : i+4])
			counts[hashString(gram)] += 1.2
		}
	}

	// Project features into dimension with signed hash (feature hashing trick)
	dimU32 := uint32(e.dimension)
	for h, weight := range counts {
		bucket := h % dimU32
		// Sign bit determined by bit 31
		sign := float32(1.0)
		if (h & (1 << 31)) != 0 {
			sign = -1.0
		}
		// Sublinear term frequency damping: 1 + log(tf)
		tf := float32(1.0 + math.Log(float64(weight)))
		vector[bucket] += sign * tf
	}

	return Normalize(vector)
}

func hashString(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
