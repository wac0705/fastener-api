// db/db.go
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
	Conn, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("DB連線失敗:", err)
	}

	err = Conn.Ping()
	if err != nil {
		log.Fatal("無法 Ping DB:", err)
	}

	log.Println("✅ 成功連線到 PostgreSQL")
}
