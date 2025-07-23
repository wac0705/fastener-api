package routes

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest 定義登入請求格式
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims 定義 JWT token 資料
type Claims struct {
	Username  string `json:"username"`
	Role      string `json:"role"`
	CompanyID int    `json:"company_id"`
	jwt.RegisteredClaims
}

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// LoginHandler 處理登入邏輯
func LoginHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("--- 收到登入請求 /api/login ---")

		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("❌ 請求格式錯誤: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
			return
		}

		log.Printf("收到的登入帳號: %s", req.Username)

		var hashedPassword, roleName string
		var companyID int
		err := db.QueryRow(`
			SELECT u.password_hash, r.name, u.tenant_id FROM users u
			LEFT JOIN roles r ON u.role_id = r.id
			WHERE u.username = $1 AND u.is_active = true
		`, req.Username).Scan(&hashedPassword, &roleName, &companyID)

		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("⚠️ 登入失敗: 找不到帳號或帳號未啟用 - %s", req.Username)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "帳號或密碼錯誤"})
				return
			}
			log.Printf("❌ 資料庫查詢失敗: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "資料庫查詢失敗"})
			return
		}

		if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)) != nil {
			log.Printf("⚠️ 登入失敗: 密碼錯誤 - %s", req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "帳號或密碼錯誤"})
			return
		}

		now := time.Now()
		expiration := now.Add(24 * time.Hour)
		claims := &Claims{
			Username:  req.Username,
			Role:      roleName,
			CompanyID: companyID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expiration),
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, err := token.SignedString(jwtKey)
		if err != nil {
			log.Printf("❌ 無法產生 Token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "無法產生 Token"})
			return
		}

		log.Printf("✅ 登入成功，已產生 Token - 使用者: %s, 角色: %s, 公司: %d", req.Username, roleName, companyID)
		c.JSON(http.StatusOK, gin.H{
			"token":      tokenStr,
			"role":       roleName,
			"company_id": companyID,
		})
	}
}
