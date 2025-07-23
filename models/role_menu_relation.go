// models/role_menu_relation.go
package models

type RoleMenuRelation struct {
    RoleID int `json:"role_id" gorm:"primaryKey"`
    MenuID int `json:"menu_id" gorm:"primaryKey"`
}
