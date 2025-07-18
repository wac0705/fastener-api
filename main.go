// fastener-api-main/main.go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // a postgres driver

	"fastener-api/db"
	"fastener-api/handler"
	"fastener-api/middleware"
	"fastener-api/routes"
)

func main() {
	db.Init()
	defer db.Conn.Close()

	r := gin.Default()

	// --- 修正 CORS 設定 ---
	// 我們需要更詳細的設定來處理所有請求方法
	r.Use(cors.New(cors.Config{
		// 這裡建議填寫您的前端網址，用 '*' 是為了開發方便
		AllowOrigins:     []string{"https://fastener-frontend-v2.zeabur.app", "http://localhost:3000"},
		// 必須明確允許所有前端會用到的 HTTP 方法
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		// 允許前端攜帶的 Header
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// --- 路由設定 ---
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	api := r.Group("/api")
	{
		api.POST("/login", routes.LoginHandler(db.Conn))

		accounts := api.Group("/manage-accounts")
		accounts.Use(middleware.JWTAuthMiddleware())
		{
			// Gin 路由會自動處理結尾斜線的問題，所以這裡不需要改
			accounts.GET("", handler.GetAccounts)
			accounts.POST("", handler.CreateAccount)
			accounts.PUT("/:id", handler.UpdateAccount)
			accounts.DELETE("/:id", handler.DeleteAccount)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("🚀 Server starting on port %s", port)
	r.Run(":" + port)
}
