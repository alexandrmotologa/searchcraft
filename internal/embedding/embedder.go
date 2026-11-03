package embedding

import (
	"context"
	"math"
)

// Embedder defines the common interface for text embedding models.
type Embedder interface {
	Name() string
	Dimension() int
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
}

// Normalize normalizes a vector to unit length (L2 norm = 1.0).
func Normalize(v []float32) []float32 {
	var sum float64
	for _, val := range v {
		sum += float64(val) * float64(val)
	}
	norm := float32(math.Sqrt(sum))
	if norm == 0 {
		return v
	}

	result := make([]float32, len(v))
	invNorm := 1.0 / norm
	for i, val := range v {
		result[i] = val * invNorm
	}
	return result
}

// CosineSimilarity computes the cosine similarity between two float32 vectors.
// If both vectors are normalized, it is equal to their dot product.
func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dot, sumA, sumB float64
	for i := 0; i < len(a); i++ {
		va := float64(a[i])
		vb := float64(b[i])
		dot += va * vb
		sumA += va * va
		sumB += vb * vb
	}

	denominator := math.Sqrt(sumA) * math.Sqrt(sumB)
	if denominator == 0 {
		return 0
	}

	return float32(dot / denominator)
}

// DotProduct calculates the raw dot product of two vectors (assumes normalized).
func DotProduct(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	var dot float32
	for i := 0; i < len(a); i++ {
		dot += a[i] * b[i]
	}
	return dot
}
