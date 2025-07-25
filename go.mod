module github.com/wac0705/fastener-api

go 1.22.5

require (
	github.com/gofiber/fiber/v2 v2.52.5
	github.com/gofiber/jwt/v3 v3.3.10
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/joho/godotenv v1.5.1
	github.com/jackc/pgx/v5 v5.5.5 // ✅ 修正版本
	golang.org/x/crypto v0.24.0
	gorm.io/driver/postgres v1.5.9
	gorm.io/gorm v1.25.10

	// indirect（由其他套件引入）
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/gofiber/utils/v2 v2.0.0-beta.4 // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)
