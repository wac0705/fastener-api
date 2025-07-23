package handler

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "fastener-api/db"
    "fastener-api/models"
)

// 取得所有 menu
func GetMenus(c *gin.Context) {
    var menus []models.Menu
    if err := db.DB.Order("order_no asc").Find(&menus).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
        return
    }
    c.JSON(http.StatusOK, menus)
}

// 新增 menu
func CreateMenu(c *gin.Context) {
    var menu models.Menu
    if err := c.ShouldBindJSON(&menu); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "格式錯誤"})
        return
    }
    if err := db.DB.Create(&menu).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "新增失敗"})
        return
    }
    c.JSON(http.StatusOK, menu)
}

// 更新 menu
func UpdateMenu(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    var menu models.Menu
    if err := db.DB.First(&menu, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "找不到 menu"})
        return
    }
    var input models.Menu
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "格式錯誤"})
        return
    }
    db.DB.Model(&menu).Updates(input)
    c.JSON(http.StatusOK, menu)
}

// 刪除 menu
func DeleteMenu(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    db.DB.Delete(&models.Menu{}, id)
    c.Status(http.StatusNoContent)
}
