package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "fastener-api/db"
    "fastener-api/models"
)

// 查詢角色擁有哪些 menu（回傳 menu id list）
func GetRoleMenus(c *gin.Context) {
    roleID := c.Query("role_id")
    var rels []models.RoleMenuRelation
    db.DB.Where("role_id = ?", roleID).Find(&rels)
    menuIDs := []uint{}
    for _, rel := range rels {
        menuIDs = append(menuIDs, rel.MenuID)
    }
    c.JSON(http.StatusOK, menuIDs)
}

// 批次更新角色 menu 權限
func UpdateRoleMenus(c *gin.Context) {
    var input struct {
        RoleID  uint   `json:"role_id"`
        MenuIDs []uint `json:"menu_ids"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "格式錯誤"})
        return
    }
    // 先清空此角色所有 menu 關聯
    db.DB.Where("role_id = ?", input.RoleID).Delete(&models.RoleMenuRelation{})
    // 再新增
    for _, mid := range input.MenuIDs {
        db.DB.Create(&models.RoleMenuRelation{RoleID: input.RoleID, MenuID: mid})
    }
    c.Status(http.StatusOK)
}

// 單一刪除（可選）
func DeleteRoleMenu(c *gin.Context) {
    var input struct {
        RoleID uint `json:"role_id"`
        MenuID uint `json:"menu_id"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "格式錯誤"})
        return
    }
    db.DB.Where("role_id = ? AND menu_id = ?", input.RoleID, input.MenuID).Delete(&models.RoleMenuRelation{})
    c.Status(http.StatusOK)
}
