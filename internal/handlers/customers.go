package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"invoice-generator-backend/internal/models"
	"invoice-generator-backend/repository"

	"github.com/gin-gonic/gin"
)

func GetCustomers(c *gin.Context) {
	customers, err := repository.GetAllCustomers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, customers)
}

func GetCustomerByID(c *gin.Context) {
	customerId := c.Param("id")
	customer, err := repository.GetCustomerByID(customerId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, customer)
}

func CreateCustomer(c *gin.Context) {
	email, ok := userEmail(c)
	if !ok {
		return
	}

	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		logOperationError("create customer bind payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := repository.CreateCustomer(customer, email)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateCustomer) {
			logOperationError("create customer duplicate customerName="+customer.CustomerName, err)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			logOperationError("create customer customerName="+customer.CustomerName, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	logOperationSuccess("create customer id=" + created.CustomerId)
	c.JSON(http.StatusCreated, created)
}

func UpdateCustomer(c *gin.Context) {
	email, ok := userEmail(c)
	if !ok {
		return
	}

	customerId := c.Param("id")
	var customer models.Customer
	if err := c.ShouldBindJSON(&customer); err != nil {
		logOperationError("update customer bind payload id="+customerId, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := repository.UpdateCustomer(customerId, customer, email)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateCustomer) {
			logOperationError("update customer duplicate id="+customerId, err)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if errors.Is(err, sql.ErrNoRows) {
			logOperationError("update customer not found id="+customerId, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		} else {
			logOperationError("update customer id="+customerId, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	logOperationSuccess("update customer id=" + updated.CustomerId)
	c.JSON(http.StatusOK, updated)
}

func DeleteCustomer(c *gin.Context) {
	customerId := c.Param("id")
	err := repository.DeleteCustomer(customerId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "customer deleted successfully"})
}