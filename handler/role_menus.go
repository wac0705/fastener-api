// handler/role_menus.go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "fastener-api-main/db"
    "fastener-api-main/models"
)

func GetRoleMenus(c *gin.Context) {
    roleID := c.Query("role_id")
    var rels []models.RoleMenuRelation
    db.DB.Where("role_id = ?", roleID).Find(&rels)
    menuIDs := []int{}
    for _, rel := range rels {
        menuIDs = append(menuIDs, rel.MenuID)
    }
    c.JSON(http.StatusOK, menuIDs)
}

func UpdateRoleMenus(c *gin.Context) {
    var input struct {
        RoleID  int   `json:"role_id"`
        MenuIDs []int `json:"menu_ids"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "格式錯誤"})
        return
    }
    db.DB.Where("role_id = ?", input.RoleID).Delete(&models.RoleMenuRelation{})
    for _, mid := range input.MenuIDs {
        db.DB.Create(&models.RoleMenuRelation{RoleID: input.RoleID, MenuID: mid})
    }
    c.Status(http.StatusOK)
}

func DeleteRoleMenu(c *gin.Context) {
    var input struct {
        RoleID int `json:"role_id"`
        MenuID int `json:"menu_id"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "格式錯誤"})
        return
    }
    db.DB.Where("role_id = ? AND menu_id = ?", input.RoleID, input.MenuID).Delete(&models.RoleMenuRelation{})
    c.Status(http.StatusOK)
}
