// fastener-api-main/handler/companies.go
package handler

import (
	// 【修正】移除了未被使用的 "database/sql"
	"fastener-api/db"
	"fastener-api/models"
	"net/http"

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

	// 【修正】如果沒有任何公司，回傳一個空陣列而非 null
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
	// 使用 company.ParentID，因為它已經是 sql.NullInt64 型別
	err := db.Conn.QueryRow(sqlStatement, company.Name, company.ParentID).Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "建立公司失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, company)
}

// (UpdateCompany 和 DeleteCompany 函式未來需要升級以處理階層邏輯)
func GetCompanyByID(c *gin.Context) { /* ... 待升級 ... */ }
func UpdateCompany(c *gin.Context) { /* ... 待升級 ... */ }
func DeleteCompany(c *gin.Context) { /* ... 待升級 ... */ }
