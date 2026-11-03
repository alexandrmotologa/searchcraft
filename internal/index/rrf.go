package index

import (
	"math"
	"sort"

	"github.com/alexandrmotologa/searchcraft/internal/storage"
)

const DefaultRRFK = 60.0

// CandidateScore holds the scoring components for a document candidate.
type CandidateScore struct {
	DocID       string
	BM25Rank    int     // 1-based rank, 0 if not present
	VectorRank  int     // 1-based rank, 0 if not present
	BM25Score   float64 // normalized [0.0, 1.0]
	VectorScore float64 // cosine similarity [0.0, 1.0]
	RRFScore    float64
	FinalScore  float64
}

// CombineRRF merges ranked results from BM25 and Vector search using Reciprocal Rank Fusion.
// It also computes a weighted linear score using alpha (default 0.5).
func CombineRRF(bm25Results []storage.FTSResult, vectorResults []VectorMatch, alpha float64, k float64) []CandidateScore {
	if k <= 0 {
		k = DefaultRRFK
	}
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}

	candidates := make(map[string]*CandidateScore)

	// Process BM25 rankings
	for rankIdx, res := range bm25Results {
		rank := rankIdx + 1
		cs, exists := candidates[res.DocID]
		if !exists {
			cs = &CandidateScore{DocID: res.DocID}
			candidates[res.DocID] = cs
		}
		cs.BM25Rank = rank
		cs.BM25Score = res.BM25Score
		cs.RRFScore += 1.0 / (k + float64(rank))
	}

	// Process Vector rankings
	for rankIdx, res := range vectorResults {
		rank := rankIdx + 1
		cs, exists := candidates[res.DocID]
		if !exists {
			cs = &CandidateScore{DocID: res.DocID}
			candidates[res.DocID] = cs
		}
		cs.VectorRank = rank
		cs.VectorScore = res.CosineScore
		cs.RRFScore += 1.0 / (k + float64(rank))
	}

	// Compute composite final score
	result := make([]CandidateScore, 0, len(candidates))
	for _, cs := range candidates {
		// Blended hybrid score: alpha * bm25 + (1 - alpha) * vector
		cs.FinalScore = alpha*cs.BM25Score + (1.0-alpha)*cs.VectorScore
		// Round scores for clean presentation
		cs.FinalScore = math.Round(cs.FinalScore*10000) / 10000
		cs.RRFScore = math.Round(cs.RRFScore*100000) / 100000
		cs.BM25Score = math.Round(cs.BM25Score*10000) / 10000
		cs.VectorScore = math.Round(cs.VectorScore*10000) / 10000

		result = append(result, *cs)
	}

	// Sort primary by FinalScore descending, then by RRFScore
	sort.Slice(result, func(i, j int) bool {
		if result[i].FinalScore != result[j].FinalScore {
			return result[i].FinalScore > result[j].FinalScore
		}
		return result[i].RRFScore > result[j].RRFScore
	})

	return result
}
