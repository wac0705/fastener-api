# 使用官方的 Go 映像檔作為基礎
FROM golang:1.22-alpine

# 設定工作目錄
WORKDIR /app

# 在 Alpine 系統中安裝 git，因為 go mod 可能會需要它
RUN apk add --no-cache git

# 僅複製 go.mod 和 go.sum 檔案
# 這樣只有在依賴變更時，才會重新下載
COPY go.mod go.sum ./

# 下載所有依賴
RUN go mod download

# 複製所有專案原始碼
COPY . .

# 編譯 Go 應用程式
RUN go build -o main .

# 設定容器啟動時執行的命令
CMD ["./main"]

# 向 Docker 宣告容器在執行時監聽的埠號 (僅為文件作用)
EXPOSE 3001
