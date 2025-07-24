package models

type Role struct {
    ID   uint   `json:"id" gorm:"primaryKey"`
    Name string `json:"name"`
    // 你有 permissions 欄位的話也可以加
    Permissions []string `json:"permissions" gorm:"type:jsonb"`
}
