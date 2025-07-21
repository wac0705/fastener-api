package models

import (
    "time"
)

type Company struct {
    ID        int64      `json:"id"`
    Name      string     `json:"name"`
    ParentID  *int64     `json:"parent_id"` // null or 數字都支援
    Currency  string     `json:"currency"`
    Language  string     `json:"language"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    Children  []*Company `json:"children,omitempty"` // 保持原樹狀架構
}
