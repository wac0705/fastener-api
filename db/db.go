package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	// 從環境變數取得連線字串
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("❌ 缺少 DATABASE_URL 環境變數")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ GORM 連線失敗:", err)
	}

	log.Println("✅ 成功連接 PostgreSQL (GORM)")
}
