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

func GetCategories(c *gin.Context) {
	categories, err := repository.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func GetCategoryByID(c *gin.Context) {
	categoryID := c.Param("id")
	category, err := repository.GetCategoryByID(categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, category)
}

func CreateCategory(c *gin.Context) {
	var rawData json.RawMessage
	if err := c.ShouldBindJSON(&rawData); err != nil {
		logOperationError("create category bind payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if it's an array or single object
	if len(rawData) > 0 && rawData[0] == '[' {
		// Handle as array (bulk)
		handleBulkCategories(c, rawData)
	} else {
		// Handle as single object
		var category models.Category
		if err := json.Unmarshal(rawData, &category); err != nil {
			logOperationError("create category unmarshal payload", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		created, err := repository.CreateCategory(category)
		if err != nil {
			if errors.Is(err, repository.ErrDuplicateCategory) {
				logOperationError("create category duplicate name="+category.CategoryName, err)
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else {
				logOperationError("create category name="+category.CategoryName, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		logOperationSuccess("create category id=" + created.CategoryId)
		c.JSON(http.StatusCreated, created)
	}
}

func handleBulkCategories(c *gin.Context, rawData json.RawMessage) {
	var categories []models.Category
	if err := json.Unmarshal(rawData, &categories); err != nil {
		logOperationError("bulk create categories unmarshal payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(categories) == 0 {
		logOperationError("bulk create categories empty payload", errors.New("no categories provided"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "no categories provided"})
		return
	}

	created := make([]models.Category, 0)
	var failedCategories []gin.H

	for _, category := range categories {
		result, err := repository.CreateCategory(category)
		if err != nil {
			logOperationError("bulk create category name="+category.CategoryName, err)
			failedCategories = append(failedCategories, gin.H{
				"categoryName": category.CategoryName,
				"error":        err.Error(),
			})
		} else {
			logOperationSuccess("bulk create category id=" + result.CategoryId)
			created = append(created, *result)
		}
	}

	response := gin.H{
		"created": created,
		"total":   len(categories),
		"success": len(created),
		"failed":  len(failedCategories),
	}

	if len(failedCategories) > 0 {
		response["failures"] = failedCategories
		c.JSON(http.StatusMultiStatus, response)
	} else {
		c.JSON(http.StatusCreated, response)
	}
}

func UpdateCategory(c *gin.Context) {
	categoryID := c.Param("id")
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		logOperationError("update category bind payload id="+categoryID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := repository.UpdateCategory(categoryID, category)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateCategory) {
			logOperationError("update category duplicate id="+categoryID, err)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if errors.Is(err, sql.ErrNoRows) {
			logOperationError("update category not found id="+categoryID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		} else {
			logOperationError("update category id="+categoryID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	logOperationSuccess("update category id=" + updated.CategoryId)
	c.JSON(http.StatusOK, updated)
}

func DeleteCategory(c *gin.Context) {
	categoryID := c.Param("id")
	err := repository.DeleteCategory(categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "category deleted successfully"})
}
