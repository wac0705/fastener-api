// fastener-api-main/main.go
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
	"github.comcom/gin-gonic/gin"
	_ "github.com/lib/pq" // a postgres driver
)

func main() {
	// 初始化資料庫連線
	db.Init()
	defer db.Conn.Close()

	// 建立 Gin 引擎
	r := gin.Default()

	// 設定 CORS 中介軟體
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://fastener-frontend-v2.zeabur.app", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// --- 路由設定 ---

	// 健康檢查路由 (不需驗證)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	// 建立 /api 路由群組
	api := r.Group("/api")
	{
		// 登入路由 (不需驗證)
		api.POST("/login", routes.LoginHandler(db.Conn))

		// --- 基礎資料管理 API 群組 (需要 JWT 驗證) ---
		definitions := api.Group("/definitions")
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

			// 客戶管理的路由
			customers := definitions.Group("/customers")
			{
				customers.POST("", handler.CreateCustomer)
				customers.GET("", handler.GetCustomers)
				customers.GET("/:id", handler.GetCustomerByID)
				customers.PUT("/:id", handler.UpdateCustomer)
				customers.DELETE("/:id", handler.DeleteCustomer)
			}

			// 產品類別管理的路由
			categories := definitions.Group("/product-categories")
			{
				categories.POST("", handler.CreateProductCategory)
				categories.GET("", handler.GetProductCategories)
				categories.PUT("/:id", handler.UpdateProductCategory)
				categories.DELETE("/:id", handler.DeleteProductCategory)
			}
		}

		// --- 帳號管理 API 群組 (需要 JWT 驗證) ---
		accounts := api.Group("/manage-accounts")
		accounts.Use(middleware.JWTAuthMiddleware())
		{
			accounts.GET("", handler.GetAccounts)
			accounts.POST("", handler.CreateAccount)
			accounts.PUT("/:id", handler.UpdateAccount)
			accounts.DELETE("/:id", handler.DeleteAccount)
		}
	}

	// --- 啟動伺服器 ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	
	// 使用 log.Fatal 可以在伺服器啟動失敗時記錄錯誤並退出程式
	log.Fatal(r.Run(":" + port))
}
