// fastener-api-main/handler/customers.go
package handler

import (
	"database/sql"
	"fastener-api/db"
	"fastener-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// --- 建立新客戶 ---
func CreateCustomer(c *gin.Context) {
    var customer models.Customer
    if err := c.ShouldBindJSON(&customer); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式: " + err.Error()})
        return
    }

    sqlStatement := `
		INSERT INTO customers (group_customer_code, group_customer_name, remarks) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at, updated_at
	`
    err := db.Conn.QueryRow(sqlStatement, customer.GroupCustomerCode, customer.GroupCustomerName, customer.Remarks).Scan(&customer.ID, &customer.CreatedAt, &customer.UpdatedAt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "建立客戶失敗: " + err.Error()})
        return
    }

    c.JSON(http.StatusCreated, customer)
}

// --- 查詢所有客戶 (簡化列表) ---
func GetCustomers(c *gin.Context) {
	rows, err := db.Conn.Query(`
		SELECT id, group_customer_code, group_customer_name, remarks, created_at, updated_at 
		FROM customers 
		ORDER BY group_customer_code
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢客戶資料失敗: " + err.Error()})
		return
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		// 使用 sql.NullString 來處理可能為 NULL 的 remarks 欄位
		var remarks sql.NullString 
		if err := rows.Scan(&customer.ID, &customer.GroupCustomerCode, &customer.GroupCustomerName, &remarks, &customer.CreatedAt, &customer.UpdatedAt); err == nil {
			customer.Remarks = remarks.String
			customers = append(customers, customer)
		}
	}
	c.JSON(http.StatusOK, customers)
}

// --- 查詢單一客戶 (包含所有交易條件) ---
func GetCustomerByID(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer
	var remarks sql.NullString

	// 1. 查詢客戶主檔
	err := db.Conn.QueryRow(
		`SELECT id, group_customer_code, group_customer_name, remarks, created_at, updated_at FROM customers WHERE id = $1`,
		id,
	).Scan(&customer.ID, &customer.GroupCustomerCode, &customer.GroupCustomerName, &remarks, &customer.CreatedAt, &customer.UpdatedAt)
	
    if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "找不到指定的客戶"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢客戶主檔失敗: " + err.Error()})
		return
	}
	customer.Remarks = remarks.String

	// 2. 查詢該客戶的所有交易條件
	rows, err := db.Conn.Query(`
		SELECT id, customer_id, company_id, incoterm, currency_code 
		FROM customer_transaction_terms 
		WHERE customer_id = $1
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢交易條件失敗: " + err.Error()})
		return
	}
	defer rows.Close()

	var terms []models.CustomerTransactionTerm
	for rows.Next() {
		var term models.CustomerTransactionTerm
		// 範例：掃描部分欄位，您可以根據需求擴充
		if err := rows.Scan(&term.ID, &term.CustomerID, &term.CompanyID, &term.Incoterm, &term.CurrencyCode); err == nil {
			terms = append(terms, term)
		}
	}
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

    sqlStatement := `
        UPDATE customers 
        SET group_customer_code = $1, group_customer_name = $2, remarks = $3, updated_at = NOW() 
        WHERE id = $4 
        RETURNING id, created_at, updated_at
    `
    err := db.Conn.QueryRow(sqlStatement, customer.GroupCustomerCode, customer.GroupCustomerName, customer.Remarks, id).Scan(&customer.ID, &customer.CreatedAt, &customer.UpdatedAt)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "更新客戶失敗: " + err.Error()})
        return
    }

    c.JSON(http.StatusOK, customer)
}

// --- 刪除客戶 ---
func DeleteCustomer(c *gin.Context) {
    id := c.Param("id")
    
    // ON DELETE CASCADE 會自動刪除關聯的交易條件
    sqlStatement := `DELETE FROM customers WHERE id = $1`
    res, err := db.Conn.Exec(sqlStatement, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "刪除客戶失敗: " + err.Error()})
        return
    }

    count, _ := res.RowsAffected()
    if count == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "找不到要刪除的客戶"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "客戶刪除成功"})
}
