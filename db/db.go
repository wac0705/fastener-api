package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var Conn *sql.DB

func Init() {
	var err error

	// 從環境變數讀取 DATABASE_URL
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("❌ 缺少 DATABASE_URL 環境變數")
	}

	Conn, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ 資料庫開啟失敗:", err)
	}

	err = Conn.Ping()
	if err != nil {
		log.Fatal("❌ 資料庫無法 Ping:", err)
	}

	log.Println("✅ 成功連接 PostgreSQL")
}
