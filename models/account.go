package models

// UserAccount 用於 API 回傳給前端的使用者資訊
type UserAccount struct {
	ID          int    `json:"id" db:"id"`
	Username    string `json:"username" db:"username"`
	Role        string `json:"role" db:"role"`
	IsActive    bool   `json:"is_active" db:"is_active"`
	CompanyID   int    `json:"company_id" db:"company_id"`         // 新增
	CompanyName string `json:"company_name" db:"company_name"`     // 新增（查詢時 join company.name 用）
}

// CreateAccountRequest 用於接收前端建立帳號的請求
type CreateAccountRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Role      string `json:"role"`
	CompanyID int    `json:"company_id"`   // 新增
}

// UpdateAccountRequest 用於接收前端更新帳號的請求
type UpdateAccountRequest struct {
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CompanyID int    `json:"company_id"`   // 新增
}
