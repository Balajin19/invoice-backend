package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"invoice-generator-backend/internal/models"
	"invoice-generator-backend/repository"

	"github.com/gin-gonic/gin"
)

func GetInvoiceSettings(c *gin.Context) {
	companyID := c.Query("companyId")

	var (
		settings []models.InvoiceSettings
		err      error
	)

	if companyID != "" {
		settings, err = repository.GetInvoiceSettingsByCompanyID(companyID)
	} else {
		settings, err = repository.GetInvoiceSettings()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func GetInvoiceSettingsByID(c *gin.Context) {
	s, err := repository.GetInvoiceSettingsByID(c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s, err = repository.GetInvoiceSettingsByCompanyIDSingle(c.Param("id"))
			if err == nil {
				c.JSON(http.StatusOK, s)
				return
			}
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "invoice settings not found"})
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func CreateInvoiceSettings(c *gin.Context) {
	email, ok := userEmail(c)
	if !ok {
		return
	}
	var payload models.InvoiceSettings
	if err := c.ShouldBindJSON(&payload); err != nil {
		logOperationError("create invoice settings bind payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s, err := repository.CreateInvoiceSettings(email, payload)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateInvoiceSettingsFY) {
			logOperationError("create invoice settings duplicate FY email="+email, err)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		logOperationError("create invoice settings email="+email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logOperationSuccess("create invoice settings id=" + s.ID + " email=" + email)
	c.JSON(http.StatusCreated, s)
}

func UpdateInvoiceSettings(c *gin.Context) {
	email, ok := userEmail(c)
	if !ok {
		return
	}
	settingID := c.Param("id")
	var payload models.InvoiceSettings
	if err := c.ShouldBindJSON(&payload); err != nil {
		logOperationError("update invoice settings bind payload id="+settingID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s, err := repository.UpdateInvoiceSettings(settingID, email, payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s, err = repository.UpdateInvoiceSettingsByCompanyID(settingID, email, payload)
			if err == nil {
				logOperationSuccess("update invoice settings by company id=" + settingID + " email=" + email)
				c.JSON(http.StatusOK, s)
				return
			}
			if errors.Is(err, sql.ErrNoRows) {
				logOperationError("update invoice settings not found id="+settingID, err)
				c.JSON(http.StatusNotFound, gin.H{"error": "invoice settings not found"})
				return
			}
		}
		logOperationError("update invoice settings id="+settingID+" email="+email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logOperationSuccess("update invoice settings id=" + s.ID + " email=" + email)
	c.JSON(http.StatusOK, s)
}

func DeleteInvoiceSettings(c *gin.Context) {
	err := repository.DeleteInvoiceSettingsByID(c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice settings not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "invoice settings deleted"})
}
