package models

// 產品主類別
type ProductCategory struct {
	ID           uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	CategoryCode string `json:"category_code" binding:"required"`
	Name         string `json:"name" binding:"required"`
}

// 產品形狀
type ProductShape struct {
	ID        uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	ShapeCode string `json:"shape_code" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

// 產品功能
type ProductFunction struct {
	ID           uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	FunctionCode string `json:"function_code" binding:"required"`
	Name         string `json:"name" binding:"required"`
}

// 產品規格
type ProductSpecification struct {
	ID       uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	SpecCode string `json:"spec_code" binding:"required"`
	Name     string `json:"name" binding:"required"`
	ParentID *uint  `json:"parent_id"` // null or reference another spec
}
