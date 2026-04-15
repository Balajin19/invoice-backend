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

func GetProducts(c *gin.Context) {
	categoryID := c.Query("categoryId")

	var (
		products []models.Product
		err      error
	)

	if categoryID != "" {
		products, err = repository.GetProductsByCategoryID(categoryID)
	} else {
		products, err = repository.GetAllProducts()
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

func GetProductByID(c *gin.Context) {
	productID := c.Param("id")
	product, err := repository.GetProductByID(productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, product)
}

func CreateProduct(c *gin.Context) {
	userEmail, ok := userEmail(c)
	if !ok {
		return
	}

	var rawData json.RawMessage
	if err := c.ShouldBindJSON(&rawData); err != nil {
		logOperationError("create product bind payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if it's an array or single object
	if len(rawData) > 0 && rawData[0] == '[' {
		// Handle as array (bulk)
		handleBulkProducts(c, rawData, userEmail)
	} else {
		// Handle as single object
		var product models.Product
		if err := json.Unmarshal(rawData, &product); err != nil {
			logOperationError("create product unmarshal payload", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Names come from master tables via IDs; ignore payload names/text values.
		product.CategoryName = ""
		product.Unit = ""
		created, err := repository.CreateProduct(product, userEmail)
		if err != nil {
			if errors.Is(err, repository.ErrDuplicateProduct) {
				logOperationError("create product duplicate name="+product.ProductName+" categoryId="+product.CategoryId, err)
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			} else if errors.Is(err, repository.ErrUnitNotFound) {
				logOperationError("create product invalid unitId name="+product.ProductName+" categoryId="+product.CategoryId, err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			} else {
				logOperationError("create product name="+product.ProductName+" categoryId="+product.CategoryId+" email="+userEmail, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		logOperationSuccess("create product id=" + created.ProductId + " email=" + userEmail)
		c.JSON(http.StatusCreated, created)
	}
}

func handleBulkProducts(c *gin.Context, rawData json.RawMessage, userEmail string) {
	var products []models.Product
	if err := json.Unmarshal(rawData, &products); err != nil {
		logOperationError("bulk create products unmarshal payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(products) == 0 {
		logOperationError("bulk create products empty payload", errors.New("no products provided"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "no products provided"})
		return
	}

	created := make([]models.Product, 0)
	var failedProducts []gin.H

	for _, product := range products {
		// Names come from master tables via IDs; ignore payload names/text values.
		product.CategoryName = ""
		product.Unit = ""
		result, err := repository.CreateProduct(product, userEmail)
		if err != nil {
			logOperationError("bulk create product name="+product.ProductName+" categoryId="+product.CategoryId+" email="+userEmail, err)
			failedProducts = append(failedProducts, gin.H{
				"productName": product.ProductName,
				"categoryId":  product.CategoryId,
				"error":       err.Error(),
			})
		} else {
			logOperationSuccess("bulk create product id=" + result.ProductId + " email=" + userEmail)
			created = append(created, *result)
		}
	}

	response := gin.H{
		"created": created,
		"total":   len(products),
		"success": len(created),
		"failed":  len(failedProducts),
	}

	if len(failedProducts) > 0 {
		response["failures"] = failedProducts
		if len(created) == 0 {
			c.JSON(http.StatusBadRequest, response)
		} else {
			c.JSON(http.StatusMultiStatus, response)
		}
	} else {
		c.JSON(http.StatusCreated, response)
	}
}

func UpdateProduct(c *gin.Context) {
	productID := c.Param("id")
	userEmail, ok := userEmail(c)
	if !ok {
		return
	}

	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		logOperationError("update product bind payload id="+productID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Names come from master tables via IDs; ignore payload names/text values.
	product.CategoryName = ""
	product.Unit = ""
	updated, err := repository.UpdateProduct(productID, product, userEmail)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateProduct) {
			logOperationError("update product duplicate id="+productID, err)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if errors.Is(err, repository.ErrUnitNotFound) {
			logOperationError("update product invalid unitId id="+productID, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else if errors.Is(err, sql.ErrNoRows) {
			logOperationError("update product not found id="+productID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		} else {
			logOperationError("update product id="+productID+" email="+userEmail, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	logOperationSuccess("update product id=" + updated.ProductId + " email=" + userEmail)
	c.JSON(http.StatusOK, updated)
}

func DeleteProduct(c *gin.Context) {
	productID := c.Param("id")
	err := repository.DeleteProduct(productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
}
