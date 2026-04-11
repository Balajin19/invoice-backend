package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"invoice-generator-backend/internal/models"
	"invoice-generator-backend/repository"

	"github.com/gin-gonic/gin"
)


func GetCompanyDetails(c *gin.Context) {
	if _, ok := userEmail(c); !ok {
		return
	}
	companies, err := repository.GetCompanySettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"exists": len(companies) > 0, "companies": companies})
}

func GetCompanyDetailsByID(c *gin.Context) {
	company, err := repository.GetCompanySettingsByID(c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, company)
}

func CreateCompany(c *gin.Context) {
	email, ok := userEmail(c)
	if !ok {
		return
	}
	var payload models.Companies
	if err := c.ShouldBindJSON(&payload); err != nil {
		logOperationError("create company bind payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	company, err := repository.CreateCompanySettings(email, payload)
	if err != nil {
		logOperationError("create company email="+email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logOperationSuccess("create company id=" + company.CompanyID + " email=" + email)
	c.JSON(http.StatusCreated, company)
}

func UpdateCompany(c *gin.Context) {
	email, ok := userEmail(c)
	if !ok {
		return
	}
	companyID := c.Param("id")
	var payload models.Companies
	if err := c.ShouldBindJSON(&payload); err != nil {
		logOperationError("update company bind payload id="+companyID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	company, err := repository.UpdateCompanySettingsByID(companyID, email, payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logOperationError("update company not found id="+companyID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
			return
		}
		logOperationError("update company id="+companyID+" email="+email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logOperationSuccess("update company id=" + company.CompanyID + " email=" + email)
	c.JSON(http.StatusOK, company)
}

func DeleteCompany(c *gin.Context) {
	if err := repository.DeleteCompanySettingsByID(c.Param("id")); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "company not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "company deleted"})
}
