package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"
)

type Config struct {
	Sites  []string `json:"sites"`
	Emails []string `json:"emails"`
	Smtp   SmtpConfig `json:"smtp"`
}

type SmtpConfig struct {
	Server string `json:"server"`
	From   string `json:"from"`
	Password string `json:"password"`
}

func main() {
	// Load config from file
	configFile := "config.json"
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	// Monitor each site
	for _, site := range config.Sites {
		// Connect to the site and get the TLS certificate
		conn, err := tls.Dial("tcp", site, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		// Get the TLS certificate
		cert := conn.ConnectionState().PeerCertificates[0]

		// Check if the certificate is expired or near expiration
		expirationDate := cert.NotAfter
		now := time.Now()
		if expirationDate.Before(now) || expirationDate.Sub(now) < 7*24*time.Hour {
			// Send an email with the expiration date
			sendEmail(config.Emails, config.Smtp, site, expirationDate)
		}
	}
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func sendEmail(emails []string, smtpConfig SmtpConfig, site string, expirationDate time.Time) {
	for _, email := range emails {
		// Set up the email message
		subject := "Certificate Expiration Warning"
		body := fmt.Sprintf("The certificate on %s is set to expire on %s.", site, expirationDate.Format(time.RFC822))
		msg := fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body)

		// Set up the SMTP connection
		auth := smtp.PlainAuth("", smtpConfig.From, smtpConfig.Password, smtpConfig.Server)
		err := smtp.SendMail(smtpConfig.Server+":587", auth, smtpConfig.From, []string{email}, []byte(msg))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Email sent to %s successfully\n", email)
	}
}