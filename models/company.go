package models

import (
    "time"
)

type Company struct {
    ID        int64      `json:"id"`
    Name      string     `json:"name"`
    ParentID  *int64     `json:"parent_id"` // pointer 型別，支援 null/數字
    Currency  string     `json:"currency"`
    Language  string     `json:"language"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    Children  []*Company `json:"children,omitempty"` // 你有用階層結構時可保留
}
