// fastener-api-main/handler/companies.go
package handler

import (
	"database/sql"
	"fastener-api/db"
	"fastener-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// --- 建立公司 ---
func CreateCompany(c *gin.Context) {
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式: " + err.Error()})
		return
	}

	sqlStatement := `INSERT INTO companies (name) VALUES ($1) RETURNING id, created_at, updated_at`
	err := db.Conn.QueryRow(sqlStatement, company.Name).Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "建立公司失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, company)
}

// --- 查詢所有公司 ---
func GetCompanies(c *gin.Context) {
	rows, err := db.Conn.Query("SELECT id, name, created_at, updated_at FROM companies ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢公司資料失敗: " + err.Error()})
		return
	}
	defer rows.Close()

	var companies []models.Company
	for rows.Next() {
		var company models.Company
		if err := rows.Scan(&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt); err == nil {
			companies = append(companies, company)
		}
	}
	c.JSON(http.StatusOK, companies)
}

// --- 查詢單一公司 ---
func GetCompanyByID(c *gin.Context) {
	id := c.Param("id")

	var company models.Company
	sqlStatement := `SELECT id, name, created_at, updated_at FROM companies WHERE id = $1`
	err := db.Conn.QueryRow(sqlStatement, id).Scan(&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "找不到指定的公司"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢公司資料失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, company)
}

// --- 更新公司 ---
func UpdateCompany(c *gin.Context) {
	id := c.Param("id")
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式: " + err.Error()})
		return
	}

	sqlStatement := `UPDATE companies SET name = $1, updated_at = NOW() WHERE id = $2 RETURNING id, name, created_at, updated_at`
	err := db.Conn.QueryRow(sqlStatement, company.Name, id).Scan(&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "找不到要更新的公司"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新公司失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, company)
}

// --- 刪除公司 ---
func DeleteCompany(c *gin.Context) {
	id := c.Param("id")

	// 增加保護機制，不允許刪除 ID 為 1 的公司 (通常是總部)
	if id == "1" {
		c.JSON(http.StatusForbidden, gin.H{"error": "無法刪除主要的總部公司"})
		return
	}

	sqlStatement := `DELETE FROM companies WHERE id = $1`
	res, err := db.Conn.Exec(sqlStatement, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除公司失敗: " + err.Error()})
		return
	}

	count, err := res.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "無法獲取影響的行數: " + err.Error()})
		return
	}

	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到要刪除的公司"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "公司刪除成功"})
}
