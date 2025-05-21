package services

import (
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// PDFService handles PDF generation
type PDFService struct{}

// NewPDFService creates a new PDFService
func NewPDFService() *PDFService {
	return &PDFService{}
}

// GenerateConcursosPDF generates a PDF with concursos information
func (s *PDFService) GenerateConcursosPDF(items []struct {
	Referencia string
	Entidade   string
	Objeto     int
	Data       string
	Hora       string
	Tipo       string
}) (*gofpdf.Fpdf, error) {
	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// PDF settings
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Concursos Futuros")
	pdf.Ln(12)

	// Table header
	pdf.SetFont("Arial", "B", 10)
	widths := []float64{40, 40, 20, 30, 20, 30}
	headers := []string{"REF.", "ENTIDADE", "OBJETO", "DATA", "HORA", "TIPO"}

	for i, header := range headers {
		pdf.CellFormat(widths[i], 10, header, "1", 0, "", false, 0, "")
	}
	pdf.Ln(-1)

	// Table data
	pdf.SetFont("Arial", "", 10)
	for _, item := range items {
		objetoStr := map[int]string{
			1: "CTE",
			2: "CON",
			3: "INF",
			4: "CI",
			5: "ROB",
		}[item.Objeto]

		pdf.CellFormat(widths[0], 10, item.Referencia, "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[1], 10, item.Entidade, "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[2], 10, objetoStr, "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[3], 10, item.Data, "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[4], 10, item.Hora, "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[5], 10, item.Tipo, "1", 0, "", false, 0, "")
		pdf.Ln(-1)
	}

	// Footer
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, fmt.Sprintf("Atualizado em: %s", time.Now().Format("2006-01-02 15:04:05")))

	return pdf, nil
}
