package models

type RoleMenuRelation struct {
    RoleID uint `json:"role_id" gorm:"primaryKey"`
    MenuID uint `json:"menu_id" gorm:"primaryKey"`
}
