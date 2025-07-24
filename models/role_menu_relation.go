package models

// RoleMenuRelation 代表角色與菜單的關聯
type RoleMenuRelation struct {
    RoleID uint `gorm:"primaryKey"` // 角色 ID，作為複合主鍵的一部分
    MenuID uint `gorm:"primaryKey"` // 菜單 ID，作為複合主鍵的一部分
    // 您可以根據需要在此處添加其他字段，例如：
    // IsActive  bool `gorm:"default:true"` // 表示此關聯是否啟用
    // OrderNo   int  // 如果菜單在特定角色下有顯示順序
}
