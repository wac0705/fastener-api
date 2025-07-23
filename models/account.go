package models

import "time"

// GORM ORM 用的 User struct，對應 users 資料表
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`      // 密碼雜湊
	RoleID    uint      `json:"role_id"`
	CompanyID uint      `json:"company_id"`    // tenant_id
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 用於 API 回傳給前端的帳號資訊
type UserAccount struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	IsActive    bool   `json:"is_active"`
	CompanyID   uint   `json:"company_id"`
	CompanyName string `json:"company_name"`
}

// 前端建立帳號請求
type CreateAccountRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	CompanyID uint   `json:"company_id"`
}

// 前端更新帳號請求
type UpdateAccountRequest struct {
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CompanyID uint   `json:"company_id"`
}
