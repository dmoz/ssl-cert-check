package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"time"
)

type Config struct {
	Sites                      []string   `json:"sites"`
	Emails                     []string   `json:"emails"`
	Smtp                       SmtpConfig `json:"smtp"`
	ExpirationWarningThreshold Duration   `json:"expirationWarningThreshold"`
	Timeout                    Duration   `json:"timeout"`
}

type SmtpConfig struct {
	Server   string `json:"server"`
	From     string `json:"from"`
	Password string `json:"password"`
}

type Duration time.Duration

// UnmarshalJSON implements the json.Unmarshaler interface.
// It parses a duration from a string, e.g. "1h" or "2h45m".
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

// loadConfig reads a configuration file in JSON format from the specified filename,
// unmarshals its contents into a Config struct, and returns a pointer to the Config.
// It returns an error if the file cannot be read or if the JSON cannot be unmarshaled.
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

// sendEmail sends a warning email to the specified email addresses
// when the certificate for the specified site is about to expire.
// The email is sent using the specified SMTP config.
func sendEmail(emails []string, smtpConfig SmtpConfig, site string, expirationDate time.Time) {
	for _, email := range emails {
		// Set up the email message
		subject := fmt.Sprintf("%s Certificate Expiration Warning", site)
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

// main runs the certificate expiration monitor. It loads the configuration from
// config.json, monitors each site specified in the configuration, and sends a
// warning email if the certificate on the site is about to expire.
func main() {
	// Set up logging
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer file.Close()

	log.SetOutput(file)

	// Load config from file
	configFile := "config.json"
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	// Monitor each site
	for _, site := range config.Sites {
		// Split the site into host and port
		host, port, err := net.SplitHostPort(site)
		if err != nil {
			// If no port is specified, use the default port 443
			host = site
			port = "443"
		}

		// Connect to the site and get the TLS certificate with a timeout
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: time.Duration(config.Timeout)}, "tcp", net.JoinHostPort(host, port), &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			log.Printf("Error connecting to %s: %v", site, err)
			continue
		}

		defer conn.Close()

		// Set a deadline for the connection
		conn.SetDeadline(time.Now().Add(time.Duration(config.Timeout)))

		// Get the TLS certificate
		cert := conn.ConnectionState().PeerCertificates[0]

		// Check if the certificate is expired or near expiration
		expirationDate := cert.NotAfter
		now := time.Now()
		if expirationDate.Before(now) || expirationDate.Sub(now) < time.Duration(config.ExpirationWarningThreshold) {
			// Send an email with the expiration date
			sendEmail(config.Emails, config.Smtp, site, expirationDate)
		}
		log.Printf("Certificate on %s is valid until %s\n", site, expirationDate.Format(time.RFC822))
	}
}
