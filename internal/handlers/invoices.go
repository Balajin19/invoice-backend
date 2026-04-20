package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"invoice-generator-backend/internal/models"
	"invoice-generator-backend/repository"

	"github.com/gin-gonic/gin"
)

type invoicePayload struct {
	InvoiceNumber string                `json:"invoiceNumber"`
	InvoiceDate   string                `json:"invoiceDate"`
	PONumber      string                `json:"po"`
	PONumberAlt   string                `json:"poNumber"`
	PODate        string                `json:"poDate"`
	InvoiceId     string                `json:"invoiceId"`
	CompanyId     string                `json:"companyId"`
	CustomerId    string                `json:"customerId"`
	CustomerName  string                `json:"customerName"`
	Address       string                `json:"customerAddress"`
	GSTIN         string                `json:"gstIn"`
	PaymentTerms  string                `json:"paymentTerms"`
	Amount        float64               `json:"subTotal"`
	OverallDiscount float64             `json:"overallDiscount"`
	CGST          float64               `json:"cgst"`
	SGST          float64               `json:"sgst"`
	IGST          float64               `json:"igst"`
	RoundedOff    float64               `json:"roundedOff"`
	TotalTax      float64               `json:"totalTax"`
	Total         float64               `json:"totalAmount"`
	TotalInWords  string                `json:"amountInWords"`
	Products      []models.InvoiceProduct `json:"products"`
}

func parseInvoiceDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, errors.New("invoiceDate is required")
	}

	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, errors.New("invalid invoiceDate format; use YYYY-MM-DD or RFC3339")
}

func parseOptionalDate(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	layouts := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			result := parsed
			return &result, nil
		}
	}

	return nil, errors.New("invalid poDate format; use YYYY-MM-DD or RFC3339")
}

func toInvoiceModel(payload invoicePayload) (models.Invoice, error) {
	invoiceDate, err := parseInvoiceDate(payload.InvoiceDate)
	if err != nil {
		return models.Invoice{}, err
	}

	poDate, err := parseOptionalDate(payload.PODate)
	if err != nil {
		return models.Invoice{}, err
	}

	poNumber := strings.TrimSpace(payload.PONumber)
	if poNumber == "" {
		poNumber = strings.TrimSpace(payload.PONumberAlt)
	}

	return models.Invoice{
		InvoiceNumber: payload.InvoiceNumber,
		InvoiceDate:   invoiceDate,
		PONumber:      poNumber,
		PODate:        poDate,
		InvoiceId:     payload.InvoiceId,
		CompanyId:     payload.CompanyId,
		CustomerId:    payload.CustomerId,
		CustomerName:  payload.CustomerName,
		Address:       payload.Address,
		GSTIN:         payload.GSTIN,
		PaymentTerms:  payload.PaymentTerms,
		Amount:        payload.Amount,
		OverallDiscount: payload.OverallDiscount,
		CGST:          payload.CGST,
		SGST:          payload.SGST,
		IGST:          payload.IGST,
		RoundedOff:    payload.RoundedOff,
		TotalTax:      payload.TotalTax,
		Total:         payload.Total,
		TotalInWords:  payload.TotalInWords,
		Products:      payload.Products,
	}, nil
}

func GetInvoices(c *gin.Context) {
	companyID := strings.TrimSpace(c.Query("companyId"))
	invoices, err := repository.GetAllInvoices(companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, invoices)
}

func GetInvoiceByID(c *gin.Context) {
	invoiceID := c.Param("id")
	invoice, err := repository.GetInvoiceByID(invoiceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, invoice)
}

func CreateInvoice(c *gin.Context) {
	email, ok := userEmail(c)
	if !ok {
		return
	}

	var payload invoicePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		logOperationError("create invoice bind payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := toInvoiceModel(payload)
	if err != nil {
		logOperationError("create invoice payload validation", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := repository.CreateInvoice(email, invoice)
	if err != nil {
		logOperationError("create invoice number="+invoice.InvoiceNumber+" email="+email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logOperationSuccess("create invoice id=" + created.InvoiceId + " number=" + created.InvoiceNumber + " email=" + email)
	c.JSON(http.StatusCreated, created)
}

func UpdateInvoice(c *gin.Context) {
	invoiceID := c.Param("id")
	email, ok := userEmail(c)
	if !ok {
		return
	}

	var payload invoicePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		logOperationError("update invoice bind payload id="+invoiceID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := toInvoiceModel(payload)
	if err != nil {
		logOperationError("update invoice payload validation id="+invoiceID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := repository.UpdateInvoice(invoiceID, invoice, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logOperationError("update invoice not found id="+invoiceID, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		} else {
			logOperationError("update invoice id="+invoiceID+" email="+email, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	logOperationSuccess("update invoice id=" + updated.InvoiceId + " number=" + updated.InvoiceNumber + " email=" + email)
	c.JSON(http.StatusOK, updated)
}

func DeleteInvoice(c *gin.Context) {
	invoiceID := c.Param("id")
	err := repository.DeleteInvoice(invoiceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "invoice deactivated successfully"})
}
