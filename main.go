package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/wac0705/fastener-api/db"
	"github.com/wac0705/fastener-api/handler"
	"github.com/wac0705/fastener-api/middleware"
	"github.com/wac0705/fastener-api/routes"
)

func setupRoutes(app *fiber.App) {
	// Auth routes
	app.Post("/api/login", routes.Login)

	// API Group with JWT middleware protection
	api := app.Group("/api", middleware.Protected())

	// A simple welcome route to test JWT
	api.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to the protected area!")
	})

	// Company Routes
	api.Get("/companies", handler.GetCompanies)
	api.Get("/companies/tree", handler.GetCompaniesTree) // Get companies as a tree structure
	api.Get("/companies/:id", handler.GetCompany)
	api.Post("/companies", handler.CreateCompany)
	api.Put("/companies/:id", handler.UpdateCompany)
	api.Delete("/companies/:id", handler.DeleteCompany)

	// Role Routes
	api.Get("/roles", handler.GetRoles)
	api.Get("/roles/:id", handler.GetRole)
	api.Post("/roles", handler.CreateRole)
	api.Put("/roles/:id", handler.UpdateRole)
	api.Delete("/roles/:id", handler.DeleteRole)

	// Menu Routes
	api.Get("/menus", handler.GetMenus)             // Get flat list of menus
	api.Get("/menus/tree", handler.GetAllMenusTree) // NEW: Get full menu tree for admin pages
	api.Get("/menus/:id", handler.GetMenu)
	api.Post("/menus", handler.CreateMenu)
	api.Put("/menus/:id", handler.UpdateMenu)
	api.Delete("/menus/:id", handler.DeleteMenu)

	// User-specific menu route
	api.Get("/user-menus", handler.GetUserMenus) // NEW: Get menu tree for the logged-in user's sidebar

	// Role-Menu Relation Routes
	api.Get("/roles/:id/menus", handler.GetRoleMenus)
	api.Put("/roles/:id/menus", handler.UpdateRoleMenus)

	// Account Management Routes
	api.Get("/manage-accounts", handler.GetAccounts)
	api.Post("/manage-accounts", handler.CreateAccount)
	api.Put("/manage-accounts/:id", handler.UpdateAccount)
	api.Delete("/manage-accounts/:id", handler.DeleteAccount)

	// Customer Routes
	api.Get("/customers", handler.GetCustomers)
	api.Post("/customers", handler.CreateCustomer)
	api.Get("/customers/:id", handler.GetCustomer)
	api.Put("/customers/:id", handler.UpdateCustomer)
	api.Delete("/customers/:id", handler.DeleteCustomer)
	api.Get("/customers/:id/transaction-terms", handler.GetCustomerTransactionTerms)
	api.Post("/customers/:id/transaction-terms", handler.CreateCustomerTransactionTerm)
	api.Put("/customer-transaction-terms/:termId", handler.UpdateCustomerTransactionTerm)
	api.Delete("/customer-transaction-terms/:termId", handler.DeleteCustomerTransactionTerm)

	// Product Definition Routes
	api.Get("/definitions/product-categories", handler.GetProductCategories)
	api.Post("/definitions/product-categories", handler.CreateProductCategory)
	// Add other product definition routes here if needed
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using environment variables")
	}

	app := fiber.New()

	// CORS Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("FRONTEND_URL"),
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// Connect to the database
	db.ConnectDB()

	// Setup routes
	setupRoutes(app)

	// Start server
	log.Fatal(app.Listen(":3001"))
}
