package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type SmtpConfig struct {
	hostname string
	port     int
	username string
	password string

	// Max Mustermann <max.mustermann@musterfirma.com>
	senderMail string
	replyMail  string
}

func SendMail(config *SmtpConfig, recipEmail string, subject string, text string) error {

	auth := smtp.PlainAuth("", config.username, config.password, config.hostname)
	msg := "From: " + config.replyMail + "\nTo: <" + recipEmail + ">\nSubject: " + subject + "\n\n" + text

	conn, err := net.Dial("tcp", fmt.Sprint(config.hostname, ":", config.port))
	if err != nil {
		return err
	}
	tlsConn := tls.Client(conn, &tls.Config{ServerName: config.hostname})

	client, err := smtp.NewClient(tlsConn, config.hostname)
	if err != nil {
		return err
	}
	client.Hello("stud.hwr-berlin.de")

	err = client.Auth(auth)
	if err != nil {
		return err
	}
	err = client.Mail(config.senderMail)
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
