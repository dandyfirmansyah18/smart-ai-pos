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
// @Summary List all products
// @Description Get all products available in inventory
// @Tags Products
// @Produce json
// @Success 200 {array} domain.Product
// @Router /products [get]
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
// @Summary Get product by SKU
// @Description Get specific product details by SKU
// @Tags Products
// @Produce json
// @Param sku path string true "Product SKU"
// @Success 200 {object} domain.Product
// @Router /products/{sku} [get]
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

// CreateProduct POST /api/products
// @Summary Create a new product
// @Description Create a new product in inventory (Admin/Warehouse only)
// @Tags Products
// @Accept json
// @Produce json
// @Param request body domain.Product true "Product details"
// @Success 201 {object} domain.Product
// @Router /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req domain.Product
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body format", "details": err.Error()})
		return
	}

	if req.SKU == "" || req.Name == "" || req.Price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sku, name and non-negative price are required"})
		return
	}

	err := h.repo.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}
