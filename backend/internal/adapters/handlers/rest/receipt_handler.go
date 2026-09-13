package rest

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/inbound"
)

type ReceiptHandler struct {
	useCase inbound.ReceiptUseCase
}

func NewReceiptHandler(useCase inbound.ReceiptUseCase) *ReceiptHandler {
	return &ReceiptHandler{useCase: useCase}
}

// ScanReceipt POST /api/receipts/scan
func (h *ReceiptHandler) ScanReceipt(c *gin.Context) {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		fileHeader, err = c.FormFile("file")
	}
	if err != nil {
		fileHeader, err = c.FormFile("receipt")
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "receipt image file is required in multipart form (field 'image' or 'file')"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read uploaded file bytes"})
		return
	}

	audit, err := h.useCase.ScanReceipt(c.Request.Context(), fileHeader.Filename, imageBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI OCR receipt scanning failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, audit)
}

// ListAudits GET /api/receipts/audits
func (h *ReceiptHandler) ListAudits(c *gin.Context) {
	audits, err := h.useCase.ListAudits(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve receipt audits"})
		return
	}
	if audits == nil {
		audits = []domain.ReceiptAudit{}
	}
	c.JSON(http.StatusOK, audits)
}
