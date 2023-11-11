package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type SmtpConfig struct {
	Hostname string
	Port     int
	Username string
	Password string

	// Max Mustermann <max.mustermann@musterfirma.com>
	SenderMail string
	ReplyMail  string
}

func SendMail(config *SmtpConfig, recipEmail string, subject string, text string) error {

	auth := smtp.PlainAuth("", config.Username, config.Password, config.Hostname)
	msg := "From: " + config.ReplyMail + "\nTo: <" + recipEmail + ">\nSubject: " + subject + "\nContent-Type: text/html; charset=\"UTF-8\";\n\n" + text

	conn, err := net.Dial("tcp", fmt.Sprint(config.Hostname, ":", config.Port))
	if err != nil {
		return err
	}
	tlsConn := tls.Client(conn, &tls.Config{ServerName: config.Hostname})

	client, err := smtp.NewClient(tlsConn, config.Hostname)
	if err != nil {
		return err
	}
	client.Hello(config.Hostname)

	err = client.Auth(auth)
	if err != nil {
		return err
	}
	err = client.Mail(config.SenderMail)
	if err != nil {
		return err
	}
	err = client.Rcpt(recipEmail)
	if err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	fmt.Fprintln(writer, msg)
	err = client.Quit()
	if err != nil {
		if !strings.HasPrefix(err.Error(), "250 2.0.0 Ok:") {
			return err
		}
		client.Text.Close()
	}
	return nil
}
