package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"fastener-api/db"
	"fastener-api/models"
)

// 權限驗證（多層級管理員分權）
func getRoleAndCompanyID(c *gin.Context) (string, int, bool) {
	role, exists := c.Get("role")
	companyID, exists2 := c.Get("company_id")
	if !exists || !exists2 {
		return "", 0, false
	}
	roleStr, _ := role.(string)
	companyIDInt := 0
	switch v := companyID.(type) {
	case int:
		companyIDInt = v
	case int64:
		companyIDInt = int(v)
	case float64:
		companyIDInt = int(v)
	case string:
		companyIDInt, _ = strconv.Atoi(v)
	}
	return roleStr, companyIDInt, true
}

// 查詢帳號列表
func GetAccounts(c *gin.Context) {
	role, companyID, ok := getRoleAndCompanyID(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "權限不足"})
		return
	}

	var rows *sql.Rows
	var err error

	if role == "superadmin" {
		// 查全部
		rows, err = db.Conn.Query(`
			SELECT u.id, u.username, r.name as role, u.is_active, u.tenant_id, c.name as company_name
			FROM users u
			LEFT JOIN roles r ON u.role_id = r.id
			LEFT JOIN companies c ON u.tenant_id = c.id
			ORDER BY u.id
		`)
	} else if role == "company_admin" {
		// 只查自己公司與子公司（假設子公司ID已知，需進階可用遞迴CTE查所有下層）
		subCompanyIDs := getDescendantCompanyIDs(companyID)
		placeholder := strings.Repeat("?,", len(subCompanyIDs))
		placeholder = strings.TrimRight(placeholder, ",")
		args := make([]interface{}, len(subCompanyIDs))
		for i, v := range subCompanyIDs {
			args[i] = v
		}
		query := `
			SELECT u.id, u.username, r.name as role, u.is_active, u.tenant_id, c.name as company_name
			FROM users u
			LEFT JOIN roles r ON u.role_id = r.id
			LEFT JOIN companies c ON u.tenant_id = c.id
			WHERE u.tenant_id IN (` + placeholder + `)
			ORDER BY u.id
		`
		rows, err = db.Conn.Query(query, args...)
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "權限不足"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "資料庫查詢失敗: " + err.Error()})
		return
	}
	defer rows.Close()

	var accounts []models.UserAccount
	for rows.Next() {
		var acc models.UserAccount
		if err := rows.Scan(&acc.ID, &acc.Username, &acc.Role, &acc.IsActive, &acc.CompanyID, &acc.CompanyName); err == nil {
			accounts = append(accounts, acc)
		}
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

	// company_admin 只能建立自己或子公司帳號
	if role == "company_admin" {
		allowedIDs := getDescendantCompanyIDs(companyID)
		isAllowed := false
		for _, id := range allowedIDs {
			if id == req.CompanyID {
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

	var roleID int
	err = db.Conn.QueryRow("SELECT id FROM roles WHERE name = $1", req.Role).Scan(&roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢角色失敗: " + err.Error()})
		return
	}

	_, err = db.Conn.Exec(`
		INSERT INTO users (username, password_hash, role_id, tenant_id, is_active)
		VALUES ($1, $2, $3, $4, true)
	`, req.Username, string(hashed), roleID, req.CompanyID)
	if err != nil {
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

	// company_admin 只能更新自己或子公司帳號
	if role == "company_admin" {
		allowedIDs := getDescendantCompanyIDs(companyID)
		isAllowed := false
		for _, id := range allowedIDs {
			if id == req.CompanyID {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "無法異動此公司帳號"})
			return
		}
	}

	var roleID int
	err := db.Conn.QueryRow("SELECT id FROM roles WHERE name = $1", req.Role).Scan(&roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢角色失敗: " + err.Error()})
		return
	}

	_, err = db.Conn.Exec(`
		UPDATE users SET role_id = $1, is_active = $2, tenant_id = $3 WHERE id = $4
	`, roleID, req.IsActive, req.CompanyID, id)
	if err != nil {
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

	// company_admin 只能刪自己公司/子公司帳號
	// 這裡請補一個查詢該帳號的 tenant_id，並比對可控公司ID
	if role == "company_admin" {
		var targetCompanyID int
		err := db.Conn.QueryRow("SELECT tenant_id FROM users WHERE id = $1", id).Scan(&targetCompanyID)
		if err != nil {
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

	_, err := db.Conn.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除帳號失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "帳號刪除成功"})
}

// ==== 這個 function 請改成用你自己的公司遞迴查詢方式 ====
func getDescendantCompanyIDs(companyID int) []int {
	// TODO: 實作真正的遞迴查詢資料庫
	// 目前只是自己公司
	return []int{companyID}
}
