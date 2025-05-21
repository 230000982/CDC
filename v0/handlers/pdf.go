package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// PDFHandler handles PDF generation and download
type PDFHandler struct {
	db *sql.DB
}

// NewPDFHandler creates a new PDFHandler
func NewPDFHandler(db *sql.DB) *PDFHandler {
	return &PDFHandler{
		db: db,
	}
}

// Download handles the download PDF request
func (h *PDFHandler) Download(w http.ResponseWriter, r *http.Request) {
	// Get current date and time
	now := time.Now()
	currentDate := now.Format("2006-01-02")
	currentTime := now.Format("15:04:05")

	// Query concursos with estado_id = 1
	rows, err := h.db.Query(`
        SELECT c.id_concurso, c.entidade, c.estado_id,
               c.dia_erro, c.hora_erro, 
               c.dia_proposta, c.hora_proposta,
               c.dia_audiencia, c.hora_audiencia,
               c.tipo_id, c.referencia
        FROM concurso c
        WHERE c.estado_id = 1
        ORDER BY 
            COALESCE(c.dia_proposta, c.dia_erro, c.dia_audiencia) ASC,
            COALESCE(c.hora_proposta, c.hora_erro, c.hora_audiencia) ASC
    `)
	if err != nil {
		log.Printf("Error querying concursos: %v", err)
		http.Error(w, "Erro ao buscar concursos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ConcursoItem struct {
		Referencia string
		Entidade   string
		Objeto     int
		Data       string
		Hora       string
		Tipo       string
	}

	var items []ConcursoItem

	// Function to check if date is in the future
	isFutureDate := func(date, time string) bool {
		if date > currentDate {
			return true
		}
		if date == currentDate && time > currentTime {
			return true
		}
		return false
	}

	// Scan rows
	for rows.Next() {
		var id, estado, objeto int
		var entidade, referencia string
		var diaErro, horaErro sql.NullString
		var diaProposta, horaProposta sql.NullString
		var diaAudiencia, horaAudiencia sql.NullString

		err := rows.Scan(
			&id, &entidade, &estado,
			&diaErro, &horaErro,
			&diaProposta, &horaProposta,
			&diaAudiencia, &horaAudiencia,
			&objeto, &referencia,
		)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Add each valid date as a separate item, only if it's in the future
		if diaProposta.Valid && horaProposta.Valid && isFutureDate(diaProposta.String, horaProposta.String) {
			items = append(items, ConcursoItem{
				Referencia: referencia,
				Entidade:   entidade,
				Objeto:     objeto,
				Data:       diaProposta.String,
				Hora:       horaProposta.String,
				Tipo:       "Proposta",
			})
		}

		if diaErro.Valid && horaErro.Valid && isFutureDate(diaErro.String, horaErro.String) {
			items = append(items, ConcursoItem{
				Referencia: referencia,
				Entidade:   entidade,
				Objeto:     objeto,
				Data:       diaErro.String,
				Hora:       horaErro.String,
				Tipo:       "Erro",
			})
		}

		if diaAudiencia.Valid && horaAudiencia.Valid && isFutureDate(diaAudiencia.String, horaAudiencia.String) {
			items = append(items, ConcursoItem{
				Referencia: referencia,
				Entidade:   entidade,
				Objeto:     objeto,
				Data:       diaAudiencia.String,
				Hora:       horaAudiencia.String,
				Tipo:       "Audiencia",
			})
		}
	}

	// Sort items by date and time
	sort.Slice(items, func(i, j int) bool {
		dataHoraI := items[i].Data + " " + items[i].Hora
		dataHoraJ := items[j].Data + " " + items[j].Hora

		tI, errI := time.Parse("2006-01-02 15:04", dataHoraI)
		tJ, errJ := time.Parse("2006-01-02 15:04", dataHoraJ)

		if errI != nil || errJ != nil {
			return false
		}

		return tI.Before(tJ)
	})

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
			1: "",
			2: "CTE",
			3: "CON",
			4: "INF",
			5: "CI",
			6: "ROB",
		}[item.Objeto] // acessa diretamente o valor correspondente à chave item.Objeto

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
	pdf.Cell(0, 10, fmt.Sprintf("Atualizado em: %s", now.Format("2006-01-02 15:04:05")))

	// Send PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=concursos.pdf")

	if err := pdf.Output(w); err != nil {
		log.Printf("Error generating PDF: %v", err)
		http.Error(w, "Erro ao gerar PDF", http.StatusInternalServerError)
	}
}
