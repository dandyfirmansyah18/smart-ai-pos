package vision_test

import (
	"context"
	"testing"

	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/adapters/infrastructure/vision"
)

func TestVisionClientImpl_FallbackParser(t *testing.T) {
	cfg := &config.Config{}
	client := vision.NewVisionClientImpl(cfg)
	ctx := context.Background()

	mockImageBytes := []byte("fake-jpeg-image-bytes")

	result, err := client.AnalyzeReceipt(ctx, mockImageBytes)
	if err != nil {
		t.Fatalf("expected no error analyzing receipt, got %v", err)
	}

	if result.MerchantName == "" {
		t.Errorf("expected merchant name to be populated")
	}
	if len(result.Items) == 0 {
		t.Errorf("expected items to be parsed")
	}
	if result.TotalAmount <= 0 {
		t.Errorf("expected positive total amount, got %.2f", result.TotalAmount)
	}
}
