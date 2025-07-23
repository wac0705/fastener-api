package handler

import (
	"fastener-api/db"
	"fastener-api/models"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
)

// --- 建立新客戶 ---
func CreateCustomer(c *gin.Context) {
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式: " + err.Error()})
		return
	}
	if err := db.DB.Create(&customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "建立客戶失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, customer)
}

// --- 查詢所有客戶 (簡化列表) ---
func GetCustomers(c *gin.Context) {
	var customers []models.Customer
	if err := db.DB.Order("group_customer_code").Find(&customers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢客戶資料失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, customers)
}

// --- 查詢單一客戶 (包含所有交易條件) ---
func GetCustomerByID(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer
	if err := db.DB.First(&customer, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到指定的客戶"})
		return
	}
	// 查詢所有該客戶的交易條件
	var terms []models.CustomerTransactionTerm
	db.DB.Where("customer_id = ?", customer.ID).Find(&terms)
	customer.TransactionTerms = terms
	c.JSON(http.StatusOK, customer)
}

// --- 更新客戶主檔 ---
func UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式: " + err.Error()})
		return
	}
	customer.ID, _ = strconv.ParseUint(id, 10, 64)
	if err := db.DB.Model(&models.Customer{}).Where("id = ?", id).Updates(customer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新客戶失敗: " + err.Error()})
		return
	}
	// 查回更新後結果
	db.DB.First(&customer, id)
	c.JSON(http.StatusOK, customer)
}

// --- 刪除客戶 ---
func DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Delete(&models.Customer{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除客戶失敗: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "客戶刪除成功"})
}

// --- 依 group_customer_code 查詢客戶 (支援 /code/:code) ---
func GetCustomerByCode(c *gin.Context) {
	code := c.Param("code")
	var customer models.Customer
	if err := db.DB.Where("group_customer_code = ?", code).First(&customer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到指定的客戶"})
		return
	}
	// 查詢所有該客戶的交易條件
	var terms []models.CustomerTransactionTerm
	db.DB.Where("customer_id = ?", customer.ID).Find(&terms)
	customer.TransactionTerms = terms
	c.JSON(http.StatusOK, customer)
}
