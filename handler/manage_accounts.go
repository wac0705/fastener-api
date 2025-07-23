package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"fastener-api/db"
	"fastener-api/models"
)

// 權限驗證（僅 admin 可進行）
func checkAdminPermission(c *gin.Context) bool {
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "權限不足"})
		return false
	}
	return true
}

// 查詢帳號列表
func GetAccounts(c *gin.Context) {
	if !checkAdminPermission(c) {
		return
	}

	rows, err := db.Conn.Query(`
		SELECT u.id, u.username, r.name as role, u.is_active, u.tenant_id, c.name as company_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		LEFT JOIN companies c ON u.tenant_id = c.id
		ORDER BY u.id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "資料庫查詢失敗: " + err.Error()})
		return
	}
	defer rows.Close()

	var accounts []models.UserAccount
	for rows.Next() {
		var acc models.UserAccount
		// 新增 company_id、company_name 欄位
		if err := rows.Scan(&acc.ID, &acc.Username, &acc.Role, &acc.IsActive, &acc.CompanyID, &acc.CompanyName); err == nil {
			accounts = append(accounts, acc)
		}
	}
	c.JSON(http.StatusOK, accounts)
}

// 新增帳號
func CreateAccount(c *gin.Context) {
	if !checkAdminPermission(c) {
		return
	}

	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// 必填檢查
	if req.Username == "" || req.Password == "" || req.Role == "" || req.CompanyID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "所有欄位皆為必填"})
		return
	}

	// 密碼加密
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密碼加密失敗"})
		return
	}

	// 找 role_id
	var roleID int
	err = db.Conn.QueryRow("SELECT id FROM roles WHERE name = $1", req.Role).Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效的角色名稱"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢角色失敗: " + err.Error()})
		return
	}

	// 插入新使用者
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
	if !checkAdminPermission(c) {
		return
	}

	id := c.Param("id")
	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// role 必填
	var roleID int
	err := db.Conn.QueryRow("SELECT id FROM roles WHERE name = $1", req.Role).Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效的角色名稱"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢角色失敗: " + err.Error()})
		return
	}

	// 允許更新公司
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
	if !checkAdminPermission(c) {
		return
	}

	id := c.Param("id")
	// 防呆: 不可刪掉 ID=1 的主帳號
	if id == "1" {
		c.JSON(http.StatusForbidden, gin.H{"error": "無法刪除主要的管理員帳號"})
		return
	}
	// 確認 id 有轉成數字
	if _, err := strconv.Atoi(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的帳號 ID"})
		return
	}

	_, err := db.Conn.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除帳號失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "帳號刪除成功"})
}
