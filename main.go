// fastener-api-main/main.go (完整重寫修正版)
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
	// 初始化資料庫連線。如果失敗，db.Init() 內部會呼叫 log.Fatal() 結束程式。
	db.Init()
	// 確保在 main 函式結束時關閉資料庫連線。
	defer db.Conn.Close()

	// 建立一個預設的 Gin 引擎
	r := gin.Default()

	// 設定 CORS (跨來源資源共用) 中介軟體
	r.Use(cors.New(cors.Config{
		// 允許的前端來源網域
		AllowOrigins:     []string{"https://fastener-frontend-v2.zeabur.app", "http://localhost:3000"},
		// 允許的 HTTP 方法
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		// 允許的請求標頭
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		// 允許前端讀取的標頭
		ExposeHeaders:    []string{"Content-Length"},
		// 允許傳送 cookies
		AllowCredentials: true,
		// pre-flight 請求的快取時間
		MaxAge:           12 * time.Hour,
	}))

	// --- 路由設定 ---

	// 健康檢查路由，用於確認服務是否正常運行
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	// 登入路由，不需 JWT 驗證
	r.POST("/api/login", routes.LoginHandler(db.Conn))

	// --- 需要 JWT 驗證的路由群組 ---

	// 基礎資料管理 API 群組
	definitions := r.Group("/api/definitions")
	definitions.Use(middleware.JWTAuthMiddleware())
	{
		// 公司管理的路由
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

	// 從環境變數讀取埠號，若無則使用預設值 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)

	// 啟動 HTTP 伺服器並監聽指定埠號
	// 如果啟動失敗，log.Fatal 會印出錯誤訊息並結束程式
	log.Fatal(r.Run(":" + port))
}
