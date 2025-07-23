package models

import (
    "time"
)

type Company struct {
    ID        uint       `json:"id" gorm:"primaryKey;autoIncrement"`
    Name      string     `json:"name"`
    ParentID  *uint      `json:"parent_id"`            // 用 uint 指標支援 null
    Currency  string     `json:"currency"`
    Language  string     `json:"language"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    Children  []*Company `json:"children,omitempty" gorm:"-"`
}
