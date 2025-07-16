package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"        // ✅ 加上這行
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"  // PostgreSQL驅動
)

type Estimation struct {
	InquiryID     int             `json:"inquiry_id"`
	Materials     json.RawMessage `json:"materials"` // e.g., [{"code":"M8碳鋼","cost":0.5}]
	Processes     json.RawMessage `json:"processes"`
	Logistics     json.RawMessage `json:"logistics"`
	TotalCost     float64         `json:"total_cost"`
	AISuggestions float64         `json:"ai_suggestions"` // AI預測調整
}

var db *sql.DB

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

func main() {
	initDB()
	r := gin.Default()
	
	r.Use(cors.Default()) // ✅ 允許所有來源跨域，測試或前端呼叫用
	
	r.POST("/api/estimations", createEstimation)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
