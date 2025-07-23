package models

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
