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

func companyContactInfo(company *models.Companies) string {
	if company == nil {
		return ""
	}

	contactInfo := ""
	if strings.TrimSpace(company.Email) != "" {
		contactInfo = "Email: " + strings.TrimSpace(company.Email)
	}
	if strings.TrimSpace(company.Website) != "" {
		if contactInfo != "" {
			contactInfo += " | "
		}
		contactInfo += "Website: " + strings.TrimSpace(company.Website)
	}

	return contactInfo
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
	orientation := parsePDFOrientation(c.Query("orientation"))
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

	pdfBytes, err := buildInvoicePDF(invoice, pdfData, orientation)
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
	headers := []string{"S.No", "Description of Products", "HSN/SAC", "Unit", "Qty", "Rate", "Disc%", "Amount", "CGST%", "SGST%", "IGST%"}

	pdf.SetFont("Arial", "B", 9)
	for i, h := range headers {
		pdf.CellFormat(widths[i], 7, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 9)
}

func parsePDFOrientation(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "l", "landscape", "horizontal":
		return "L"
	default:
		return "P"
	}
}

func scaleWidthsToContent(base []float64, targetTotal float64) []float64 {
	scaled := make([]float64, len(base))
	if len(base) == 0 {
		return scaled
	}

	baseTotal := 0.0
	for _, w := range base {
		baseTotal += w
	}
	if baseTotal <= 0 {
		copy(scaled, base)
		return scaled
	}

	factor := targetTotal / baseTotal
	sumScaled := 0.0
	for i, w := range base {
		scaled[i] = w * factor
		sumScaled += scaled[i]
	}

	// Keep exact total width to avoid overflow from floating-point drift.
	scaled[len(scaled)-1] += targetTotal - sumScaled
	return scaled
}

func buildInvoicePDF(invoice *models.Invoice, data *invoicePDFData, orientation string) ([]byte, error) {

	pdf := fpdf.New(orientation, "mm", "A4", "")
	pdf.SetMargins(8, 8, 8)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	pageWidth, pageHeight := pdf.GetPageSize()
	contentWidth := pageWidth - 16
	isLandscape := strings.EqualFold(orientation, "L")

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
	contactText := companyContactInfo(data.Company)
	addressWidth := 110.0
	addressX := (pageWidth - addressWidth) / 2
	companyNameTop := headerTop + 3
	addressTop := companyNameTop + 6

	leftLineHeight := 4.5
	addressLineHeight := 4.5
	contactLineHeight := 3.5
	topPadding := 2.5
	bottomPadding := 2.5

	leftLines := pdf.SplitLines([]byte(leftText), 60)
	addrLines := pdf.SplitLines([]byte(addressText), addressWidth)
	contactLines := pdf.SplitLines([]byte(contactText), addressWidth)

	leftBottom := headerTop + topPadding + float64(len(leftLines))*leftLineHeight
	addressBottom := addressTop + float64(len(addrLines))*addressLineHeight
	contactTop := addressBottom + 1.0
	contactBottom := addressBottom
	if strings.TrimSpace(contactText) != "" {
		contactBottom = contactTop + float64(len(contactLines))*contactLineHeight
	}
	logoBottom := headerTop + 2 + 10 // approximate logo visual footprint for box fitting

	headerBottom := leftBottom
	if addressBottom > headerBottom {
		headerBottom = addressBottom
	}
	if contactBottom > headerBottom {
		headerBottom = contactBottom
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

	if strings.TrimSpace(contactText) != "" {
		pdf.SetXY(addressX, contactTop)
		pdf.SetFont("Arial", "", 8)
		pdf.MultiCell(addressWidth, contactLineHeight, contactText, "", "C", false)
	}

	// Logo
	if data.LogoPath != "" {
		pdf.Image(data.LogoPath, pageWidth-36, headerTop+3.5, 28, 0, false, "", 0, "")
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
	widths := []float64{9, 49, 19, 12, 13, 18, 12, 23, 13, 13, 13}
	if isLandscape {
		widths = scaleWidthsToContent(widths, contentWidth)
	}

drawProductTableHeader(pdf, widths)
productHeaderY := pdf.GetY()

var totalQty float64
var lineHeight float64 = 7.5
var multiLineHeight float64 = 6.0
leftMargin := 8.0
totalRowHeight := 7.0

_, pageHeight = pdf.GetPageSize()
bottomMargin := 8.0

descriptions := make([]string, len(invoice.Products))
lineCounts := make([]int, len(invoice.Products))
rowHeights := make([]float64, len(invoice.Products))

for i, p := range invoice.Products {
	description := strings.TrimSpace(strings.ReplaceAll(p.ProductName, "\n", " "))
	if description == "" {
		description = "-"
	}
	productLines := pdf.SplitLines([]byte(description), widths[1])
	lineCount := len(productLines)
	if lineCount == 0 {
		lineCount = 1
	}
	lineUnitHeight := lineHeight
	if lineCount > 1 {
		lineUnitHeight = multiLineHeight
	}
	rowHeight := float64(lineCount) * lineUnitHeight
	if rowHeight < 6 {
		rowHeight = 6
	}

	descriptions[i] = description
	lineCounts[i] = lineCount
	rowHeights[i] = rowHeight
}

remainingRowHeights := make([]float64, len(rowHeights)+1)
for i := len(rowHeights) - 1; i >= 0; i-- {
	remainingRowHeights[i] = remainingRowHeights[i+1] + rowHeights[i]
}

landscapeFooterReserve := 0.0
if isLandscape {
	termsWForFooter := contentWidth / 2
	termsTextWidthForFooter := termsWForFooter - 5

	termsLinesForFooter := pdf.SplitLines([]byte(data.Terms), termsTextWidthForFooter)
	termsHeightForFooter := float64(len(termsLinesForFooter))*5.5 + 11
	boxHeightForFooter := termsHeightForFooter
	if 7.0*7.0 > boxHeightForFooter {
		boxHeightForFooter = 7.0 * 7.0
	}

	amountWordsLabelForFooter := "Amount in Words: "
	amountWordsValueForFooter := strings.TrimSpace(invoice.TotalInWords)
	if amountWordsValueForFooter == "" {
		amountWordsValueForFooter = "-"
	}
	amountWordsTextForFooter := amountWordsLabelForFooter + amountWordsValueForFooter
	amountWordsLinesForFooter := pdf.SplitLines([]byte(amountWordsTextForFooter), contentWidth-2)
	amountWordsHeightForFooter := 7.0
	if len(amountWordsLinesForFooter) > 1 {
		amountWordsHeightForFooter = float64(len(amountWordsLinesForFooter)) * 5.5
	}

	bankRowsCountForFooter := 5
	upiValueForFooter := strings.TrimSpace(bankField(data.Bank, func(bank *models.Banks) string { return bank.UPI }))
	if upiValueForFooter != "" {
		bankRowsCountForFooter++
	}
	bankHeightForFooter := float64(bankRowsCountForFooter+1)*5.5 + 7

	landscapeFooterGap := 10.0
	landscapeFooterReserve = boxHeightForFooter + amountWordsHeightForFooter + bankHeightForFooter + landscapeFooterGap
}

// ---- PRODUCT LOOP ----
for i, p := range invoice.Products {

	totalQty += p.Qty

	description := descriptions[i]
	rowHeight := rowHeights[i]
	descLineHeight := rowHeight / float64(lineCounts[i])

	// Keep product rows visible and reserve footer space on the last product page.
	breakLimit := pageHeight - bottomMargin
	rowRequiredHeight := rowHeight
	if isLandscape {
		remainingWithTotalAndFooter := remainingRowHeights[i] + totalRowHeight + landscapeFooterReserve
		// Force a break only when all remaining content (rows+total+footer) fits on a
		// fresh page but cannot fit on the current page — ensures last page groups
		// products together with the footer.
		if productHeaderY+remainingWithTotalAndFooter <= pageHeight-bottomMargin &&
			pdf.GetY()+remainingWithTotalAndFooter > pageHeight-bottomMargin {
			pdf.AddPage()
			drawProductTableHeader(pdf, widths)
		}
	}

	if pdf.GetY()+rowRequiredHeight > breakLimit {
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
	pdf.MultiCell(widths[1], descLineHeight, description, "1", "L", false)

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
	pdf.CellFormat(widths[7], rowHeight, fmt.Sprintf("%.2f", p.Total), "1", 0, "R", false, 0, "")
	currentX += widths[7]

	// CGST%
	pdf.SetXY(currentX, rowStartY)
	pdf.CellFormat(widths[8], rowHeight, fmt.Sprintf("%.2f", p.CGSTRate), "1", 0, "C", false, 0, "")
	currentX += widths[8]

	// SGST%
	pdf.SetXY(currentX, rowStartY)
	pdf.CellFormat(widths[9], rowHeight, fmt.Sprintf("%.2f", p.SGSTRate), "1", 0, "C", false, 0, "")
	currentX += widths[9]

	// IGST%
	pdf.SetXY(currentX, rowStartY)
	pdf.CellFormat(widths[10], rowHeight, fmt.Sprintf("%.2f", p.IGSTRate), "1", 1, "C", false, 0, "")
}

// ---- TOTAL ROW ----
totalBreakLimit := pageHeight - bottomMargin
if isLandscape {
	totalBreakLimit = pageHeight - bottomMargin - landscapeFooterReserve
}

if pdf.GetY()+totalRowHeight > totalBreakLimit {
	pdf.AddPage()
	drawProductTableHeader(pdf, widths)
}

pdf.SetFont("Arial", "B", 9)
pdf.SetX(leftMargin)
// Empty cells for S.No through Unit
emptyWidth := widths[0] + widths[1] + widths[2] + widths[3]
pdf.CellFormat(emptyWidth, totalRowHeight, "", "1", 0, "", false, 0, "")
pdf.CellFormat(widths[4], totalRowHeight, fmt.Sprintf("%.0f", totalQty), "1", 0, "C", false, 0, "")
pdf.CellFormat(widths[5]+widths[6], totalRowHeight, "Total Amount", "1", 0, "R", false, 0, "")
pdf.CellFormat(widths[7], totalRowHeight, fmt.Sprintf("%.2f", invoice.Amount), "1", 0, "R", false, 0, "")
emptyTaxWidth := widths[8] + widths[9] + widths[10]
pdf.CellFormat(emptyTaxWidth, totalRowHeight, "", "1", 1, "", false, 0, "")

_, pageHeight = pdf.GetPageSize()
bottomMargin = 8.0

termsX := 8.0
termsW := 97.0
gstX := 105.0
gstW := 97.0
if isLandscape {
	termsW = contentWidth / 2
	gstW = contentWidth - termsW
	gstX = termsX + termsW
}

termsTextWidth := 92.0
if isLandscape {
	termsTextWidth = termsW - 5
}

upiValue := strings.TrimSpace(bankField(data.Bank, func(bank *models.Banks) string { return bank.UPI }))

bankRows := []struct {
	label string
	value string
}{
	{"Account Name", bankField(data.Bank, func(bank *models.Banks) string { return bank.AccountName })},
	{"Account Number", bankField(data.Bank, func(bank *models.Banks) string { return bank.AccountNumber })},
	{"Bank Name", bankField(data.Bank, func(bank *models.Banks) string { return bank.BankName })},
	{"Branch", bankField(data.Bank, func(bank *models.Banks) string { return bank.BranchName })},
	{"IFSC", bankField(data.Bank, func(bank *models.Banks) string { return bank.IFSC })},
}
if upiValue != "" {
	bankRows = append(bankRows, struct {
		label string
		value string
	}{"UPI", upiValue})
}

bankHeight := float64(len(bankRows)+1)*5.5 + 7

termsLines := pdf.SplitLines([]byte(data.Terms), termsTextWidth)
termsHeight := float64(len(termsLines))*5.5 + 11
gstHeight := 7.0 * 7.0
boxHeight := termsHeight
if gstHeight > boxHeight {
	boxHeight = gstHeight
}

amountWordsLabel := "Amount in Words: "
amountWordsValue := strings.TrimSpace(invoice.TotalInWords)
if amountWordsValue == "" {
	amountWordsValue = "-"
}
amountWordsText := amountWordsLabel + amountWordsValue
amountWordsLines := pdf.SplitLines([]byte(amountWordsText), contentWidth-2)
amountWordsHeight := 7.0
if len(amountWordsLines) > 1 {
	amountWordsHeight = float64(len(amountWordsLines)) * 5.5
}

// Use available space for each footer section in both orientations.
if pdf.GetY()+boxHeight > pageHeight-bottomMargin {
	pdf.AddPage()
}
	// ================= TERMS + GST =================

	y := pdf.GetY()

	termsLines = pdf.SplitLines([]byte(data.Terms), termsTextWidth)
	termsHeight = float64(len(termsLines))*5.5 + 11

	gstHeight = 7.0 * 7.0

	boxHeight = termsHeight
	if gstHeight > boxHeight {
		boxHeight = gstHeight
	}

	pdf.Rect(termsX, y, termsW, boxHeight, "D")
	pdf.SetXY(termsX+2, y+2)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(termsTextWidth, 6, "Terms & Conditions:")
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	pdf.MultiCell(termsTextWidth, 5.5, data.Terms, "", "L", false)

	pdf.Rect(gstX, y, gstW, boxHeight, "D")

	labelW := 47.0
	valueW := 50.0
	if isLandscape {
		labelW = gstW * 0.4845
		valueW = gstW - labelW
	}

	rows := []struct {
		label string
		value float64
	}{
		{"Overall Discount", invoice.OverallDiscount},
		{"CGST", invoice.CGST},
		{"SGST", invoice.SGST},
		{"IGST", invoice.IGST},
		{"Total Tax", invoice.TotalTax},
		{"Rounded Off", invoice.RoundedOff},
		{"Gross Total", invoice.Total},
	}

	pdf.SetXY(gstX, y)

	for _, r := range rows {
		isBold := r.label == "Gross Total" || r.label == "Total Tax" || r.label == "Overall Discount"

		if isBold {
			pdf.SetFont("Arial", "B", 9)
		} else {
			pdf.SetFont("Arial", "", 9)
		}

		pdf.SetX(gstX)
		pdf.CellFormat(labelW, 7, r.label, "1", 0, "", false, 0, "")

		if isBold {
			pdf.SetFont("Arial", "B", 9)
		} else {
			pdf.SetFont("Arial", "", 9)
		}

		pdf.CellFormat(valueW, 7, fmt.Sprintf("%.2f", r.value), "1", 1, "R", false, 0, "")
	}

	pdf.SetY(y + boxHeight)

	// ================= AMOUNT WORDS =================
	if pdf.GetY()+amountWordsHeight > pageHeight-bottomMargin {
		pdf.AddPage()
	}

	if len(amountWordsLines) <= 1 {
		pdf.SetFont("Arial", "B", 9)
		labelWidth := pdf.GetStringWidth(amountWordsLabel) + 2
		if labelWidth > contentWidth-20 {
			labelWidth = contentWidth * 0.35
		}
		wordsY := pdf.GetY()
		pdf.Rect(8, wordsY, contentWidth, 7, "D")
		pdf.CellFormat(labelWidth, 7, amountWordsLabel, "", 0, "L", false, 0, "")

		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(contentWidth-labelWidth, 7, amountWordsValue, "", 1, "L", false, 0, "")
	} else {
		pdf.SetFont("Arial", "", 9)
		pdf.MultiCell(contentWidth, 5.5, amountWordsText, "1", "L", false)
	}

	// ================= BANK =================
	if pdf.GetY()+bankHeight > pageHeight-bottomMargin {
		pdf.AddPage()
	}

	y = pdf.GetY()

	pdf.Rect(termsX, y, termsW, bankHeight, "D")

	pdf.SetXY(termsX+2, y+2)

	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(termsTextWidth, 6, "Bank Details:")
	pdf.Ln(6)

	bankLabelW := 28.0
	bankValueW := 64.0
	if isLandscape {
		bankLabelW = termsTextWidth * 0.3043
		bankValueW = termsTextWidth - bankLabelW
	}
	for _, row := range bankRows {
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(bankLabelW, 5.5, row.label+":", "", 0, "", false, 0, "")
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(bankValueW, 5.5, row.value, "", 0, "", false, 0, "")
		pdf.Ln(5.5)
	}

	pdf.Rect(gstX, y, gstW, bankHeight, "D")

	pdf.SetXY(gstX, y+10)
	prefix := "For "
	companyName := companyField(data.Company, func(company *models.Companies) string { return company.CompanyName })
	rightBlockPadding := 8.0
	pdf.SetFont("Arial", "", 9)
	prefixWidth := pdf.GetStringWidth(prefix)
	pdf.SetFont("Arial", "B", 9)
	companyNameWidth := pdf.GetStringWidth(companyName)
	lineStartX := gstX + gstW - rightBlockPadding - (prefixWidth + companyNameWidth)

	pdf.SetXY(lineStartX, y+10)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(prefixWidth, 5, prefix, "", 0, "", false, 0, "")
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(companyNameWidth, 5, companyName, "", 0, "", false, 0, "")

	pdf.SetXY(gstX, y+bankHeight-8)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(gstW-rightBlockPadding, 5, "Authorised Signatory", "", 0, "R", false, 0, "")

	pdf.SetY(y + bankHeight)

	// ================= OUTPUT =================

	var out bytes.Buffer
	err := pdf.Output(&out)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}