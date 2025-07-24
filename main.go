package main

import (
	"log"
	"net/http"
	"os"

	"fastener-api/db"
	"fastener-api/handler"
	"fastener-api/middleware"
	"fastener-api/models"
	"fastener-api/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加載 .env 檔案
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ 無法加載 .env 檔案 (可能在生產環境中)")
	}

	// 設置 Gin 模式
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		log.Println("🚧 Gin 運行在 Debug 模式，請在生產環境中設置 GIN_MODE=release")
	}

	// 初始化資料庫連接
	db.Init()

	// 自動遷移資料庫模型
	// 確保所有需要的模型都在這裡
	err = db.DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Company{},
		&models.Customer{},
		&models.CustomerTransactionTerm{},
		&models.Inquiry{},
		&models.InquiryItem{},
		&models.Estimation{},
		&models.Quotation{},
		&models.Material{},
		&models.Process{},
		&models.Port{},
		&models.LogisticsRule{},
		&models.QuotationTemplate{},
		&models.ProductCategory{},
		&models.ProductShape{},
		&models.ProductFunction{},
		&models.ProductSpecification{},
		&models.CategoryShapeRelation{},
		&models.CategoryFunctionRelation{},
		&models.ShapeSpecRelation{},
		&models.FunctionSpecRelation{},
		&models.Menu{},
		&models.RoleMenuRelation{}, // <-- 在這裡添加 RoleMenuRelation 模型
		&models.Review{},
	)
	if err != nil {
		log.Fatalf("❌ GORM AutoMigrate 錯誤:%v", err)
	}
	log.Println("✅ 資料庫模型自動遷移完成")

	// 初始化 Gin 路由器
	r := gin.Default()

	// CORS 配置 (允許所有來源，僅用於開發，生產環境請限制)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true") // 如果需要支持憑證
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 註冊公共路由 (無需身份驗證)
	routes.SetupAuthRoutes(r)

	// 身份驗證中間件 (對所有受保護路由生效)
	authorized := r.Group("/")
	authorized.Use(middleware.AuthMiddleware())
	{
		// 帳戶管理
		authorized.GET("/users", handler.GetUsers)
		authorized.GET("/users/:id", handler.GetUserByID)
		authorized.POST("/users", handler.CreateUser)
		authorized.PATCH("/users/:id", handler.UpdateUser)
		authorized.DELETE("/users/:id", handler.DeleteUser)

		// 角色管理
		authorized.GET("/roles", handler.GetRoles)
		authorized.GET("/roles/:id", handler.GetRoleByID)
		authorized.POST("/roles", handler.CreateRole)
		authorized.PATCH("/roles/:id", handler.UpdateRole)
		authorized.DELETE("/roles/:id", handler.DeleteRole)

		// 公司管理
		authorized.GET("/companies", handler.GetCompanies)
		authorized.GET("/companies/:id", handler.GetCompanyByID)
		authorized.POST("/companies", handler.CreateCompany)
		authorized.PATCH("/companies/:id", handler.UpdateCompany)
		authorized.DELETE("/companies/:id", handler.DeleteCompany)
		authorized.GET("/companies/tree", handler.GetCompaniesTree) // 新增的公司樹狀結構

		// 客戶管理
		authorized.GET("/customers", handler.GetCustomers)
		authorized.GET("/customers/:id", handler.GetCustomerByID)
		authorized.POST("/customers", handler.CreateCustomer)
		authorized.PATCH("/customers/:id", handler.UpdateCustomer)
		authorized.DELETE("/customers/:id", handler.DeleteCustomer)
		authorized.POST("/customer-transaction-terms", handler.CreateCustomerTransactionTerm)
		authorized.PATCH("/customer-transaction-terms/:id", handler.UpdateCustomerTransactionTerm)
		authorized.DELETE("/customer-transaction-terms/:id", handler.DeleteCustomerTransactionTerm)


		// 產品定義管理 (暫時只有 Category，未來擴展 Shape, Function, Specification)
		authorized.GET("/product-categories", handler.GetProductCategories)
		authorized.GET("/product-categories/:id", handler.GetProductCategoryByID)
		authorized.POST("/product-categories", handler.CreateProductCategory)
		authorized.PATCH("/product-categories/:id", handler.UpdateProductCategory)
		authorized.DELETE("/product-categories/:id", handler.DeleteProductCategory)

		// 菜單管理 (新增路由)
		authorized.GET("/menus", handler.GetMenus)
		authorized.GET("/menus/:id", handler.GetMenuByID)
		authorized.POST("/menus", handler.CreateMenu)
		authorized.PATCH("/menus/:id", handler.UpdateMenu)
		authorized.DELETE("/menus/:id", handler.DeleteMenu)
		authorized.GET("/roles/:roleID/menus", handler.GetMenusByRoleID) // 新增：根據角色 ID 獲取菜單列表

		// 角色菜單關聯管理 (新增路由)
		authorized.GET("/role-menus", handler.GetRoleMenus)
		authorized.POST("/role-menus", handler.CreateRoleMenu)
		authorized.DELETE("/role-menus", handler.DeleteRoleMenu) // 通常用於刪除特定關聯，請確保 handler 中的邏輯正確

		// 其他現有路由...

	}

	// 啟動 Gin 伺服器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 默認埠號
	}
	log.Printf("🚀 伺服器啟動於 :%s 埠", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ 伺服器啟動失敗: %v", err)
	}
}
