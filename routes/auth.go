package routes

import (
	"fastener-api/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	CompanyID uint   `json:"company_id"`
	jwt.RegisteredClaims
}

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// LoginHandler 處理登入邏輯 (GORM ORM)
func LoginHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("--- 收到登入請求 /api/login ---")

		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("❌ 請求格式錯誤: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
			return
		}

		log.Printf("收到的登入帳號: %s", req.Username)

		// 查詢 user 基本資訊
		var user models.User
		if err := db.Where("username = ? AND is_active = true", req.Username).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("⚠️ 登入失敗: 找不到帳號或帳號未啟用 - %s", req.Username)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "帳號或密碼錯誤"})
				return
			}
			log.Printf("❌ 資料庫查詢失敗: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "資料庫查詢失敗"})
			return
		}

		// 查詢角色名稱（role name）
		var roleName string
		db.Raw("SELECT name FROM roles WHERE id = ?", user.RoleID).Scan(&roleName)

		// 驗證密碼 (用 PasswordHash)
		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
			log.Printf("⚠️ 登入失敗: 密碼錯誤 - %s", req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "帳號或密碼錯誤"})
			return
		}

		now := time.Now()
		expiration := now.Add(24 * time.Hour)
		claims := &Claims{
			Username:  req.Username,
			Role:      roleName,
			CompanyID: user.CompanyID,
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

		log.Printf("✅ 登入成功，已產生 Token - 使用者: %s, 角色: %s, 公司: %d", req.Username, roleName, user.CompanyID)
		c.JSON(http.StatusOK, gin.H{
			"token":      tokenStr,
			"role":       roleName,
			"company_id": user.CompanyID,
		})
	}
}
