package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Estimation struct {
	InquiryID     int             `json:"inquiry_id"`
	Materials     json.RawMessage `json:"materials"`
	Processes     json.RawMessage `json:"processes"`
	Logistics     json.RawMessage `json:"logistics"`
	TotalCost     float64         `json:"total_cost"`
	AISuggestions float64         `json:"ai_suggestions"`
}

type UserAccount struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var (
	db     *sql.DB
	jwtKey = []byte("mysecretkey")
)

func initDB() {
	dsn := os.Getenv("DATABASE_URL")
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("DB unreachable:", err)
	}
	log.Println("✅ Connected to PostgreSQL")
}

func calculateCost(est Estimation) float64 {
	return 100.0 + est.AISuggestions
}

func getAISuggestion() float64 {
	return 0.05
}

func createEstimation(c *gin.Context) {
	var est Estimation
	if err := c.ShouldBindJSON(&est); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	est.AISuggestions = getAISuggestion()
	est.TotalCost = calculateCost(est)

	c.JSON(http.StatusOK, est)
}

func login(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	var hashedPassword, roleName string
	err := db.QueryRow(`
		SELECT u.password_hash, r.name FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1
	`, creds.Username).Scan(&hashedPassword, &roleName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
		return
	}

	expiration := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: creds.Username,
		Role:     roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenStr, "role": roleName})
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}
		tokenStr := authHeader[len("Bearer ") : len(authHeader)]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func getAccounts(c *gin.Context) {
	if c.GetString("role") != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	rows, err := db.Query(`SELECT u.id, u.username, r.name, u.is_active FROM users u LEFT JOIN roles r ON u.role_id = r.id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var users []UserAccount
	for rows.Next() {
		var u UserAccount
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.IsActive); err == nil {
			users = append(users, u)
		}
	}
	c.JSON(http.StatusOK, users)
}

func createAccount(c *gin.Context) {
	if c.GetString("role") != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	var roleID int
	if err := db.QueryRow("SELECT id FROM roles WHERE name = $1", req.Role).Scan(&roleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}
	_, err := db.Exec("INSERT INTO users (username, password_hash, role_id, is_active) VALUES ($1, $2, $3, true)", req.Username, string(hash), roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insert failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User created"})
}

func updateAccount(c *gin.Context) {
	if c.GetString("role") != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	id := c.Param("id")
	var req struct {
		Role     string `json:"role"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	var roleID int
	if err := db.QueryRow("SELECT id FROM roles WHERE name = $1", req.Role).Scan(&roleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}
	_, err := db.Exec("UPDATE users SET role_id = $1, is_active = $2 WHERE id = $3", roleID, req.IsActive, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

func deleteAccount(c *gin.Context) {
	if c.GetString("role") != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func main() {
	initDB()
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 或限制特定前端網址
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
   		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
    		ExposeHeaders:    []string{"Content-Length"},
    		AllowCredentials: true,
    		MaxAge:           12 * time.Hour,
	}))


	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})
	r.POST("/api/login", login)
	r.GET("/api/manage-accounts", authMiddleware(), getAccounts)
	r.POST("/api/manage-accounts", authMiddleware(), createAccount)
	r.PUT("/api/manage-accounts/:id", authMiddleware(), updateAccount)
	r.DELETE("/api/manage-accounts/:id", authMiddleware(), deleteAccount)
	r.POST("/api/estimations", authMiddleware(), createEstimation)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
