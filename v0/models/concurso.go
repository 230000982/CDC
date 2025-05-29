package models

import (
	"database/sql"
	"time"
)

// Concurso represents a concurso record
type Concurso struct {
	ID             int
	Preco          float64
	Referencia     string
	Entidade       string
	DiaErro        NullString
	HoraErro       NullString
	DiaProposta    NullString
	HoraProposta   NullString
	ReferenciaBC   string
	Preliminar     bool
	DiaAudiencia   NullString
	HoraAudiencia  NullString
	Final          bool
	Recurso        bool
	Impugnacao     bool
	TipoID         int
	PlataformaID   int
	EstadoID       int
	TipoDesc       string
	PlataformaDesc string
	Link           string        // New field
	Adjudicatario  string        // New field
	ResultadoID    int           // New field
	ResultadoDesc  string        // New field for display
}

// ConcursoItem represents a concurso item for the ordered view
type ConcursoItem struct {
	Referencia    string
	Entidade      string
	Objeto        int
	Data          string
	Hora          string
	Tipo          string
	DiasRestantes int
	Link          string        // New field
	Adjudicatario string        // New field
	ResultadoID   int           // New field
}

// GetConcursoByID retrieves a concurso by ID
func GetConcursoByID(db *sql.DB, id string) (*Concurso, error) {
	var concurso Concurso

	err := db.QueryRow(`
        SELECT c.id_concurso, c.preco, c.referencia, c.entidade, c.dia_erro, c.hora_erro, 
               c.dia_proposta, c.hora_proposta, c.referencia_bc, c.preliminar, 
               c.dia_audiencia, c.hora_audiencia, c.final, c.recurso, c.impugnacao, 
               c.tipo_id, c.plataforma_id, c.estado_id, c.link, c.adjudicatario, c.resultado_id
        FROM concurso c
        WHERE c.id_concurso = ?
    `, id).Scan(
		&concurso.ID, &concurso.Preco, &concurso.Referencia, &concurso.Entidade,
		&concurso.DiaErro.NullString, &concurso.HoraErro.NullString,
		&concurso.DiaProposta.NullString, &concurso.HoraProposta.NullString,
		&concurso.ReferenciaBC, &concurso.Preliminar,
		&concurso.DiaAudiencia.NullString, &concurso.HoraAudiencia.NullString,
		&concurso.Final, &concurso.Recurso, &concurso.Impugnacao,
		&concurso.TipoID, &concurso.PlataformaID, &concurso.EstadoID,
		&concurso.Link, &concurso.Adjudicatario, &concurso.ResultadoID,
	)

	if err != nil {
		return nil, err
	}

	return &concurso, nil
}

// GetConcursos retrieves all concursos with optional filtering
func GetConcursos(db *sql.DB, entidade string) ([]Concurso, error) {
	// Construct the base query
	query := `
        SELECT c.id_concurso, c.preco, c.referencia, c.entidade, c.dia_erro, c.hora_erro, c.dia_proposta, c.hora_proposta, 
               c.referencia_bc, c.preliminar, c.dia_audiencia, c.hora_audiencia, c.final, c.recurso, c.impugnacao, 
               t.descricao AS tipo, p.descricao AS plataforma, c.estado_id AS estado, c.tipo_id, c.plataforma_id,
               c.link, c.adjudicatario, c.resultado_id, r.descricao AS resultado
        FROM concurso c
        JOIN tipo t ON c.tipo_id = t.id_tipo
        JOIN plataforma p ON c.plataforma_id = p.id_platforma
        LEFT JOIN resultado r ON c.resultado_id = r.id_resultado
    `

	// Add filter if there's a search by entidade
	var args []interface{}
	if entidade != "" {
		query += " WHERE c.entidade LIKE ?"
		args = append(args, "%"+entidade+"%")
	}

	// Add ordering
	query += " ORDER BY c.dia_proposta DESC, c.hora_proposta DESC"

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Prepare data
	var concursos []Concurso

	// Scan rows
	for rows.Next() {
		var c Concurso
		var preco float64
		err := rows.Scan(
			&c.ID, &preco, &c.Referencia, &c.Entidade,
			&c.DiaErro.NullString, &c.HoraErro.NullString,
			&c.DiaProposta.NullString, &c.HoraProposta.NullString,
			&c.ReferenciaBC, &c.Preliminar,
			&c.DiaAudiencia.NullString, &c.HoraAudiencia.NullString,
			&c.Final, &c.Recurso, &c.Impugnacao,
			&c.TipoDesc, &c.PlataformaDesc, &c.EstadoID, &c.TipoID, &c.PlataformaID,
			&c.Link, &c.Adjudicatario, &c.ResultadoID, &c.ResultadoDesc,
		)
		if err != nil {
			return nil, err
		}

		c.Preco = preco
		concursos = append(concursos, c)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return concursos, nil
}

// GetFutureConcursos retrieves concursos with future dates
func GetFutureConcursos(db *sql.DB) ([]ConcursoItem, error) {
	// Get current date and time
	now := time.Now()
	currentDate := now.Format("2006-01-02")
	currentTime := now.Format("15:04:05")

	// Query concursos with estado_id = 1
	rows, err := db.Query(`
        SELECT c.id_concurso, c.entidade, c.estado_id,
               c.dia_erro, c.hora_erro, 
               c.dia_proposta, c.hora_proposta,
               c.dia_audiencia, c.hora_audiencia,
               c.tipo_id, c.referencia, c.link, c.adjudicatario, c.resultado_id
        FROM concurso c
        WHERE c.estado_id = 2
        ORDER BY 
            COALESCE(c.dia_proposta, c.dia_erro, c.dia_audiencia) ASC,
            COALESCE(c.hora_proposta, c.hora_erro, c.hora_audiencia) ASC
    `)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ConcursoItem

	// Function to calculate remaining days
	calcularDiasRestantes := func(data string) int {
		concursoDate, err := time.Parse("2006-01-02", data)
		if err != nil {
			return -1
		}

		diff := concursoDate.Sub(now)
		return int(diff.Hours() / 24)
	}

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
		var entidade, referencia, link, adjudicatario string
		var resultadoID int
		var diaErro, horaErro sql.NullString
		var diaProposta, horaProposta sql.NullString
		var diaAudiencia, horaAudiencia sql.NullString

		err := rows.Scan(
			&id, &entidade, &estado,
			&diaErro, &horaErro,
			&diaProposta, &horaProposta,
			&diaAudiencia, &horaAudiencia,
			&objeto, &referencia, &link, &adjudicatario, &resultadoID,
		)
		if err != nil {
			return nil, err
		}

		// Add each valid date as a separate item, only if it's in the future
		if diaProposta.Valid && horaProposta.Valid && isFutureDate(diaProposta.String, horaProposta.String) {
			dias := calcularDiasRestantes(diaProposta.String)
			items = append(items, ConcursoItem{
				Referencia:    referencia,
				Entidade:      entidade,
				Objeto:        objeto,
				Data:          diaProposta.String,
				Hora:          horaProposta.String,
				Tipo:          "Proposta",
				DiasRestantes: dias,
				Link:          link,
				Adjudicatario: adjudicatario,
				ResultadoID:   resultadoID,
			})
		}

		if diaErro.Valid && horaErro.Valid && isFutureDate(diaErro.String, horaErro.String) {
			dias := calcularDiasRestantes(diaErro.String)
			items = append(items, ConcursoItem{
				Referencia:    referencia,
				Entidade:      entidade,
				Objeto:        objeto,
				Data:          diaErro.String,
				Hora:          horaErro.String,
				Tipo:          "Erro",
				DiasRestantes: dias,
				Link:          link,
				Adjudicatario: adjudicatario,
				ResultadoID:   resultadoID,
			})
		}

		if diaAudiencia.Valid && horaAudiencia.Valid && isFutureDate(diaAudiencia.String, horaAudiencia.String) {
			dias := calcularDiasRestantes(diaAudiencia.String)
			items = append(items, ConcursoItem{
				Referencia:    referencia,
				Entidade:      entidade,
				Objeto:        objeto,
				Data:          diaAudiencia.String,
				Hora:          horaAudiencia.String,
				Tipo:          "Audiencia",
				DiasRestantes: dias,
				Link:          link,
				Adjudicatario: adjudicatario,
				ResultadoID:   resultadoID,
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// CreateConcurso creates a new concurso
func CreateConcurso(db *sql.DB, c *Concurso) error {
	_, err := db.Exec(`
        INSERT INTO concurso (
            referencia, entidade, dia_erro, hora_erro, 
            dia_proposta, hora_proposta, preco, tipo_id, 
            plataforma_id, referencia_bc, preliminar, dia_audiencia, 
            hora_audiencia, final, recurso, impugnacao, estado_id,
            link, adjudicatario, resultado_id
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `,
		c.Referencia, c.Entidade, c.DiaErro.NullString, c.HoraErro.NullString,
		c.DiaProposta.NullString, c.HoraProposta.NullString, c.Preco, c.TipoID,
		c.PlataformaID, c.ReferenciaBC, c.Preliminar, c.DiaAudiencia.NullString,
		c.HoraAudiencia.NullString, c.Final, c.Recurso, c.Impugnacao, c.EstadoID,
		c.Link, c.Adjudicatario, c.ResultadoID)

	return err
}

// UpdateConcurso updates an existing concurso
func UpdateConcurso(db *sql.DB, id string, c *Concurso) error {
	_, err := db.Exec(`
        UPDATE concurso
        SET preco = ?, referencia = ?, entidade = ?, referencia_bc = ?,
            dia_erro = ?, hora_erro = ?, dia_proposta = ?, hora_proposta = ?,
            dia_audiencia = ?, hora_audiencia = ?, preliminar = ?, final = ?,
            recurso = ?, impugnacao = ?, tipo_id = ?, plataforma_id = ?, estado_id = ?,
            link = ?, adjudicatario = ?, resultado_id = ?
        WHERE id_concurso = ?
    `,
		c.Preco, c.Referencia, c.Entidade, c.ReferenciaBC,
		c.DiaErro.NullString, c.HoraErro.NullString,
		c.DiaProposta.NullString, c.HoraProposta.NullString,
		c.DiaAudiencia.NullString, c.HoraAudiencia.NullString,
		c.Preliminar, c.Final, c.Recurso, c.Impugnacao,
		c.TipoID, c.PlataformaID, c.EstadoID,
		c.Link, c.Adjudicatario, c.ResultadoID, id)

	return err
}

// DeleteConcurso deletes a concurso by ID
func DeleteConcurso(db *sql.DB, id string) error {
	_, err := db.Exec("DELETE FROM concurso WHERE id_concurso = ?", id)
	return err
}

// GetAllEmails retrieves all emails from the user table
func GetAllEmails(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT email FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return emails, nil
}
