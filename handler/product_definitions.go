package handler

import (
	"fastener-api/db"
	"fastener-api/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// isForeignKeyViolation 判斷是否為外鍵約束錯誤。
func isForeignKeyViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23503"
	}
	return false
}

// --- ProductCategory CRUD ---

// 新增產品類別
func CreateProductCategory(c *gin.Context) {
	var category models.ProductCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	category.CategoryCode = strings.TrimSpace(category.CategoryCode)
	category.Name = strings.TrimSpace(category.Name)

	if err := db.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product category: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// 取得所有產品類別
func GetProductCategories(c *gin.Context) {
	var categories []models.ProductCategory
	if err := db.DB.Order("category_code").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query product categories: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// 更新產品類別
func UpdateProductCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var category models.ProductCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// 只更新 category_code, name
	if err := db.DB.Model(&models.ProductCategory{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"category_code": category.CategoryCode,
			"name":          category.Name,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product category: " + err.Error()})
		return
	}

	category.ID = uint(id)
	c.JSON(http.StatusOK, category)
}

// 刪除產品類別
func DeleteProductCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// 嘗試刪除（GORM 回傳原始 pq error 可判斷 foreign key violation）
	err = db.DB.Delete(&models.ProductCategory{}, id).Error
	if err != nil {
		if isForeignKeyViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete category as it is currently in use"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product category: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product category deleted successfully"})
}

// (未來可在這裡擴充 Shape, Function, Specification 等 handler)
