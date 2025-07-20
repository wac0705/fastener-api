// fastener-api/routes/auth.go (增加日誌)
package routes

import (
	"database/sql"
	"log" // 匯入 log 套件
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
		// --- 增加日誌 ---
		log.Println("--- 收到登入請求 /api/login ---")
		// ---------------

		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("❌ 請求格式錯誤: %v", err) // 增加錯誤日誌
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
			return
		}

		// --- 增加日誌 ---
		log.Printf("收到的登入帳號: %s", req.Username)
		// ---------------

		var hashedPassword, roleName string
		err := db.QueryRow(`
			SELECT u.password_hash, r.name FROM users u
			LEFT JOIN roles r ON u.role_id = r.id
			WHERE u.username = $1 AND u.is_active = true
		`, req.Username).Scan(&hashedPassword, &roleName)

		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("⚠️ 登入失敗: 找不到帳號或帳號未啟用 - %s", req.Username) // 增加日誌
				c.JSON(http.StatusUnauthorized, gin.H{"error": "帳號或密碼錯誤"})
				return
			}
			log.Printf("❌ 資料庫查詢失敗: %v", err) // 增加錯誤日誌
			c.JSON(http.StatusInternalServerError, gin.H{"error": "資料庫查詢失敗"})
			return
		}

		// 比較雜湊後的密碼
		if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)) != nil {
			log.Printf("⚠️ 登入失敗: 密碼錯誤 - %s", req.Username) // 增加日誌
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
			log.Printf("❌ 無法產生 Token: %v", err) // 增加錯誤日誌
			c.JSON(http.StatusInternalServerError, gin.H{"error": "無法產生 Token"})
			return
		}

		log.Printf("✅ 登入成功，已產生 Token - 使用者: %s, 角色: %s", req.Username, roleName) // 增加成功日誌
		c.JSON(http.StatusOK, gin.H{"token": tokenStr, "role": roleName})
	}
}
