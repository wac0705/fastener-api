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
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq" // a postgres driver
)

func main() {
	db.Init()
	defer db.Conn.Close()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://fastener-frontend-v2.zeabur.app", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
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

		// --- 基礎資料管理 API 群組 ---
		definitions := api.Group("/definitions")
		definitions.Use(middleware.JWTAuthMiddleware()) // 所有基礎資料 API 都需要驗證
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
		accounts := api.Group("/manage-accounts")
		accounts.Use(middleware.JWTAuthMiddleware())
		{
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
