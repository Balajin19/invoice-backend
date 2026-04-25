package routes

import (
	handler "invoice-generator-backend/internal/handlers"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func isAllowedOrigin(origin string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return false
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return false
	}

	if host == "localhost" || host == "127.0.0.1" {
		return true
	}

	// Allow all other web origins (public URLs)
	return true
}

func SetupRoutes() *gin.Engine {
	r := gin.Default()
	trustedProxies := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES"))
	if trustedProxies == "" {
		_ = r.SetTrustedProxies(nil)
	} else {
		parts := strings.Split(trustedProxies, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		_ = r.SetTrustedProxies(parts)
	}
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Use(cors.New(cors.Config{
		AllowOriginFunc:  isAllowedOrigin,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}))

	// Auth
	r.POST("/auth/register", handler.Register)
	r.POST("/auth/login", handler.Login)
	r.POST("/auth/forgot-password", handler.ForgotPassword)
	r.POST("/auth/reset-password", handler.ResetPassword)

	protected := r.Group("/")
	protected.Use(handler.AuthRequired())
	protected.GET("/user", handler.GetCurrentUser)
	protected.POST("/user/change-password", handler.ChangePassword)
	// Company settings
	protected.GET("/companies", handler.GetCompanyDetails)
	protected.POST("/companies", handler.CreateCompany)
	protected.GET("/company/:id", handler.GetCompanyDetailsByID)
	protected.PUT("/company/:id", handler.UpdateCompany)
	protected.DELETE("/company/:id", handler.DeleteCompany)
	// Bank settings
	protected.GET("/banks", handler.GetBankDetails)
	protected.POST("/banks", handler.CreateBank)
	protected.GET("/bank/:id", handler.GetBankDetailsByID)
	protected.PUT("/bank/:id", handler.UpdateBank)
	protected.DELETE("/bank/:id", handler.DeleteBank)
	// Invoice settings
	protected.GET("/invoice/setting", handler.GetInvoiceSettings)
	protected.GET("/invoice/setting/:id", handler.GetInvoiceSettingsByID)
	protected.POST("/invoice/setting", handler.CreateInvoiceSettings)
	protected.PUT("/invoice/setting/:id", handler.UpdateInvoiceSettings)
	protected.DELETE("/invoice/setting/:id", handler.DeleteInvoiceSettings)

	// Customers
	protected.GET("/customers", handler.GetCustomers)
	protected.POST("/customers", handler.CreateCustomer)
	protected.GET("/customer/:id", handler.GetCustomerByID)
	protected.PUT("/customer/:id", handler.UpdateCustomer)
	protected.DELETE("/customer/:id", handler.DeleteCustomer)

	// Products
	protected.GET("/products", handler.GetProducts)
	protected.POST("/products", handler.CreateProduct)
	protected.GET("/product/:id", handler.GetProductByID)
	protected.PUT("/product/:id", handler.UpdateProduct)
	protected.DELETE("/product/:id", handler.DeleteProduct)

	// Categories
	protected.GET("/categories", handler.GetCategories)
	protected.POST("/categories", handler.CreateCategory)
	protected.GET("/category/:id", handler.GetCategoryByID)
	protected.PUT("/category/:id", handler.UpdateCategory)
	protected.DELETE("/category/:id", handler.DeleteCategory)

	// Units
	protected.GET("/units", handler.GetUnits)
	protected.POST("/units", handler.CreateUnit)
	protected.GET("/unit/:id", handler.GetUnitByID)
	protected.PUT("/unit/:id", handler.UpdateUnit)
	protected.DELETE("/unit/:id", handler.DeleteUnit)

	// Invoices
	protected.GET("/invoices", handler.GetInvoices)
	protected.POST("/invoices", handler.CreateInvoice)
	protected.GET("/invoice/:id", handler.GetInvoiceByID)
	protected.GET("/invoice/:id/pdf", handler.BuildInvoicePDF)
	protected.PUT("/invoice/:id", handler.UpdateInvoice)
	protected.DELETE("/invoice/:id", handler.DeleteInvoice)

	return r
}