package app

import (
	"fmt"
	"os"
	"strconv"
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

type BotConfig struct {
	Token            string
	Channel          string
	VerificationRole string
}

type Config struct {
	Smtp *SmtpConfig
	Bot  *BotConfig
}

func (c *Config) Load() {
	c.loadSmtp()
	c.loadBot()
}

func (c *Config) loadSmtp() {
	smtpPort, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 16)

	if err != nil {
		err = fmt.Errorf("an error has occurred while reading the smtp configuration: %v", err)
		panic(err)
	}

	c.Smtp = &SmtpConfig{
		Hostname:   os.Getenv("SMTP_HOST"),
		Port:       int(smtpPort),
		Username:   os.Getenv("SMTP_AUTH_USERNAME"),
		Password:   os.Getenv("SMTP_AUTH_PASSWORD"),
		SenderMail: os.Getenv("SMTP_SENDER_EMAIL"),
		ReplyMail:  os.Getenv("SMTP_REPLY_EMAIL"),
	}
}

func (c *Config) loadBot() {
	c.Bot = &BotConfig{
		Token:            os.Getenv("DISCORD_TOKEN"),
		Channel:          os.Getenv("DISCORD_CHANNEL"),
		VerificationRole: os.Getenv("VERIFICATION_ROLE"),
	}
}
