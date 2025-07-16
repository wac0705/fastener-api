package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"        // ✅ 加上這行
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"  // PostgreSQL驅動

	"github.com/golang-jwt/jwt/v5"
	"time"
	
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

var users = []User{
	{Username: "admin", Password: "123456", Role: "admin"},
	{Username: "sales", Password: "123456", Role: "sales"},
	{Username: "engineer", Password: "123456", Role: "engineer"},
	{Username: "logistics", Password: "123456", Role: "logistics"},
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
	connStr := os.Getenv("DATABASE_URL")  // Zeabur env var
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
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

	// 找出對應的帳號
	var user *User
	for _, u := range users {
		if u.Username == creds.Username && u.Password == creds.Password {
			user = &u
			break
		}
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: user.Username,
		Role:     user.Role,
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
		"role":  user.Role,
	})
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

	r.POST("/api/estimations", authMiddleware(), createEstimation)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
