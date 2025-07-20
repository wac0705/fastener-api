// fastener-api-main/handler/product_definitions.go
package handler

import (
	"database/sql"
	"fastener-api/db"
	"fastener-api/models"
	"net/http"
	"strconv" // 匯入 strconv 套件用於字串轉換

	"github.com/gin-gonic/gin"
)

// --- 產品主類別 (ProductCategory) CRUD ---

// CreateProductCategory 建立新的產品類別
func CreateProductCategory(c *gin.Context) {
	var category models.ProductCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式: " + err.Error()})
		return
	}

	sqlStatement := `INSERT INTO product_categories (category_code, name) VALUES ($1, $2) RETURNING id`
	err := db.Conn.QueryRow(sqlStatement, category.CategoryCode, category.Name).Scan(&category.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "建立產品類別失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// GetProductCategories 查詢所有產品類別
func GetProductCategories(c *gin.Context) {
	rows, err := db.Conn.Query("SELECT id, category_code, name FROM product_categories ORDER BY category_code")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢產品類別失敗: " + err.Error()})
		return
	}
	defer rows.Close()

	// 【修正】確保 categories 是一個 non-nil 的 slice，即使沒有資料也是一個空 slice
	categories := make([]models.ProductCategory, 0)
	for rows.Next() {
		var category models.ProductCategory
		if err := rows.Scan(&category.ID, &category.CategoryCode, &category.Name); err == nil {
			categories = append(categories, category)
		}
	}

	// 檢查迴圈中是否有錯誤
	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "讀取資料時發生錯誤: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// UpdateProductCategory 更新產品類別
func UpdateProductCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID 格式"})
		return
	}

	var category models.ProductCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式: " + err.Error()})
		return
	}

	sqlStatement := `UPDATE product_categories SET category_code = $1, name = $2 WHERE id = $3`
	res, err := db.Conn.Exec(sqlStatement, category.CategoryCode, category.Name, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新產品類別失敗: " + err.Error()})
		return
	}

	count, _ := res.RowsAffected()
    if count == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "找不到要更新的產品類別"})
        return
    }

	category.ID = id
	c.JSON(http.StatusOK, category)
}

// DeleteProductCategory 刪除產品類別
func DeleteProductCategory(c *gin.Context) {
	id := c.Param("id")
	sqlStatement := `DELETE FROM product_categories WHERE id = $1`
	res, err := db.Conn.Exec(sqlStatement, id)
	if err != nil {
		// 【修正】處理外鍵約束錯誤，提供更友善的提示
		if isForeignKeyViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "無法刪除，此類別可能已被其他資料關聯使用"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除產品類別失敗: " + err.Error()})
		return
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到要刪除的產品類別"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "產品類別刪除成功"})
}

// (未來我們將在此檔案中繼續添加 Shape, Function, Specification 的 CRUD 函式)
