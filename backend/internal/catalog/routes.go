package catalog

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan semua rute katalog persis seperti router
// lama: baca (kategori/produk/varian/price-tier) di grup auth, mutasi
// dan impor di grup admin.
func RegisterRoutes(auth, admin *gin.RouterGroup, h *Handler) {
	auth.GET("/categories", h.ListCategories)
	auth.GET("/products", h.ListProducts)
	auth.GET("/products/:id", h.GetProduct)
	auth.GET("/products/:id/variants", h.ListVariants)
	auth.GET("/products/:id/price-tiers", h.ListPriceTiers)
	auth.GET("/products/:id/variants/:variantId/price-tiers", h.ListPriceTiers)

	admin.POST("/categories", h.CreateCategory)
	admin.DELETE("/categories/:id", h.DeleteCategory)
	admin.POST("/products", h.CreateProduct)
	admin.POST("/products/import", h.ImportProducts)
	admin.PUT("/products/:id", h.UpdateProduct)
	admin.POST("/products/:id/variants", h.CreateVariant)
	admin.PUT("/products/:id/variants/:variantId", h.UpdateVariant)
	admin.DELETE("/products/:id/variants/:variantId", h.DeleteVariant)
	admin.POST("/products/:id/price-tiers", h.CreatePriceTier)
	admin.PUT("/products/:id/price-tiers/:tierId", h.UpdatePriceTier)
	admin.DELETE("/products/:id/price-tiers/:tierId", h.DeletePriceTier)
	admin.POST("/products/:id/variants/:variantId/price-tiers", h.CreatePriceTier)
	admin.PUT("/products/:id/variants/:variantId/price-tiers/:tierId", h.UpdatePriceTier)
	admin.DELETE("/products/:id/variants/:variantId/price-tiers/:tierId", h.DeletePriceTier)
	admin.PATCH("/products/:id/active", h.SetProductActive)
	admin.DELETE("/products/:id", h.DeleteProduct)
}
