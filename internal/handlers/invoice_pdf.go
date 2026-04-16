package handler

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"invoice-generator-backend/internal/models"
	"invoice-generator-backend/repository"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
)

type invoicePDFData struct {
	Company         *models.Companies
	Bank            *models.Banks
	Settings        *models.InvoiceSettings
	CompanyAddress  string
	BuyerDetails    string
	ConsigneeDetails string
	Terms           string
	LogoPath        string
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func joinAddress(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	return strings.Join(filtered, ", ")
}

func companyAddress(company *models.Companies) string {
	if company == nil {
		return ""
	}
	return joinAddress(company.BuildingNumber, company.Street, company.City, company.District, company.State, company.Pincode)
}

func companyField(company *models.Companies, getter func(*models.Companies) string) string {
	if company == nil {
		return ""
	}
	return getter(company)
}

func bankField(bank *models.Banks, getter func(*models.Banks) string) string {
	if bank == nil {
		return ""
	}
	return getter(bank)
}

func customerInvoiceDetails(invoice *models.Invoice) string {
	return strings.TrimSpace(fmt.Sprintf("%s\n%s\nGSTIN: %s", invoice.CustomerName, invoice.Address, firstNonEmpty(invoice.GSTIN, "-")))
}

func buildTermsConditions(template, paymentTerms string) string {
	termsTemplate := strings.TrimSpace(template)
	if termsTemplate == "" {
		return ""
	}

	lines := strings.Split(termsTemplate, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "payment terms") || strings.Contains(lower, "{{payment_terms}}") {
			continue
		}
		filtered = append(filtered, trimmed)
	}

	paymentTermsLine := fmt.Sprintf("%d. Payment Terms : {{payment_terms}}", len(filtered)+1)
	termsWithPayment := strings.Join(append(filtered, paymentTermsLine), "\n")

	paymentValue := strings.TrimSpace(paymentTerms)
	if paymentValue == "" {
		paymentValue = "0"
	}

	return strings.ReplaceAll(termsWithPayment, "{{payment_terms}}", fmt.Sprintf("Within %s days", paymentValue))
}

func selectCompany(companies []models.Companies, userEmail string) *models.Companies {
	filtered := make([]models.Companies, 0)
	for _, company := range companies {
		if strings.EqualFold(strings.TrimSpace(company.CreatedBy), strings.TrimSpace(userEmail)) {
			filtered = append(filtered, company)
		}
	}
	if len(filtered) == 0 {
		filtered = companies
	}
	if len(filtered) == 0 {
		return nil
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].IsPrimary != filtered[j].IsPrimary {
			return filtered[i].IsPrimary
		}
		return filtered[i].CompanyName < filtered[j].CompanyName
	})
	selected := filtered[0]
	return &selected
}

func selectBank(banks []models.Banks, userEmail string) *models.Banks {
	filtered := make([]models.Banks, 0)
	for _, bank := range banks {
		if strings.EqualFold(strings.TrimSpace(bank.CreatedBy), strings.TrimSpace(userEmail)) {
			filtered = append(filtered, bank)
		}
	}
	if len(filtered) == 0 {
		filtered = banks
	}
	if len(filtered) == 0 {
		return nil
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].IsPrimary != filtered[j].IsPrimary {
			return filtered[i].IsPrimary
		}
		return filtered[i].BankName < filtered[j].BankName
	})
	selected := filtered[0]
	return &selected
}

func buildInvoicePDFData(userEmail string, invoice *models.Invoice) (*invoicePDFData, error) {
	companies, err := repository.GetCompanySettings()
	if err != nil {
		return nil, err
	}
	banks, err := repository.GetBankSettings()
	if err != nil {
		return nil, err
	}

	company := selectCompany(companies, userEmail)
	bank := selectBank(banks, userEmail)

	var settings *models.InvoiceSettings
	if company != nil {
		settings, err = repository.GetInvoiceSettingsByCompanyIDSingle(company.CompanyID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if errors.Is(err, sql.ErrNoRows) {
			settings = nil
		}
	}

	termsTemplate := ""
	if settings != nil {
		termsTemplate = settings.TermsConditions
	}
	terms := buildTermsConditions(termsTemplate, invoice.PaymentTerms)

	return &invoicePDFData{
		Company:          company,
		Bank:             bank,
		Settings:         settings,
		CompanyAddress:   companyAddress(company),
		BuyerDetails:     customerInvoiceDetails(invoice),
		ConsigneeDetails: customerInvoiceDetails(invoice),
		Terms:            terms,
		LogoPath:         filepath.Clean("./assets/sk-logo.png"),
	}, nil
}

func BuildInvoicePDF(c *gin.Context) {
	invoiceID := c.Param("id")
	userEmail, ok := userEmail(c)
	if !ok {
		return
	}

	invoice, err := repository.GetInvoiceByID(invoiceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	pdfData, err := buildInvoicePDFData(userEmail, invoice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pdfBytes, err := buildInvoicePDF(invoice, pdfData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filenameSanitizer := strings.NewReplacer("/", "_", "\\", "_", " ", "_")
	filename := fmt.Sprintf("%s.pdf", filenameSanitizer.Replace(invoice.InvoiceNumber))

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
func drawProductTableHeader(pdf *fpdf.Fpdf, widths []float64) {
	headers := []string{"S.No", "Description of Products", "HSN", "Unit", "Qty", "Rate", "Disc%", "Amount"}

	pdf.SetFont("Arial", "B", 9)
	for i, h := range headers {
		pdf.CellFormat(widths[i], 7, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 9)
}

func moveToBottomIfLastPage(pdf *fpdf.Fpdf, requiredHeight float64) {
	pageHeight, _ := pdf.GetPageSize()
	bottomMargin := 8.0

	currentY := pdf.GetY()
	remaining := pageHeight - bottomMargin - currentY

	if remaining > requiredHeight {
		pdf.SetY(pageHeight - bottomMargin - requiredHeight)
	}
}

func buildInvoicePDF(invoice *models.Invoice, data *invoicePDFData) ([]byte, error) {

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(8, 8, 8)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	pageWidth, _ := pdf.GetPageSize()
	contentWidth := pageWidth - 16

	// ================= HEADER =================

	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(contentWidth, 1.6, "TAX INVOICE", "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(contentWidth, 3.4, "Original Receipt", "", 1, "R", false, 0, "")
	headerTop := pdf.GetY()

	leftText := fmt.Sprintf("GSTIN: %s\nCT: %s\nMobile: %s",
		firstNonEmpty(companyField(data.Company, func(company *models.Companies) string { return company.GSTIN }), "-"),
		firstNonEmpty(companyField(data.Company, func(company *models.Companies) string { return company.OwnerName }), "-"),
		firstNonEmpty(companyField(data.Company, func(company *models.Companies) string { return company.MobileNumber }), "-"))

	addressText := data.CompanyAddress
	addressWidth := 110.0
	addressX := (pageWidth - addressWidth) / 2
	companyNameTop := headerTop + 3
	addressTop := companyNameTop + 6

	leftLineHeight := 4.5
	addressLineHeight := 4.5
	topPadding := 2.5
	bottomPadding := 2.5

	leftLines := pdf.SplitLines([]byte(leftText), 60)
	addrLines := pdf.SplitLines([]byte(addressText), addressWidth)

	leftBottom := headerTop + topPadding + float64(len(leftLines))*leftLineHeight
	addressBottom := addressTop + float64(len(addrLines))*addressLineHeight
	logoBottom := headerTop + 2 + 10 // approximate logo visual footprint for box fitting

	headerBottom := leftBottom
	if addressBottom > headerBottom {
		headerBottom = addressBottom
	}
	if logoBottom > headerBottom {
		headerBottom = logoBottom
	}

	headerHeight := (headerBottom - headerTop) + bottomPadding

	pdf.Rect(8, headerTop, contentWidth, headerHeight, "D")

	// Left
	pdf.SetXY(10, headerTop+3.5)
	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(60, 4.5, leftText, "", "L", false)

	// Center name
	pdf.SetXY(0, companyNameTop)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(pageWidth, 6, companyField(data.Company, func(company *models.Companies) string { return company.CompanyName }), "", 0, "C", false, 0, "")

	// Address
	pdf.SetXY(addressX, addressTop)
	pdf.SetFont("Arial", "", 9)
	pdf.MultiCell(addressWidth, 4.5, addressText, "", "C", false)

	// Logo
	if data.LogoPath != "" {
		pdf.Image(data.LogoPath, pageWidth-33, headerTop+3.5, 25, 0, false, "", 0, "")
	}

	pdf.SetY(headerTop + headerHeight)

	// ================= BUYER / CONSIGNEE =================

	buyerCol := contentWidth * 0.35
	consigneeCol := contentWidth * 0.35
	invoiceCol := contentWidth - buyerCol - consigneeCol

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(buyerCol, 7.5, "Details of Buyer", "1", 0, "", false, 0, "")
	pdf.CellFormat(consigneeCol, 7.5, "Details of Consignee", "1", 0, "", false, 0, "")
	pdf.CellFormat(invoiceCol, 7.5, "Invoice Details", "1", 1, "", false, 0, "")

	yStart := pdf.GetY()

	// Buyer
	pdf.SetXY(8, yStart+1.5)
	pdf.SetFont("Arial", "B", 10)
	pdf.MultiCell(buyerCol, 5, invoice.CustomerName, "", "L", false)

	pdf.SetFont("Arial", "", 9)
	pdf.SetX(8)
	pdf.MultiCell(buyerCol, 5, invoice.Address, "", "L", false)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetX(8)
	pdf.MultiCell(buyerCol, 5, "GSTIN: "+invoice.GSTIN, "", "L", false)
	pdf.Ln(1.5)

	y1 := pdf.GetY()

	// Consignee
	pdf.SetXY(8+buyerCol, yStart+1.5)
	pdf.SetFont("Arial", "B", 10)
	pdf.MultiCell(consigneeCol, 5, invoice.CustomerName, "", "L", false)

	pdf.SetFont("Arial", "", 9)
	pdf.SetX(8 + buyerCol)
	pdf.MultiCell(consigneeCol, 5, invoice.Address, "", "L", false)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetX(8 + buyerCol)
	pdf.MultiCell(consigneeCol, 5, "GSTIN: "+invoice.GSTIN, "", "L", false)
	pdf.Ln(1.5)

	y2 := pdf.GetY()

	// Invoice Details
	detailsX := 8 + buyerCol + consigneeCol
	detailsY := yStart + 1.5
	poValue := strings.TrimSpace(invoice.PONumber)
	if poValue == "" {
		poValue = "-"
	}
	poDateValue := "-"
	if invoice.PODate != nil && !invoice.PODate.IsZero() {
		poDateValue = invoice.PODate.Format("02-01-2006")
	}
	drawDetailValue := func(label, value string) {
		labelText := label + ": "
		pdf.SetXY(detailsX, detailsY)
		pdf.SetFont("Arial", "", 9)
		labelWidth := pdf.GetStringWidth(labelText)
		pdf.CellFormat(labelWidth, 5, labelText, "", 0, "", false, 0, "")
		pdf.SetFont("Arial", "B", 9)
		pdf.MultiCell(invoiceCol-labelWidth, 5, value, "", "L", false)
		detailsY = pdf.GetY()
	}

	drawDetailValue("Invoice Number", invoice.InvoiceNumber)
	drawDetailValue("Date", invoice.InvoiceDate.Format("02-01-2006"))
	drawDetailValue("PO", poValue)
	drawDetailValue("PO Date", poDateValue)

	y3 := pdf.GetY()
	pdf.Ln(1.5)

	maxY := y1
	if y2 > maxY {
		maxY = y2
	}
	if y3 > maxY {
		maxY = y3
	}

	pdf.Rect(8, yStart, buyerCol, maxY-yStart, "D")
	pdf.Rect(8+buyerCol, yStart, consigneeCol, maxY-yStart, "D")
	pdf.Rect(8+buyerCol+consigneeCol, yStart, invoiceCol, maxY-yStart, "D")

	pdf.SetY(maxY)

	// ================= PRODUCTS =================
widths := []float64{10, 59, 30, 15, 15, 20, 15, 30}

drawProductTableHeader(pdf, widths)

var totalQty float64
lineHeight := 6.0
leftMargin := 8.0
totalRowHeight := 7.0

pageHeight, _ := pdf.GetPageSize()
bottomMargin := 8.0

// ---- PRODUCT LOOP ----
for i, p := range invoice.Products {

	totalQty += p.Qty

	description := strings.TrimSpace(strings.ReplaceAll(p.ProductName, "\n", " "))
	productLines := pdf.SplitLines([]byte(description), widths[1])

	rowHeight := float64(len(productLines)) * lineHeight
	if rowHeight < 6 {
		rowHeight = 6
	}

	if pdf.GetY()+rowHeight > pageHeight-bottomMargin {
		pdf.AddPage()
		drawProductTableHeader(pdf, widths)
	}

	rowStartX := leftMargin
	rowStartY := pdf.GetY()

	// S.No
	pdf.SetXY(rowStartX, rowStartY)
	pdf.CellFormat(widths[0], rowHeight, fmt.Sprintf("%d", i+1), "1", 0, "CM", false, 0, "")

	// Description
	pdf.SetXY(rowStartX+widths[0], rowStartY)
	pdf.MultiCell(widths[1], lineHeight, description, "1", "L", false)

	currentX := rowStartX + widths[0] + widths[1]

	// HSN
	pdf.SetXY(currentX, rowStartY)
	pdf.CellFormat(widths[2], rowHeight, p.HSN, "1", 0, "", false, 0, "")
	currentX += widths[2]

	// Unit
	pdf.SetXY(currentX, rowStartY)
	pdf.CellFormat(widths[3], rowHeight, p.Unit, "1", 0, "C", false, 0, "")
	currentX += widths[3]

	// Qty
	pdf.SetXY(currentX, rowStartY)
	pdf.CellFormat(widths[4], rowHeight, fmt.Sprintf("%.0f", p.Qty), "1", 0, "C", false, 0, "")
	currentX += widths[4]

	// Rate
	pdf.SetXY(currentX, rowStartY)
	pdf.CellFormat(widths[5], rowHeight, fmt.Sprintf("%.2f", p.Price), "1", 0, "R", false, 0, "")
	currentX += widths[5]

	// Disc%
	pdf.SetXY(currentX, rowStartY)
	pdf.CellFormat(widths[6], rowHeight, fmt.Sprintf("%.2f", p.Discount), "1", 0, "C", false, 0, "")
	currentX += widths[6]

	// Amount
	pdf.SetXY(currentX, rowStartY)
	pdf.CellFormat(widths[7], rowHeight, fmt.Sprintf("%.2f", p.Total), "1", 1, "R", false, 0, "")
}

// ---- TOTAL ROW ----
if pdf.GetY()+totalRowHeight > pageHeight-bottomMargin {
	pdf.AddPage()
	drawProductTableHeader(pdf, widths)
}

pdf.SetFont("Arial", "B", 9)
pdf.SetX(leftMargin)
pdf.CellFormat(114, totalRowHeight, "", "1", 0, "", false, 0, "")
pdf.CellFormat(15, totalRowHeight, fmt.Sprintf("%.0f", totalQty), "1", 0, "C", false, 0, "")
pdf.CellFormat(35, totalRowHeight, "Total Amount", "1", 0, "R", false, 0, "")
pdf.CellFormat(30, totalRowHeight, fmt.Sprintf("%.2f", invoice.Amount), "1", 1, "R", false, 0, "")

pageHeight, _ = pdf.GetPageSize()
bottomMargin = 8.0

termsLines := pdf.SplitLines([]byte(data.Terms), 92)
termsHeight := float64(len(termsLines))*5.5 + 11
gstHeight := 6.0 * 7.0
boxHeight := termsHeight
if gstHeight > boxHeight {
	boxHeight = gstHeight
}

// Amount words
footerAmountWordsText := "Amount in Words: " + invoice.TotalInWords
footerAmountLines := pdf.SplitLines([]byte(footerAmountWordsText), contentWidth-2)
footerAmountHeight := 7.0
if len(footerAmountLines) > 1 {
	footerAmountHeight = float64(len(footerAmountLines)) * 5.5
}

// Bank
footerBankHeight := float64(6+len([]string{
	"Account Name", "Account Number", "Bank Name", "Branch", "IFSC", "UPI",
})) * 5.5 + 7

// TOTAL FOOTER HEIGHT
footerHeight := boxHeight + footerAmountHeight + footerBankHeight + 6

// SINGLE PAGE BREAK DECISION
if pdf.GetY()+footerHeight > pageHeight-bottomMargin {
	pdf.AddPage()
}
	// ================= TERMS + GST =================

	y := pdf.GetY()

	termsLines = pdf.SplitLines([]byte(data.Terms), 92)
	termsHeight = float64(len(termsLines))*5.5 + 11

	gstHeight = 6.0 * 7.0

	boxHeight = termsHeight
	if gstHeight > boxHeight {
		boxHeight = gstHeight
	}

	pdf.Rect(8, y, 97, boxHeight, "D")
	pdf.SetXY(10, y+2)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(92, 6, "Terms & Conditions:")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(92, 5.5, data.Terms, "", "L", false)

	pdf.Rect(105, y, 97, boxHeight, "D")

	labelW := 47.0
	valueW := 50.0

	rows := []struct {
		label string
		value float64
	}{
		{"CGST 9%", invoice.CGST},
		{"SGST 9%", invoice.SGST},
		{"IGST", invoice.IGST},
		{"Rounded Off", invoice.RoundedOff},
		{"Total Tax", invoice.TotalTax},
		{"Gross Total", invoice.Total},
	}

	pdf.SetXY(105, y)

	for _, r := range rows {
		isGrossTotal := r.label == "Gross Total"
		isTotalTax := r.label == "Total Tax"

		// Keep gross total label bold; other labels stay regular.
		if isGrossTotal {
			pdf.SetFont("Arial", "B", 9)
		} else {
			pdf.SetFont("Arial", "", 9)
		}

		pdf.SetX(105)
		pdf.CellFormat(labelW, 7, r.label, "1", 0, "", false, 0, "")

		// Make Total Tax value bold as requested; keep Gross Total bold too.
		if isGrossTotal || isTotalTax {
			pdf.SetFont("Arial", "B", 9)
		} else {
			pdf.SetFont("Arial", "", 9)
		}

		pdf.CellFormat(valueW, 7, fmt.Sprintf("%.2f", r.value), "1", 1, "R", false, 0, "")
	}

	pdf.SetY(y + boxHeight)

	// ================= AMOUNT WORDS =================

	amountWordsText := "Amount in Words: " + invoice.TotalInWords
	amountWordsLines := pdf.SplitLines([]byte(amountWordsText), contentWidth-2)

	if len(amountWordsLines) <= 1 {
		pdf.CellFormat(contentWidth, 7, amountWordsText, "1", 1, "", false, 0, "")
	} else {
		pdf.MultiCell(contentWidth, 5.5, amountWordsText, "1", "L", false)
	}

	// ================= BANK =================

	y = pdf.GetY()

	bankRows := []struct {
		label string
		value string
	}{
		{"Account Name", bankField(data.Bank, func(bank *models.Banks) string { return bank.AccountName })},
		{"Account Number", bankField(data.Bank, func(bank *models.Banks) string { return bank.AccountNumber })},
		{"Bank Name", bankField(data.Bank, func(bank *models.Banks) string { return bank.BankName })},
		{"Branch", bankField(data.Bank, func(bank *models.Banks) string { return bank.BranchName })},
		{"IFSC", bankField(data.Bank, func(bank *models.Banks) string { return bank.IFSC })},
		{"UPI", bankField(data.Bank, func(bank *models.Banks) string { return bank.UPI })},
	}

	bankHeight := float64(len(bankRows)+1)*5.5 + 7

	pdf.Rect(8, y, 97, bankHeight, "D")

	pdf.SetXY(10, y+2)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(92, 6, "Bank Details:")
	pdf.Ln(6)

	bankLabelW := 28.0
	bankValueW := 64.0
	for _, row := range bankRows {
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(bankLabelW, 5.5, row.label+":", "", 0, "", false, 0, "")
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(bankValueW, 5.5, row.value, "", 0, "", false, 0, "")
		pdf.Ln(5.5)
	}

	pdf.Rect(105, y, 97, bankHeight, "D")

	pdf.SetXY(105, y+10)
	prefix := "For "
	companyName := companyField(data.Company, func(company *models.Companies) string { return company.CompanyName })
	rightBlockPadding := 8.0
	pdf.SetFont("Arial", "", 9)
	prefixWidth := pdf.GetStringWidth(prefix)
	pdf.SetFont("Arial", "B", 9)
	companyNameWidth := pdf.GetStringWidth(companyName)
	lineStartX := 105 + 97 - rightBlockPadding - (prefixWidth + companyNameWidth)

	pdf.SetXY(lineStartX, y+10)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(prefixWidth, 5, prefix, "", 0, "", false, 0, "")
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(companyNameWidth, 5, companyName, "", 0, "", false, 0, "")

	pdf.SetXY(105, y+bankHeight-8)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(97-rightBlockPadding, 5, "Authorised Signatory", "", 0, "R", false, 0, "")

	pdf.SetY(y + bankHeight)

	// ================= OUTPUT =================

	var out bytes.Buffer
	err := pdf.Output(&out)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}