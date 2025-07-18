// fastener-api-main/models/company.go
package models

import "time"

// Company 定義了公司資料的結構
type Company struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" binding:"required"` // binding:"required" 確保前端傳來時不能是空值
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
