package models

type Account struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"` // 密碼雜湊值
	Role     string `json:"role" db:"role"`         // 角色名稱
	IsActive bool   `json:"is_active" db:"is_active"`
}
