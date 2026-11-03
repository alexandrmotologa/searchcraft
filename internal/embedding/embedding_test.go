package embedding

import (
	"context"
	"math"
	"testing"
)

func TestFastLocalEmbedder(t *testing.T) {
	emb := NewFastLocalEmbedder(256)
	ctx := context.Background()

	vec1, err := emb.Embed(ctx, "Mechanical Keyboard Wireless")
	if err != nil {
		t.Fatalf("embed text 1: %v", err)
	}
	if len(vec1) != 256 {
		t.Errorf("expected dim 256, got %d", len(vec1))
	}

	// Verify unit norm
	var sum float64
	for _, v := range vec1 {
		sum += float64(v) * float64(v)
	}
	norm := math.Sqrt(sum)
	if math.Abs(norm-1.0) > 1e-4 {
		t.Errorf("expected unit norm approx 1.0, got %f", norm)
	}

	// Verify determinism
	vec1Duplicate, _ := emb.Embed(ctx, "mechanical keyboard wireless")
	simIdentical := CosineSimilarity(vec1, vec1Duplicate)
	if simIdentical < 0.999 {
		t.Errorf("expected identical similarity ~1.0, got %f", simIdentical)
	}

	// Verify semantic/lexical similarity
	vecRelated, _ := emb.Embed(ctx, "Wireless Mechanical Keyboards")
	simRelated := CosineSimilarity(vec1, vecRelated)
	if simRelated < 0.75 {
		t.Errorf("expected related similarity > 0.75, got %f", simRelated)
	}

	// Verify unrelated text has significantly lower similarity
	vecUnrelated, _ := emb.Embed(ctx, "Banana strawberry smoothie recipes")
	simUnrelated := CosineSimilarity(vec1, vecUnrelated)
	if simUnrelated > 0.4 {
		t.Errorf("expected unrelated similarity < 0.4, got %f", simUnrelated)
	}
}

func TestCosineSimilarity(t *testing.T) {
	v1 := []float32{1.0, 0.0, 0.0}
	v2 := []float32{1.0, 0.0, 0.0}
	v3 := []float32{0.0, 1.0, 0.0}

	if sim := CosineSimilarity(v1, v2); math.Abs(float64(sim)-1.0) > 1e-5 {
		t.Errorf("expected 1.0, got %f", sim)
	}
	if sim := CosineSimilarity(v1, v3); sim != 0 {
		t.Errorf("expected 0.0, got %f", sim)
	}
}
