package handler

import (
    "net/http"
    "fastener-api/db"
    "fastener-api/models"
    "github.com/gin-gonic/gin"
)

// 查詢所有角色
func GetRoles(c *gin.Context) {
    var roles []models.Role
    if err := db.DB.Find(&roles).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢角色失敗"})
        return
    }
    c.JSON(http.StatusOK, roles)
}
