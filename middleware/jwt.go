package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims 定義了 JWT token 中儲存的資訊
type Claims struct {
	Username  string `json:"username"`
	Role      string `json:"role"`
	CompanyID int    `json:"company_id"`
	jwt.RegisteredClaims
}

// jwtKey 從環境變數讀取密鑰
var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// JWTAuthMiddleware 是一個 Gin 中介軟體，用於驗證 JWT
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少 Authorization Header"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader { // 如果沒有 "Bearer " 前綴
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization Header 格式錯誤"})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 已過期"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "無效的 Token"})
			}
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "無效的 Token"})
			c.Abort()
			return
		}

		// 將驗證後的使用者資訊存入 context，供後續 handler 使用
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("company_id", claims.CompanyID)

		c.Next()
	}
}
