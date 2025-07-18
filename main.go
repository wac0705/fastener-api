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

// ... 剩下的程式碼和之前一樣，不用變 ...

func main() {
	// 初始化資料庫連線
	db.Init()
	// 在程式結束時關閉資料庫連線
	defer db.Conn.Close()

	// 建立 Gin 引擎
	r := gin.Default()

	// 設定 CORS 中介軟體
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 在生產環境建議指定前端網址
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// --- 路由設定 ---

	// 健康檢查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	// API 路由群組
	api := r.Group("/api")
	{
		// 登入路由，不需要驗證
		api.POST("/login", routes.LoginHandler(db.Conn))

		// 帳號管理路由，需要 JWT 驗證
		accounts := api.Group("/manage-accounts")
		accounts.Use(middleware.JWTAuthMiddleware())
		{
			accounts.GET("/", handler.GetAccounts)
			accounts.POST("/", handler.CreateAccount)
			accounts.PUT("/:id", handler.UpdateAccount)
			accounts.DELETE("/:id", handler.DeleteAccount)
		}
	}

	// 讀取 PORT 環境變數，若無則使用 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("🚀 Server starting on port %s", port)
	r.Run(":" + port)
}
