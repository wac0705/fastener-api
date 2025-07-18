// handler/manage_accounts.go
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wujohnny/fastener-api/db"
	"github.com/wujohnny/fastener-api/models"
	"golang.org/x/crypto/bcrypt"
)

// GET /api/manage-accounts
func GetAccounts(c *gin.Context) {
	rows, err := db.Conn.Query(`SELECT id, username, password, role, is_active FROM accounts ORDER BY id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var acc models.Account
		if err := rows.Scan(&acc.ID, &acc.Username, &acc.Password, &acc.Role, &acc.IsActive); err == nil {
			accounts = append(accounts, acc)
		}
	}

	c.JSON(http.StatusOK, accounts)
}

// POST /api/manage-accounts
func CreateAccount(c *gin.Context) {
	var acc models.Account
	if err := json.NewDecoder(c.Request.Body).Decode(&acc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(acc.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hashing failed"})
		return
	}

	_, err = db.Conn.Exec(`
		INSERT INTO accounts (username, password, role, is_active)
		VALUES ($1, $2, $3, $4)
	`, acc.Username, string(hashed), acc.Role, true)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insert failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created"})
}

// PUT /api/manage-accounts/:id
func UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Role     string `json:"role"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	_, err := db.Conn.Exec(`
		UPDATE accounts SET role = $1, is_active = $2 WHERE id = $3
	`, req.Role, req.IsActive, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

// DELETE /api/manage-accounts/:id
func DeleteAccount(c *gin.Context) {
	id := c.Param("id")

	// 🔐 取得 JWT Claims（你需確認 middleware 有 c.Set("role", role)）
	roleVal, exists := c.Get("role")
	if !exists || roleVal != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can delete accounts"})
		return
	}

	// 🛑 不允許刪除 admin 自己
	if id == "1" { // 如果你 admin 帳號固定是 id=1
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete admin account"})
		return
	}

	_, err := db.Conn.Exec(`DELETE FROM accounts WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
