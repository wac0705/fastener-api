// fastener-api-main/main.go (最終修正版)
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"fastener-api/db"
	"fastener-api/handler"
	"fastener-api/middleware"
	"fastener-api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // a postgres driver
)

func main() {
	// 初始化資料庫連線
	db.Init()
	// 無條件設定 defer，因為如果 Init 失敗，程式會直接退出，不會執行到這裡
	// 如果 Init 成功，db.Conn 必定有值
	defer db.Conn.Close()

	r := gin.Default()

	// CORS 中介軟體設定
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://fastener-frontend-v2.zeabur.app", "http://localhost:3000"},
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

	// 登入路由
	r.POST("/api/login", routes.LoginHandler(db.Conn))

	// 基礎資料管理 API 群組
	definitions := r.Group("/api/definitions")
	definitions.Use(middleware.JWTAuthMiddleware())
	{
		companies := definitions.Group("/companies")
		{
			companies.POST("", handler.CreateCompany)
			companies.GET("", handler.GetCompanies)
			companies.GET("/:id", handler.GetCompanyByID)
			companies.PUT("/:id", handler.UpdateCompany)
			companies.DELETE("/:id", handler.DeleteCompany)
		}
	}

	// 帳號管理 API 群組
	accounts := r.Group("/api/manage-accounts")
	accounts.Use(middleware.JWTAuthMiddleware())
	{
		accounts.GET("", handler.GetAccounts)
		accounts.POST("", handler.CreateAccount)
		accounts.PUT("/:id", handler.UpdateAccount)
		accounts.DELETE("/:id", handler.DeleteAccount)
	}

	// --- 啟動伺服器 ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)

	// 使用 log.Fatal 包裹 r.Run 是更常見且穩健的做法
	// 如果 r.Run 回傳錯誤，程式會記錄錯誤並立即以非 0 狀態退出
	log.Fatal(r.Run(":" + port))
}
