// fastener-api-main/handler/product_definitions.go
package handler

import (
	"fastener-api/db"
	"fastener-api/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq" // Import the postgres driver to identify specific errors
)

// isForeignKeyViolation is a helper function to check for foreign key constraint errors.
func isForeignKeyViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		// 23503 is the PostgreSQL error code for foreign_key_violation
		return pqErr.Code == "23503"
	}
	return false
}

// --- ProductCategory CRUD ---

// CreateProductCategory creates a new product category.
func CreateProductCategory(c *gin.Context) {
	var category models.ProductCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	category.CategoryCode = strings.TrimSpace(category.CategoryCode)
	category.Name = strings.TrimSpace(category.Name)

	sqlStatement := `INSERT INTO product_categories (category_code, name) VALUES ($1, $2) RETURNING id`
	err := db.Conn.QueryRow(sqlStatement, category.CategoryCode, category.Name).Scan(&category.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product category: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// GetProductCategories retrieves all product categories.
func GetProductCategories(c *gin.Context) {
	rows, err := db.Conn.Query("SELECT id, category_code, name FROM product_categories ORDER BY category_code")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query product categories: " + err.Error()})
		return
	}
	defer rows.Close()

	categories := make([]models.ProductCategory, 0)
	for rows.Next() {
		var category models.ProductCategory
		if err := rows.Scan(&category.ID, &category.CategoryCode, &category.Name); err == nil {
			categories = append(categories, category)
		}
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading product category data: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// UpdateProductCategory updates an existing product category.
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

	sqlStatement := `UPDATE product_categories SET category_code = $1, name = $2 WHERE id = $3`
	res, err := db.Conn.Exec(sqlStatement, category.CategoryCode, category.Name, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product category: " + err.Error()})
		return
	}

	count, _ := res.RowsAffected()
    if count == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product category to update not found"})
        return
    }

	category.ID = id
	c.JSON(http.StatusOK, category)
}

// DeleteProductCategory deletes a product category.
func DeleteProductCategory(c *gin.Context) {
	id := c.Param("id")
	sqlStatement := `DELETE FROM product_categories WHERE id = $1`
	res, err := db.Conn.Exec(sqlStatement, id)
	if err != nil {
		if isForeignKeyViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete category as it is currently in use"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product category: " + err.Error()})
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product category to delete not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product category deleted successfully"})
}

// (Future handlers for Shape, Function, Specification will be added here)
