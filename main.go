package main

import (
	"fmt"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"        // ✅ 加上這行
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"  // PostgreSQL驅動

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt" // 🔒 密碼比對用
	
)

type Estimation struct {
	InquiryID     int             `json:"inquiry_id"`
	Materials     json.RawMessage `json:"materials"` // e.g., [{"code":"M8碳鋼","cost":0.5}]
	Processes     json.RawMessage `json:"processes"`
	Logistics     json.RawMessage `json:"logistics"`
	TotalCost     float64         `json:"total_cost"`
	AISuggestions float64         `json:"ai_suggestions"` // AI預測調整
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}


type UserAccount struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}




var db *sql.DB

var jwtKey = []byte("mysecretkey") // 實際可用 os.Getenv 讀環境變數

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}


func initDB() {
	connStr := os.Getenv("DATABASE_URL")

	// ✅ 自動加上 sslmode=require（若未附帶）
	if !containsSSLMode(connStr) {
		connStr += "?sslmode=require"
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
}

// 🔍 檢查連線字串是否包含 sslmode 參數
func containsSSLMode(s string) bool {
	return len(s) >= 10 &&
		(contains(s, "sslmode=require") ||
			contains(s, "sslmode=verify-full") ||
			contains(s, "sslmode=disable"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s[len(s)-len(substr):] == substr || s[len(s)-len(substr)-1:] == "&"+substr)
}


func calculateCost(est Estimation) float64 {
	// 簡化計算邏輯 (實際可擴展)
	// 抓DB材料價、物流規則等
	// e.g., matCost = 材料價 * 數量
	// procCost = 製程基價 * 工時
	// logCost = 運費/噸 * 重量 + 關稅率 * 總額
	// 貿易範例: if incoterms == "DDP", 加內陸階梯 (if 重量 > 1000, 折扣10%)
	return 100.0 + est.AISuggestions  // 範例總成本
}

func getAISuggestion() float64 {
	// 模擬AI (未來連OpenAI): 基於材料歷史預測波動
	return 0.05  // 5%調整
}

func createEstimation(c *gin.Context) {
	var est Estimation
	if err := c.BindJSON(&est); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	est.AISuggestions = getAISuggestion()
	est.TotalCost = calculateCost(est)
	// 保存到DB: INSERT INTO estimations ...
	// db.Exec("INSERT INTO estimations ...", est.InquiryID, est.Materials, ...)
	c.JSON(http.StatusOK, est)
}


func login(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	var hashedPassword string
	var role string

	err := db.QueryRow("SELECT password_hash, role FROM users WHERE username = $1", creds.Username).Scan(&hashedPassword, &role)
	if err != nil {
		// ✅ 印出錯誤詳細
		fmt.Println("查詢失敗：", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// ✅ 檢查密碼是否正確
	if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
		return
	}

	// ✅ JWT Token 建立
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: creds.Username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"role":  role,
	})
}



func getAccounts(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	rows, err := db.Query(`
		SELECT u.id, u.username, r.name, u.is_active
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		ORDER BY u.id`)
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
	role := c.GetString("role")
	if role != "admin" {
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

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hash failed"})
		return
	}

	var roleID int
	err = db.QueryRow("SELECT id FROM roles WHERE name = $1", req.Role).Scan(&roleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	_, err = db.Exec("INSERT INTO users (username, password_hash, role_id, is_active) VALUES ($1, $2, $3, TRUE)",
		req.Username, string(hashed), roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insert failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created"})
}

func updateAccount(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
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
	err := db.QueryRow("SELECT id FROM roles WHERE name = $1", req.Role).Scan(&roleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	_, err = db.Exec("UPDATE users SET role_id = $1, is_active = $2 WHERE id = $3", roleID, req.IsActive, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

func deleteAccount(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
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






// JWT 驗證中介層
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		tokenStr := authHeader[len("Bearer "):]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// 把角色和使用者名稱存到 context 裡
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}





func main() {
	initDB()
	r := gin.Default()
	
	r.Use(cors.Default()) // ✅ 允許所有來源跨域，測試或前端呼叫用

	r.GET("/health", func(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	r.POST("/api/login", login) // ✅ 加入登入路由

	// 📄 取得所有使用者（只限 admin）
	r.GET("/api/manage-accounts", authMiddleware(), getAccounts)
	// ➕ 新增使用者
	r.POST("/api/manage-accounts", authMiddleware(), createAccount)
	// ✏️ 修改使用者（含角色、啟用狀態）
	r.PUT("/api/manage-accounts/:id", authMiddleware(), updateAccount)
	// ❌ 刪除使用者
	r.DELETE("/api/manage-accounts/:id", authMiddleware(), deleteAccount)


	r.POST("/api/estimations", authMiddleware(), createEstimation)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
