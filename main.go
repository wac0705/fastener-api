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
	defer db.Conn.Close()

	// 建立 Gin 引擎
	r := gin.Default()

	// 設定 CORS 中介軟體
	// *** 修正點：已將此區塊中所有看不見的特殊空白字元替換為標準空格 ***
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

	// 登入路由 - 這是關鍵！
	r.POST("/api/login", routes.LoginHandler(db.Conn))

	// --- 需要 JWT 驗證的路由群組 ---

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
	
	log.Fatal(r.Run(":" + port))
}
