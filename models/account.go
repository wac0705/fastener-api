// fastener-api-main/models/account.go
package models

// UserAccount 用於 API 回傳給前端的使用者資訊
type UserAccount struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password,omitempty" db:"password"` // 在 JSON 輸出時忽略密碼
	Role     string `json:"role" db:"role"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

// CreateAccountRequest 用於接收前端建立帳號的請求
type CreateAccountRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// UpdateAccountRequest 用於接收前端更新帳號的請求
type UpdateAccountRequest struct {
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}
