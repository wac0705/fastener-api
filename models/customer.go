// fastener-api-main/models/customer.go
package models

import "time"

// CustomerTransactionTerm 定義了客戶交易條件的結構
type CustomerTransactionTerm struct {
	ID                 int     `json:"id"`
	CustomerID         int     `json:"customer_id"`
	CompanyID          int     `json:"company_id" binding:"required"`
	Incoterm           string  `json:"incoterm"`
	CurrencyCode       string  `json:"currency_code"`
	CommissionRate     float64 `json:"commission_rate"`
	ExportPort         string  `json:"export_port"`
	DestinationCountry string  `json:"destination_country"`
	IsPrimary          bool    `json:"is_primary"`
	Remarks            string  `json:"remarks"`
}

// Customer 定義了客戶主檔及其所有交易條件的完整結構
type Customer struct {
	ID                int                       `json:"id"`
	GroupCustomerCode string                    `json:"group_customer_code" binding:"required"`
	GroupCustomerName string                    `json:"group_customer_name" binding:"required"`
	Remarks           string                    `json:"remarks"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
	TransactionTerms  []CustomerTransactionTerm `json:"transaction_terms,omitempty"` // omitempty 讓這個欄位在列表查詢時可以被忽略
}
