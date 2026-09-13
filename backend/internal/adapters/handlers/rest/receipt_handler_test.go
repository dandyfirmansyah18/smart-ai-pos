package rest_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pos-backend/config"
	"github.com/pos-backend/internal/adapters/handlers/rest"
	"github.com/pos-backend/internal/adapters/infrastructure/vision"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/inbound"
	"github.com/pos-backend/internal/ports/outbound"
)

func TestReceiptHandler_ScanReceipt(t *testing.T) {
	mockProductRepo := outbound.NewMockProductRepository()
	mockOrderRepo := outbound.NewMockOrderRepository()
	mockLockService := outbound.NewMockLockService()

	cfg := &config.Config{Port: "8080", Env: "test"}
	visionClient := vision.NewVisionClientImpl(cfg)
	receiptUseCase := inbound.NewReceiptUseCaseImpl(visionClient, nil)

	orderUseCase := inbound.NewOrderUseCaseImpl(nil, mockProductRepo, mockOrderRepo, mockLockService)
	server := rest.NewServer(cfg, mockProductRepo, orderUseCase, nil, receiptUseCase)

	// Create multipart form payload with fake image file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", "sample_receipt.jpg")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte("fake-jpeg-image-bytes"))
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/api/receipts/scan", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d, body: %s", w.Code, w.Body.String())
	}

	var audit domain.ReceiptAudit
	if err := json.Unmarshal(w.Body.Bytes(), &audit); err != nil {
		t.Fatalf("failed to unmarshal ReceiptAudit JSON: %v", err)
	}

	if audit.MerchantName == "" {
		t.Errorf("expected merchant_name in OCR audit")
	}
	if audit.TotalAmount <= 0 {
		t.Errorf("expected positive total_amount, got %.2f", audit.TotalAmount)
	}
}
