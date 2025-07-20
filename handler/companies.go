// fastener-api-main/handler/companies.go
package handler

import (
	"fastener-api/db"
	"fastener-api/models"
	"net/http"
	"strconv"

	// 【核心修正點】修正了 gin 的匯入路徑
	"github.com/gin-gonic/gin" 
)

// --- 查詢所有公司 (以樹狀結構回傳) ---
func GetCompanies(c *gin.Context) {
	rows, err := db.Conn.Query("SELECT id, name, parent_id, created_at, updated_at FROM companies ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢公司資料失敗: " + err.Error()})
		return
	}
	defer rows.Close()

	companyMap := make(map[int]*models.Company)
	var allCompanies []*models.Company

	for rows.Next() {
		var company models.Company
		if err := rows.Scan(&company.ID, &company.Name, &company.ParentID, &company.CreatedAt, &company.UpdatedAt); err == nil {
			node := company 
			companyMap[node.ID] = &node
			allCompanies = append(allCompanies, &node)
		}
	}
	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "讀取公司資料時發生錯誤: " + err.Error()})
		return
	}

	var rootCompanies []*models.Company
	for _, company := range allCompanies {
		if company.ParentID.Valid {
			if parent, ok := companyMap[int(company.ParentID.Int64)]; ok {
				parent.Children = append(parent.Children, company)
			}
		} else {
			rootCompanies = append(rootCompanies, company)
		}
	}

	if rootCompanies == nil {
		rootCompanies = make([]*models.Company, 0)
	}

	c.JSON(http.StatusOK, rootCompanies)
}


// --- 建立公司 (支援 parent_id) ---
func CreateCompany(c *gin.Context) {
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式: " + err.Error()})
		return
	}

	sqlStatement := `INSERT INTO companies (name, parent_id) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	err := db.Conn.QueryRow(sqlStatement, company.Name, company.ParentID).Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "建立公司失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, company)
}

// --- 查詢單一公司 ---
func GetCompanyByID(c *gin.Context) {
	id := c.Param("id")
	var company models.Company
	err := db.Conn.QueryRow(
		"SELECT id, name, parent_id, created_at, updated_at FROM companies WHERE id = $1", id,
	).Scan(&company.ID, &company.Name, &company.ParentID, &company.CreatedAt, &company.UpdatedAt)
	
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到指定的公司"})
		return
	}
	c.JSON(http.StatusOK, company)
}

// --- 更新公司 ---
func UpdateCompany(c *gin.Context) {
	idStr := c.Param("id")
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	sqlStatement := `UPDATE companies SET name = $1, parent_id = $2, updated_at = NOW() WHERE id = $3`
	res, err := db.Conn.Exec(sqlStatement, company.Name, company.ParentID, idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新公司失敗: " + err.Error()})
		return
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到要更新的公司"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "公司更新成功"})
}

// --- 刪除公司 ---
func DeleteCompany(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	if id == 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "無法刪除 ID 為 1 的根公司"})
		return
	}

	sqlStatement := `DELETE FROM companies WHERE id = $1`
	res, err := db.Conn.Exec(sqlStatement, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除公司失敗: " + err.Error()})
		return
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到要刪除的公司"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "公司刪除成功"})
}
