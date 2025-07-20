// fastener-api-main/handler/companies.go
package handler

import (
	"fastener-api/db"
	"fastener-api/models"
	"net/http"
	"strconv"

	"github.comcom/gin-gonic/gin"
)

// --- Get All Companies (as a tree) ---
func GetCompanies(c *gin.Context) {
	rows, err := db.Conn.Query("SELECT id, name, parent_id, created_at, updated_at FROM companies ORDER BY name")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query companies: " + err.Error()})
		return
	}
	defer rows.Close()

	companyMap := make(map[int]*models.Company)
	var allCompanies []*models.Company

	for rows.Next() {
		var company models.Company
		if err := rows.Scan(&company.ID, &company.Name, &company.ParentID, &company.CreatedAt, &company.UpdatedAt); err == nil {
			node := company
			companyMap[node.ID] = &node
			allCompanies = append(allCompanies, &node)
		}
	}
	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading company data: " + err.Error()})
		return
	}

	var rootCompanies []*models.Company
	for _, company := range allCompanies {
		if company.ParentID.Valid {
			if parent, ok := companyMap[int(company.ParentID.Int64)]; ok {
				parent.Children = append(parent.Children, company)
			}
		} else {
			rootCompanies = append(rootCompanies, company)
		}
	}

	if rootCompanies == nil {
		rootCompanies = make([]*models.Company, 0)
	}
	c.JSON(http.StatusOK, rootCompanies)
}

// --- Create Company (supports parent_id) ---
func CreateCompany(c *gin.Context) {
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	sqlStatement := `INSERT INTO companies (name, parent_id) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	err := db.Conn.QueryRow(sqlStatement, company.Name, company.ParentID).Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, company)
}

// --- Get Company By ID ---
func GetCompanyByID(c *gin.Context) {
	// This function might need more complex logic if you need to fetch children too,
	// but for now, it returns the basic company data.
	id := c.Param("id")
	var company models.Company
	err := db.Conn.QueryRow(
		"SELECT id, name, parent_id, created_at, updated_at FROM companies WHERE id = $1", id,
	).Scan(&company.ID, &company.Name, &company.ParentID, &company.CreatedAt, &company.UpdatedAt)
	
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}
	c.JSON(http.StatusOK, company)
}

// --- Update Company ---
func UpdateCompany(c *gin.Context) {
	idStr := c.Param("id")
	var company models.Company
	if err := c.ShouldBindJSON(&company); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	sqlStatement := `UPDATE companies SET name = $1, parent_id = $2, updated_at = NOW() WHERE id = $3`
	res, err := db.Conn.Exec(sqlStatement, company.Name, company.ParentID, idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company: " + err.Error()})
		return
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company to update not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Company updated successfully"})
}

// --- Delete Company ---
func DeleteCompany(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	if id == 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete the root company (ID 1)"})
		return
	}

	sqlStatement := `DELETE FROM companies WHERE id = $1`
	res, err := db.Conn.Exec(sqlStatement, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete company: " + err.Error()})
		return
	}
	count, _ := res.RowsAffected()
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company to delete not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Company deleted successfully"})
}
