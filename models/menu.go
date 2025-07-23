package models

type Menu struct {
    ID       uint   `json:"id" gorm:"primaryKey;autoIncrement"`
    Name     string `json:"name"`
    Path     string `json:"path"`
    Icon     string `json:"icon"`
    ParentID *uint  `json:"parent_id"`  // 支援 null
    OrderNo  int    `json:"order_no"`
    IsActive bool   `json:"is_active"`
}
