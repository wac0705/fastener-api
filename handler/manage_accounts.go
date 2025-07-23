package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"fastener-api/db"
	"fastener-api/models"
)

// 權限驗證（多層級管理員分權）
func getRoleAndCompanyID(c *gin.Context) (string, uint, bool) {
	role, exists := c.Get("role")
	companyID, exists2 := c.Get("company_id")
	if !exists || !exists2 {
		return "", 0, false
	}
	roleStr, _ := role.(string)
	var companyIDUint uint
	switch v := companyID.(type) {
	case int:
		companyIDUint = uint(v)
	case int64:
		companyIDUint = uint(v)
	case float64:
		companyIDUint = uint(v)
	case uint:
		companyIDUint = v
	case string:
		id, _ := strconv.Atoi(v)
		companyIDUint = uint(id)
	}
	return roleStr, companyIDUint, true
}

// 查詢帳號列表
func GetAccounts(c *gin.Context) {
	role, companyID, ok := getRoleAndCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "權限不足"})
		return
	}

	var accounts []models.UserAccount

	if role == "superadmin" {
		// 查全部
		db.DB.Raw(`
			SELECT u.id, u.username, r.name as role, u.is_active, u.tenant_id as company_id, c.name as company_name
			FROM users u
			LEFT JOIN roles r ON u.role_id = r.id
			LEFT JOIN companies c ON u.tenant_id = c.id
			ORDER BY u.id
		`).Scan(&accounts)
	} else if role == "company_admin" {
		subCompanyIDs := getDescendantCompanyIDs(companyID)
		if len(subCompanyIDs) == 0 {
			c.JSON(http.StatusOK, []models.UserAccount{})
			return
		}
		var idStrings []string
		for _, v := range subCompanyIDs {
			idStrings = append(idStrings, strconv.Itoa(int(v)))
		}
		placeholders := strings.Join(idStrings, ",")
		query := `
			SELECT u.id, u.username, r.name as role, u.is_active, u.tenant_id as company_id, c.name as company_name
			FROM users u
			LEFT JOIN roles r ON u.role_id = r.id
			LEFT JOIN companies c ON u.tenant_id = c.id
			WHERE u.tenant_id IN (` + placeholders + `)
			ORDER BY u.id
		`
		db.DB.Raw(query).Scan(&accounts)
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "權限不足"})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

// 新增帳號
func CreateAccount(c *gin.Context) {
	role, companyID, ok := getRoleAndCompanyID(c)
	if !ok || (role != "superadmin" && role != "company_admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "權限不足"})
		return
	}

	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	if role == "company_admin" {
		allowedIDs := getDescendantCompanyIDs(companyID)
		isAllowed := false
		for _, id := range allowedIDs {
			if id == uint(req.CompanyID) {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "無法在此公司建立帳號"})
			return
		}
	}

	if req.Username == "" || req.Password == "" || req.Role == "" || req.CompanyID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所有欄位皆為必填"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密碼加密失敗"})
		return
	}

	var roleID uint
	if err := db.DB.Raw("SELECT id FROM roles WHERE name = ?", req.Role).Scan(&roleID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢角色失敗: " + err.Error()})
		return
	}

	user := models.User{
		Username:  req.Username,
		Password:  string(hashed),
		RoleID:    roleID,
		CompanyID: uint(req.CompanyID),
		IsActive:  true,
	}
	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "新增帳號失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "帳號建立成功"})
}

// 修改帳號
func UpdateAccount(c *gin.Context) {
	role, companyID, ok := getRoleAndCompanyID(c)
	if !ok || (role != "superadmin" && role != "company_admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "權限不足"})
		return
	}

	id := c.Param("id")
	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	if role == "company_admin" {
		allowedIDs := getDescendantCompanyIDs(companyID)
		isAllowed := false
		for _, id := range allowedIDs {
			if id == uint(req.CompanyID) {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "無法異動此公司帳號"})
			return
		}
	}

	var roleID uint
	if err := db.DB.Raw("SELECT id FROM roles WHERE name = ?", req.Role).Scan(&roleID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢角色失敗: " + err.Error()})
		return
	}

	if err := db.DB.Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"role_id":   roleID,
			"is_active": req.IsActive,
			"tenant_id": req.CompanyID,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新帳號失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "帳號更新成功"})
}

// 刪除帳號
func DeleteAccount(c *gin.Context) {
	role, companyID, ok := getRoleAndCompanyID(c)
	if !ok || (role != "superadmin" && role != "company_admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "權限不足"})
		return
	}

	id := c.Param("id")
	if id == "1" {
		c.JSON(http.StatusForbidden, gin.H{"error": "無法刪除主要的管理員帳號"})
		return
	}
	if _, err := strconv.Atoi(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的帳號 ID"})
		return
	}

	if role == "company_admin" {
		var targetCompanyID uint
		if err := db.DB.Raw("SELECT tenant_id FROM users WHERE id = ?", id).Scan(&targetCompanyID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢帳號公司失敗"})
			return
		}
		allowedIDs := getDescendantCompanyIDs(companyID)
		isAllowed := false
		for _, allowed := range allowedIDs {
			if allowed == targetCompanyID {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "無法刪除此公司帳號"})
			return
		}
	}

	if err := db.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除帳號失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "帳號刪除成功"})
}

// 重設帳號密碼
func ResetPassword(c *gin.Context) {
	role, companyID, ok := getRoleAndCompanyID(c)
	if !ok || (role != "superadmin" && role != "company_admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "權限不足"})
		return
	}
	id := c.Param("id")
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "請輸入新密碼"})
		return
	}
	// company_admin 只能改自己公司/子公司帳號
	if role == "company_admin" {
		var targetCompanyID uint
		if err := db.DB.Raw("SELECT tenant_id FROM users WHERE id = ?", id).Scan(&targetCompanyID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢帳號公司失敗"})
			return
		}
		allowedIDs := getDescendantCompanyIDs(companyID)
		isAllowed := false
		for _, allowed := range allowedIDs {
			if allowed == targetCompanyID {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "無法修改此公司帳號密碼"})
			return
		}
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密碼加密失敗"})
		return
	}
	if err := db.DB.Model(&models.User{}).Where("id = ?", id).
		Update("password_hash", string(hashed)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密碼更新失敗"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密碼已重設"})
}

// ==== 查詢自己+所有下層公司ID（支援 RECURSIVE，GORM Raw）====
func getDescendantCompanyIDs(companyID uint) []uint {
	var ids []uint
	rows, err := db.DB.Raw(`
		WITH RECURSIVE company_tree AS (
			SELECT id FROM companies WHERE id = ?
			UNION ALL
			SELECT c.id FROM companies c
			JOIN company_tree t ON c.parent_id = t.id
		)
		SELECT id FROM company_tree
	`, companyID).Rows()
	if err != nil {
		return []uint{companyID}
	}
	defer rows.Close()
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		ids = append(ids, companyID)
	}
	return ids
}
