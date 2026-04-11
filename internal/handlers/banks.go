package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"invoice-generator-backend/internal/models"
	"invoice-generator-backend/repository"

	"github.com/gin-gonic/gin"
)

func GetBankDetails(c *gin.Context) {
	if _, ok := userEmail(c); !ok {
		return
	}
	banks, err := repository.GetBankSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"exists": len(banks) > 0, "banks": banks})
}

func GetBankDetailsByID(c *gin.Context) {
	bank, err := repository.GetBankSettingsByID(c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "bank not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bank)
}

func CreateBank(c *gin.Context) {
	email, ok := userEmail(c)
	if !ok {
		return
	}
	var payload models.Banks
	if err := c.ShouldBindJSON(&payload); err != nil {
		logOperationError("create bank bind payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bank, err := repository.CreateBankSettings(email, payload)
	if err != nil {
		logOperationError("create bank email="+email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logOperationSuccess("create bank id=" + bank.ID + " email=" + email)
	c.JSON(http.StatusCreated, bank)
}

func UpdateBank(c *gin.Context) {
	email, ok := userEmail(c)
	if !ok {
		return
	}
	bankID := c.Param("id")
	var payload models.Banks
	if err := c.ShouldBindJSON(&payload); err != nil {
		logOperationError("update bank bind payload id="+bankID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bank, err := repository.UpdateBankSettingsByID(bankID, email, payload)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logOperationError("update bank not found id="+bankID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "bank not found"})
			return
		}
		logOperationError("update bank id="+bankID+" email="+email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logOperationSuccess("update bank id=" + bank.ID + " email=" + email)
	c.JSON(http.StatusOK, bank)
}

func DeleteBank(c *gin.Context) {
	if err := repository.DeleteBankSettingsByID(c.Param("id")); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "bank not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "bank deleted"})
}
