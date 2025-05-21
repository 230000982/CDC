package services

import (
	"database/sql"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"v0/config"
	"v0/database"
)

// EmailService handles email operations
type EmailService struct {
	config config.EmailConfig
}

// NewEmailService creates a new EmailService
func NewEmailService(cfg config.EmailConfig) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

// SendMail sends an email to all recipients
func (s *EmailService) SendMail(messageBody string, db *sql.DB) error {
	// Get receivers from DB
	to, err := database.GetEmailsFromDB(db)
	if err != nil {
		log.Printf("Error getting emails from database: %v", err)
		return err
	}

	if len(to) == 0 {
		log.Println("No recipients found in database")
		return fmt.Errorf("no recipients found")
	}

	// Authentication
	auth := smtp.PlainAuth("", s.config.From, s.config.Password, s.config.SMTPHost)

	// Send email
	err = smtp.SendMail(
		s.config.SMTPHost+":"+s.config.SMTPPort,
		auth,
		s.config.From,
		to,
		[]byte(messageBody),
	)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}

	log.Println("Email sent successfully to all recipients!")
	return nil
}

// SendUpdateEmail sends an email with concurso update information
func (s *EmailService) SendUpdateEmail(
	referencia, entidade, tipo, estado, diaErro, horaErro, diaProposta, horaProposta, diaAudiencia, horaAudiencia string,
	preliminar, final, recurso, impugnacao bool,
	db *sql.DB,
) error {
	// Format email subject
	subject := fmt.Sprintf("%s - %s - %s", referencia, entidade, tipo)

	// Build email body
	var body strings.Builder

	// Add dates/times if they exist
	if diaErro != "" && horaErro != "" {
		body.WriteString(fmt.Sprintf("%s %s - Esclarecimentos/Erro\n", diaErro, horaErro))
	}

	if diaProposta != "" && horaProposta != "" {
		body.WriteString(fmt.Sprintf("%s %s - Proposta\n", diaProposta, horaProposta))
	}

	if diaAudiencia != "" && horaAudiencia != "" {
		body.WriteString(fmt.Sprintf("%s %s - Audiência\n", diaAudiencia, horaAudiencia))
	}

	// Add additional information
	if preliminar {
		body.WriteString("\nPreliminar: Sim\n")
	}
	if final {
		body.WriteString("Final: Sim\n")
	}
	if recurso {
		body.WriteString("Recurso: Sim\n")
	}
	if impugnacao {
		body.WriteString("Impugnação: Sim\n")
	}

	// Format complete message with subject and body
	fullMessage := fmt.Sprintf("Subject: %s\n\n%s", subject, body.String())

	// Use the existing SendMail function
	return s.SendMail(fullMessage, db)
}
