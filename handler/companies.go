package handler

import (
	"fastener-api/db"
	"fastener-api/models"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
)

// --- 查詢所有公司 (樹狀結構) ---
func GetCompanies(c *gin.Context) {
	var companies []models.Company
	if err := db.DB.Order("name").Find(&companies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢公司資料失敗: " + err.Error()})
		return
	}

	// ID: pointer mapping for building tree
	companyMap := make(map[uint]*models.Company)
	for i := range companies {
		companies[i].Children = []*models.Company{} // 初始化 children
		companyMap[companies[i].ID] = &companies[i]
	}
	var rootCompanies []*models.Company
	for i := range companies {
		if companies[i].ParentID != nil {
			parent, ok := companyMap[*companies[i].ParentID]
			if ok {
				parent.Children = append(parent.Children, &companies[i])
			}
		} else {
			rootCompanies = append(rootCompanies, &companies[i])
		}
	}
	if rootCompanies == nil {
		rootCompanies = make([]*models.Company, 0)
	}
	c.JSON(http.StatusOK, rootCompanies)
}

// --- 建立公司 ---
func CreateCompany(c *gin.Context) {
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式: " + err.Error()})
		return
	}
	if err := db.DB.Create(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "建立公司失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, company)
}

// --- 查詢單一公司 ---
func GetCompanyByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的公司 ID"})
		return
	}
	var company models.Company
	if err := db.DB.First(&company, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到指定的公司"})
		return
	}
	c.JSON(http.StatusOK, company)
}

// --- 更新公司 ---
func UpdateCompany(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的公司 ID"})
		return
	}
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式: " + err.Error()})
		return
	}
	// GORM 的 Save 會根據主鍵有無決定 insert 或 update
	company.ID = uint(id)
	if err := db.DB.Model(&models.Company{}).Where("id = ?", id).Updates(company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新公司失敗: " + err.Error()})
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
	if err := db.DB.Delete(&models.Company{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除公司失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "公司刪除成功"})
}
