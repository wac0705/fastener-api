package models

import "time"

// 客戶主檔
type Customer struct {
	ID                uint                      `json:"id" gorm:"primaryKey;autoIncrement"`
	GroupCustomerCode string                    `json:"group_customer_code" binding:"required"`
	GroupCustomerName string                    `json:"group_customer_name" binding:"required"`
	Remarks           string                    `json:"remarks"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
	TransactionTerms  []CustomerTransactionTerm `json:"transaction_terms,omitempty" gorm:"-"`
}

// 客戶交易條件
type CustomerTransactionTerm struct {
	ID                 uint    `json:"id" gorm:"primaryKey;autoIncrement"`
	CustomerID         uint    `json:"customer_id"`
	CompanyID          uint    `json:"company_id" binding:"required"`
	Incoterm           string  `json:"incoterm"`
	CurrencyCode       string  `json:"currency_code"`
	CommissionRate     float64 `json:"commission_rate"`
	ExportPort         string  `json:"export_port"`
	DestinationCountry string  `json:"destination_country"`
	IsPrimary          bool    `json:"is_primary"`
	Remarks            string  `json:"remarks"`
}
