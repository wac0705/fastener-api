// models/menu.go
package models

type Menu struct {
    ID       int     `json:"id" gorm:"primaryKey;autoIncrement"`
    Name     string  `json:"name"`
    Path     string  `json:"path"`
    Icon     string  `json:"icon"`
    ParentID *int    `json:"parent_id"`
    OrderNo  int     `json:"order_no"`
    IsActive bool    `json:"is_active"`
}
