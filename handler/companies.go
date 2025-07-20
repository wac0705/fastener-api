// fastener-api-main/handler/companies.go
package handler

import (
	"database/sql"
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

	// 使用 map 來方便快速查找節點
	companyMap := make(map[int]*models.Company)
	var allCompanies []*models.Company

	for rows.Next() {
		var company models.Company
		if err := rows.Scan(&company.ID, &company.Name, &company.ParentID, &company.CreatedAt, &company.UpdatedAt); err == nil {
			// 複製指標，避免迴圈中的指標問題
			node := company 
			companyMap[node.ID] = &node
			allCompanies = append(allCompanies, &node)
		}
	}
	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "讀取公司資料時發生錯誤: " + err.Error()})
		return
	}

	// 建立樹狀結構
	var rootCompanies []*models.Company
	for _, company := range allCompanies {
		if company.ParentID.Valid {
			// 如果有 parent_id，就找到它的 parent，並把自己加到 parent 的 Children 裡
			if parent, ok := companyMap[int(company.ParentID.Int64)]; ok {
				parent.Children = append(parent.Children, company)
			}
		} else {
			// 如果沒有 parent_id，代表它是根節點
			rootCompanies = append(rootCompanies, company)
		}
	}

	c.JSON(http.StatusOK, rootCompanies)
}


// --- 建立公司 (支援 parent_id) ---
func CreateCompany(c *gin.Context) {
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
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

// (UpdateCompany 和 DeleteCompany 邏輯也需要相應調整，但我們先完成查詢和建立)
