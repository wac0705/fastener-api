package models

import (
    "time"
)

type Company struct {
    ID        int64      `json:"id"`
    Name      string     `json:"name"`
    ParentID  *int64     `json:"parent_id"` // 用 pointer 支援 null/數字
    Currency  string     `json:"currency"`
    Language  string     `json:"language"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    Children  []*Company `json:"children,omitempty"` // 支援樹狀回傳
}
