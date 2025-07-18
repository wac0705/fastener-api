// fastener-api-main/handler/manage_accounts.go
package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wujohnny/fastener-api/db"
	"github.com/wujohnny/fastener-api/models"
	"golang.org/x/crypto/bcrypt"
)

// permissionDenied 是一個輔助函式，用於回傳權限不足的錯誤
func permissionDenied(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": "權限不足"})
}

// GetAccounts 處理獲取所有帳號的請求
func GetAccounts(c *gin.Context) {
	// 從中介軟體取得角色資訊
	userRole, _ := c.Get("role")
	if userRole != "admin" {
		permissionDenied(c)
		return
	}

	// ⚠️ 注意：統一使用 users 和 roles 資料表
	rows, err := db.Conn.Query(`
		SELECT u.id, u.username, r.name as role, u.is_active 
		FROM users u 
		LEFT JOIN roles r ON u.role_id = r.id 
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
		if err := rows.Scan(&acc.ID, &acc.Username, &acc.Role, &acc.IsActive); err == nil {
			accounts = append(accounts, acc)
		}
	}

	c.JSON(http.StatusOK, accounts)
}

// CreateAccount 處理新增帳號的請求
func CreateAccount(c *gin.Context) {
	userRole, _ := c.Get("role")
	if userRole != "admin" {
		permissionDenied(c)
		return
	}

	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// 密碼加密
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密碼加密失敗"})
		return
	}

	// 根據角色名稱找到 role_id
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

	// 插入新使用者到 users 資料表
	_, err = db.Conn.Exec(`
		INSERT INTO users (username, password_hash, role_id, is_active)
		VALUES ($1, $2, $3, true)
	`, req.Username, string(hashed), roleID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "新增帳號失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "帳號建立成功"})
}

// UpdateAccount 處理更新帳號的請求
func UpdateAccount(c *gin.Context) {
	userRole, _ := c.Get("role")
	if userRole != "admin" {
		permissionDenied(c)
		return
	}

	id := c.Param("id")
	var req models.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
		return
	}

	// 根據角色名稱找到 role_id
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

	_, err = db.Conn.Exec(`
		UPDATE users SET role_id = $1, is_active = $2 WHERE id = $3
	`, roleID, req.IsActive, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新帳號失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "帳號更新成功"})
}

// DeleteAccount 處理刪除帳號的請求
func DeleteAccount(c *gin.Context) {
	userRole, _ := c.Get("role")
	if userRole != "admin" {
		permissionDenied(c)
		return
	}

	id := c.Param("id")

	// 🛑 保護措施：不允許刪除 ID 為 1 的帳號 (通常是超級管理員)
	if id == "1" {
		c.JSON(http.StatusForbidden, gin.H{"error": "無法刪除主要的管理員帳號"})
		return
	}

	_, err := db.Conn.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除帳號失敗: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "帳號刪除成功"})
}
