package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func userEmail(c *gin.Context) (string, bool) {
	v, exists := c.Get("userEmail")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user context missing"})
		return "", false
	}
	email, ok := v.(string)
	if !ok || strings.TrimSpace(email) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user context"})
		return "", false
	}
	return email, true
}

func logOperationError(operation string, err error) {
	if err == nil {
		return
	}
	log.Println("DB ERROR:", operation, "-", err)
}

func logOperationSuccess(operation string) {
	log.Println("DB SUCCESS:", operation)
}
