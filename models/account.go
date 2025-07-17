// models/account.go
package models

type Account struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
	Role     string `json:"role" db:"role"`
	IsActive bool   `json:"is_active" db:"is_active"`
}
