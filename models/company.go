// fastener-api-main/models/company.go
package models

import (
	"database/sql"
	"time"
)

// Company 定義了公司資料的結構 (支援階層)
type Company struct {
	ID        int             `json:"id"`
	Name      string          `json:"name" binding:"required"`
	ParentID  sql.NullInt64   `json:"parent_id"` // 使用 sql.NullInt64 來處理可能為 NULL 的 parent_id
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Children  []*Company      `json:"children,omitempty"` // 用於構建樹狀結構，omitempty 表示若無子公司則不顯示此欄位
}
