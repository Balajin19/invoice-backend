package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"invoice-generator-backend/internal/models"
	"invoice-generator-backend/repository"

	"github.com/gin-gonic/gin"
)

func GetUnits(c *gin.Context) {
	units, err := repository.GetAllUnits()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, units)
}

func GetUnitByID(c *gin.Context) {
	unitID := c.Param("id")
	unit, err := repository.GetUnitByID(unitID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, unit)
}

func CreateUnit(c *gin.Context) {
	email, ok := userEmail(c)
	if !ok {
		return
	}

	var rawData json.RawMessage
	if err := c.ShouldBindJSON(&rawData); err != nil {
		logOperationError("create unit bind payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(rawData) > 0 && rawData[0] == '[' {
		handleBulkUnits(c, rawData, email)
		return
	}

	var unit models.Unit
	if err := json.Unmarshal(rawData, &unit); err != nil {
		logOperationError("create unit unmarshal payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := repository.CreateUnit(unit, email)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateUnit) {
			logOperationError("create unit duplicate name="+unit.UnitName, err)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			logOperationError("create unit name="+unit.UnitName+" email="+email, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	logOperationSuccess("create unit id=" + created.UnitID + " email=" + email)
	c.JSON(http.StatusCreated, created)
}

func handleBulkUnits(c *gin.Context, rawData json.RawMessage, email string) {
	var units []models.Unit
	if err := json.Unmarshal(rawData, &units); err != nil {
		logOperationError("bulk create units unmarshal payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(units) == 0 {
		logOperationError("bulk create units empty payload", errors.New("no units provided"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "no units provided"})
		return
	}

	created := make([]models.Unit, 0)
	failed := make([]gin.H, 0)

	for _, unit := range units {
		result, err := repository.CreateUnit(unit, email)
		if err != nil {
			logOperationError("bulk create unit name="+unit.UnitName+" email="+email, err)
			failed = append(failed, gin.H{
				"unitName": unit.UnitName,
				"error":    err.Error(),
			})
			continue
		}

		logOperationSuccess("bulk create unit id=" + result.UnitID + " email=" + email)
		created = append(created, *result)
	}

	response := gin.H{
		"created": created,
		"total":   len(units),
		"success": len(created),
		"failed":  len(failed),
	}

	if len(failed) > 0 {
		response["failures"] = failed
		c.JSON(http.StatusMultiStatus, response)
		return
	}

	c.JSON(http.StatusCreated, response)
}

func UpdateUnit(c *gin.Context) {
	unitID := c.Param("id")
	email, ok := userEmail(c)
	if !ok {
		return
	}

	var unit models.Unit
	if err := c.ShouldBindJSON(&unit); err != nil {
		logOperationError("update unit bind payload id="+unitID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := repository.UpdateUnit(unitID, unit, email)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateUnit) {
			logOperationError("update unit duplicate id="+unitID, err)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if errors.Is(err, sql.ErrNoRows) {
			logOperationError("update unit not found id="+unitID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
		} else {
			logOperationError("update unit id="+unitID+" email="+email, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	logOperationSuccess("update unit id=" + updated.UnitID + " email=" + email)
	c.JSON(http.StatusOK, updated)
}

func DeleteUnit(c *gin.Context) {
	unitID := c.Param("id")
	err := repository.DeleteUnit(unitID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "unit not found"})
		} else if errors.Is(err, repository.ErrUnitInUse) {
			c.JSON(http.StatusConflict, gin.H{"error": "unit is in use by products"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "unit deleted successfully"})
}
