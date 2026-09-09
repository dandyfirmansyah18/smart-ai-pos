package inbound_test

import (
	"testing"

	"github.com/pos-backend/internal/ports/inbound"
)

func TestCheckoutRequestStruct(t *testing.T) {
	req := inbound.CheckoutRequest{
		IdempotencyKey: "test-idem-key-123",
		Items: []inbound.CheckoutItem{
			{
				SKU:      "SKU-001",
				Quantity: 2,
			},
		},
	}

	if req.IdempotencyKey != "test-idem-key-123" {
		t.Errorf("expected test-idem-key-123, got %s", req.IdempotencyKey)
	}

	if len(req.Items) != 1 || req.Items[0].SKU != "SKU-001" {
		t.Errorf("expected item SKU-001, got %v", req.Items)
	}
}
