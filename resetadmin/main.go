// resetadmin/main.go
package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq" // a postgres driver
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 1. 直接在此處設定您的資料庫連線 URL
	// 警告：直接將連線字串寫在程式碼中會有安全風險。
	// 這個工具僅為一次性使用，請勿將此檔案提交到公開的 Git 儲存庫。
	connStr := "postgres://root:NzW7lrhfHqt2aoe1MA64085jmpn93GZX@tpe0.clusters.zeabur.com:32072/postgres?sslmode=disable"
	fmt.Println("已讀取資料庫連線位址。")


	// 2. 連接到 PostgreSQL 資料庫
	fmt.Println("正在連接到資料庫...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("無法開啟資料庫連線: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("無法連接到資料庫: %v", err)
	}
	fmt.Println("✅ 資料庫連接成功！")

	// 3. 獲取要重設密碼的使用者 ID 和新密碼
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Print("👉 請輸入要重設密碼的使用者 ID (預設為 1，代表 admin): ")
	idInput, _ := reader.ReadString('\n')
	idInput = strings.TrimSpace(idInput)
	if idInput == "" {
		idInput = "1"
	}

	fmt.Printf("👉 請為 ID 為 %s 的使用者輸入新密碼: ", idInput)
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if password == "" {
		log.Fatal("錯誤：密碼不能為空。")
	}

	// 4. 將新密碼加密
	fmt.Println("正在將新密碼加密...")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("密碼加密失敗: %v", err)
	}
	fmt.Println("✅ 密碼加密完成！")

	// 5. 更新資料庫中的密碼
	fmt.Printf("正在更新使用者 ID %s 的密碼...\n", idInput)
	sqlStatement := `UPDATE users SET password_hash = $1 WHERE id = $2;`
	res, err := db.Exec(sqlStatement, string(hashedPassword), idInput)
	if err != nil {
		log.Fatalf("更新資料庫失敗: %v", err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("無法獲取影響的行數: %v", err)
	}

	if count == 0 {
		log.Fatalf("⚠️ 更新失敗，找不到 ID 為 %s 的使用者。", idInput)
	}

	fmt.Printf("🎉 成功！使用者 ID %s 的密碼已重設。\n現在您可以使用新密碼登入了。\n", idInput)
}
