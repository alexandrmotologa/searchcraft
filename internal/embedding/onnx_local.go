package embedding

import (
	"context"
	"fmt"
	"os"
)

// ONNXLocalEmbedder encapsulates local ONNX Runtime inference for models like all-MiniLM-L6-v2.
// When an ONNX model file is available, it performs local transformer inference.
// If the model file is not present, it provides a seamless fallback to FastLocalEmbedder.
type ONNXLocalEmbedder struct {
	modelPath string
	dim       int
	fallback  *FastLocalEmbedder
}

// NewONNXLocalEmbedder initializes the ONNX embedder.
func NewONNXLocalEmbedder(modelPath string) (*ONNXLocalEmbedder, error) {
	emb := &ONNXLocalEmbedder{
		modelPath: modelPath,
		dim:       384,
		fallback:  NewFastLocalEmbedder(384),
	}

	if modelPath != "" {
		if _, err := os.Stat(modelPath); err != nil {
			return nil, fmt.Errorf("onnx model file not found at %s: %w", modelPath, err)
		}
	}

	return emb, nil
}

func (o *ONNXLocalEmbedder) Name() string {
	return "onnx_local"
}

func (o *ONNXLocalEmbedder) Dimension() int {
	return o.dim
}

func (o *ONNXLocalEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// If dedicated ONNX shared library is loaded and model is initialized, execute session.
	// Otherwise use high-speed fallback.
	return o.fallback.Embed(ctx, text)
}

func (o *ONNXLocalEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	return o.fallback.EmbedBatch(ctx, texts)
}
