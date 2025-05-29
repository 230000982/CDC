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

// SendMailToSpecificRecipient sends an email to a specific recipient
func (s *EmailService) SendMailToSpecificRecipient(messageBody string, recipient string) error {
	// Authentication
	auth := smtp.PlainAuth("", s.config.From, s.config.Password, s.config.SMTPHost)

	// Send email
	err := smtp.SendMail(
		s.config.SMTPHost+":"+s.config.SMTPPort,
		auth,
		s.config.From,
		[]string{recipient},
		[]byte(messageBody),
	)
	if err != nil {
		log.Printf("Error sending email to %s: %v", recipient, err)
		return err
	}

	log.Printf("Email sent successfully to %s!", recipient)
	return nil
}

// SendUpdateEmail sends an email with concurso update information
func (s *EmailService) SendUpdateEmail(
	referencia, entidade, tipo, estado, diaErro, horaErro, diaProposta, horaProposta, diaAudiencia, horaAudiencia string,
	preliminar, final, recurso, impugnacao bool,
	resultado, link string,
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

	// Add new fields
	if resultado != "" && resultado != "1" {
		body.WriteString(fmt.Sprintf("Resultado: %s\n", resultado))
	}
	if link != "" {
		body.WriteString(fmt.Sprintf("Link: %s\n", link))
	}

	// Format complete message with subject and body
	fullMessage := fmt.Sprintf("Subject: %s\n\n%s", subject, body.String())

	// Use the existing SendMail function
	return s.SendMail(fullMessage, db)
}

// SendAdjudicatarioEmail sends an email to the adjudicatario
func (s *EmailService) SendAdjudicatarioEmail(
	adjudicatario, referencia, entidade, tipo, estado, diaErro, horaErro, diaProposta, horaProposta, diaAudiencia, horaAudiencia string,
	preliminar, final, recurso, impugnacao bool,
	resultado, link string,
) error {
	// Format email subject
	subject := fmt.Sprintf("Foi designado como adjudicatário: %s - %s - %s", referencia, entidade, tipo)

	// Build email body
	var body strings.Builder

	body.WriteString(fmt.Sprintf("Olá,\n\nfoi designado como adjudicatário para o seguinte concurso:\n\n"))
	body.WriteString(fmt.Sprintf("Referência: %s\n", referencia))
	body.WriteString(fmt.Sprintf("Entidade: %s\n", entidade))
	body.WriteString(fmt.Sprintf("Tipo: %s\n", tipo))
	body.WriteString(fmt.Sprintf("Estado: %s\n\n", estado))

	// Add dates/times if they exist
	if diaErro != "" && horaErro != "" {
		body.WriteString(fmt.Sprintf("Esclarecimentos/Erro: %s %s\n", diaErro, horaErro))
	}

	if diaProposta != "" && horaProposta != "" {
		body.WriteString(fmt.Sprintf("Proposta: %s %s\n", diaProposta, horaProposta))
	}

	if diaAudiencia != "" && horaAudiencia != "" {
		body.WriteString(fmt.Sprintf("Audiência: %s %s\n", diaAudiencia, horaAudiencia))
	}

	// Add additional information
	body.WriteString("\nDetalhes adicionais:\n")
	if preliminar {
		body.WriteString("- Preliminar: Sim\n")
	}
	if final {
		body.WriteString("- Final: Sim\n")
	}
	if recurso {
		body.WriteString("- Recurso: Sim\n")
	}
	if impugnacao {
		body.WriteString("- Impugnação: Sim\n")
	}

	// Add new fields
	if resultado != "" && resultado != "1" {
		body.WriteString(fmt.Sprintf("- Resultado: %s\n", resultado))
	}
	if link != "" {
		body.WriteString(fmt.Sprintf("\nLink para o concurso: %s\n", link))
	}

	body.WriteString("\n\nEste é um email automático. Por favor, não responda a este email.\n")

	// Format complete message with subject and body
	fullMessage := fmt.Sprintf("Subject: %s\n\n%s", subject, body.String())

	// Send email to adjudicatario
	return s.SendMailToSpecificRecipient(fullMessage, adjudicatario)
}
