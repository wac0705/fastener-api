// fastener-api-main/routes/auth.go
package routes

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest 定義了登入時請求的結構
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims 定義了 JWT token 中儲存的資訊
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// jwtKey 從環境變數讀取密鑰
var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// LoginHandler 處理使用者登入請求
func LoginHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
			return
		}

		var hashedPassword, roleName string
		// ⚠️ 注意：這裡我們統一使用 users 和 roles 資料表進行查詢
		err := db.QueryRow(`
			SELECT u.password_hash, r.name FROM users u
			LEFT JOIN roles r ON u.role_id = r.id
			WHERE u.username = $1 AND u.is_active = true
		`, req.Username).Scan(&hashedPassword, &roleName)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "帳號或密碼錯誤"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "資料庫查詢失敗"})
			return
		}

		// 比較雜湊後的密碼
		if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)) != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "帳號或密碼錯誤"})
			return
		}

		// 產生 JWT Token
		expiration := time.Now().Add(24 * time.Hour)
		claims := &Claims{
			Username: req.Username,
			Role:     roleName,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expiration),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString(jwtKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "無法產生 Token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": tokenStr, "role": roleName})
	}
}
