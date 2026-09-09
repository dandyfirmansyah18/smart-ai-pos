package rest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pos-backend/internal/domain"
	"github.com/pos-backend/internal/ports/outbound"
)

type ProductHandler struct {
	repo outbound.ProductRepository
}

func NewProductHandler(repo outbound.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

// ListProducts GET /api/products
func (h *ProductHandler) ListProducts(c *gin.Context) {
	products, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve products"})
		return
	}
	if products == nil {
		products = []domain.Product{}
	}
	c.JSON(http.StatusOK, products)
}

// GetProductBySKU GET /api/products/:sku
func (h *ProductHandler) GetProductBySKU(c *gin.Context) {
	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku parameter is required"})
		return
	}

	product, err := h.repo.GetBySKU(c.Request.Context(), sku)
	if err != nil {
		if errors.Is(err, domain.ErrProductNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve product"})
		return
	}

	c.JSON(http.StatusOK, product)
}
