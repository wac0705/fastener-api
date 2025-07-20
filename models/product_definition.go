// fastener-api-main/models/product_definition.go
package models

import "database/sql"

// ProductCategory 定義了產品主類別的結構
type ProductCategory struct {
	ID           int    `json:"id"`
	CategoryCode string `json:"category_code" binding:"required"`
	Name         string `json:"name" binding:"required"`
}

// ProductShape 定義了產品形狀的結構
type ProductShape struct {
	ID        int    `json:"id"`
	ShapeCode string `json:"shape_code" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// ProductFunction 定義了產品功能的結構
type ProductFunction struct {
	ID           int    `json:"id"`
	FunctionCode string `json:"function_code" binding:"required"`
	Name         string `json:"name" binding:"required"`
}

// ProductSpecification 定義了產品規格的結構
type ProductSpecification struct {
	ID       int            `json:"id"`
	SpecCode string         `json:"spec_code" binding:"required"`
	Name     string         `json:"name" binding:"required"`
	ParentID sql.NullInt64  `json:"parent_id"` // 使用 sql.NullInt64 來處理可能為 NULL 的 parent_id
}
