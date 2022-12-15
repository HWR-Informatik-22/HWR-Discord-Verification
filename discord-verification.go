package main

import (
	"fmt"
	"net/mail"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func listenButtons(s *discordgo.Session, m *discordgo.InteractionCreate) {
	if m.Type != discordgo.InteractionMessageComponent {
		return
	}
	data := m.MessageComponentData()
	switch data.CustomID {
	case "verify":
		OpenVerificationModal(s, m)
	case "leave-discord":
		fmt.Println("Kick discord")
	}
}

func listenSubmits(s *discordgo.Session, m *discordgo.InteractionCreate) {
	if m.Type != discordgo.InteractionModalSubmit {
		return
	}
	data := m.ModalSubmitData()
	switch data.CustomID {
	case "verification-modal":
		handleVerificationModal(s, m, data.Components)
	}
}

var EmailMatch *regexp.Regexp = regexp.MustCompile("(?:[a-z0-9!#$%&\\'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&\\'*+/=?^_`{|}~-]+)*|\"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\\\[\x01-\x09\x0b\x0c\x0e-\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\\])")

var InvalidEmailAddressResponse = discordgo.InteractionResponse{
	Type: discordgo.InteractionResponseChannelMessageWithSource,
	Data: &discordgo.InteractionResponseData{
		Flags:   discordgo.MessageFlagsEphemeral,
		Content: "Die angegebene E-Mail Adresse ist nicht gültig.",
	},
}

func handleVerificationModal(s *discordgo.Session, m *discordgo.InteractionCreate, messageComponent []discordgo.MessageComponent) {
	fullname := messageComponent[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	email := messageComponent[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	if !strings.HasSuffix(email, ".hwr-berlin.de") {
		s.InteractionRespond(m.Interaction, &InvalidEmailAddressResponse)
		return
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		s.InteractionRespond(m.Interaction, &InvalidEmailAddressResponse)
		return
	}

	err = SendMail(smtpConfig, email, "Hallo "+fullname+"! Jemand möchte sich mit einem Discord Account auf einem HWR-Discord verifizieren",
		"Hallo "+fullname+",\nDer Discord Account \""+m.Member.User.String()+"\" möchte sich mit Ihrer E-Mail verknüpfen. Wenn Sie das sind klicken Sie bitte auf diesen Link: http://hhhh")

	if err != nil {
		fmt.Println(err)
		err = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Es ist ein Fehler beim E-Mail senden aufgetreten. Bitte überprüfen Sie Ihre E-Mail Adresse.",
			},
		})
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	err = s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:   discordgo.MessageFlagsEphemeral,
			Content: "Bitte gucken Sie in Ihr E-Mail Postfach und bestätigen Sie die E-Mail.",
		},
	})
	if err != nil {
		fmt.Println(err)
	}

}

func OpenVerificationModal(s *discordgo.Session, m *discordgo.InteractionCreate) {
	data := discordgo.InteractionResponseData{
		Title:    "Test",
		CustomID: "verification-modal",
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "name",
						Label:       "Vor- und Nachname",
						Placeholder: "Max Mustermann",
						Required:    true,
						MinLength:   7, // 3 + space + 3
						Style:       discordgo.TextInputShort,
					},
				},
			},
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.TextInput{
						CustomID:    "email",
						Label:       "E-Mail Adresse",
						Placeholder: "s_mustermann22@stud.hwr-berlin.de",
						Required:    true,
						MinLength:   26,
						Style:       discordgo.TextInputShort,
					},
				},
			},
			// TODO: Implement role selection but currently only text field are supported: https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-modal
			// discordgo.ActionsRow{
			// 	Components: []discordgo.MessageComponent{
			// 		discordgo.SelectMenu{
			// 			MenuType:    discordgo.UserSelectMenu,
			// 			Placeholder: "teass",
			// 			CustomID:    "course",
			// 			MaxValues:   1,
			// 			MinValues:   &a, // shit of pointers
			// 			Placeholder: "a",
			// 			MinValues:   nil,
			// 			Options: []discordgo.SelectMenuOption{
			// 				{
			// 					Label:       "Kurs A",
			// 					Emoji:       discordgo.ComponentEmoji{Name: "🅰️"},
			// 					Value:       "a",
			// 					Description: "Test",
			// 				},
			// 				{
			// 					Label:       "Kurs B",
			// 					Emoji:       discordgo.ComponentEmoji{Name: "🇧"},
			// 					Value:       "b",
			// 					Description: "test",
			// 				},
			// 			},
			// 		},
			// 	},
			// },
		},
	}

	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &data,
	}

	err := s.InteractionRespond(m.Interaction, &response)

	if err != nil {
		panic(err)
	}
}

var smtpConfig *SmtpConfig

func main() {

	godotenv.Load(".env")

	{
		smtpPort, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 16)
		if err != nil {
			panic(err)
		}

		smtpConfig = &SmtpConfig{
			hostname:   os.Getenv("SMTP_HOST"),
			port:       int(smtpPort),
			username:   os.Getenv("SMTP_AUTH_USERNAME"),
			password:   os.Getenv("SMTP_AUTH_PASSWORD"),
			senderMail: os.Getenv("SMTP_SENDER_EMAIL"),
			replyMail:  os.Getenv("SMTP_REPLY_EMAIL"),
		}
	}

	session, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}
	session.Identify.Intents = discordgo.IntentsAll

	session.AddHandler(listenButtons)
	session.AddHandler(listenSubmits)

	err = session.Open()
	if err != nil {
		panic(err)
	}

	channelId := os.Getenv("DISCORD_CHANNEL")

	messageSend := discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Type:        discordgo.EmbedTypeArticle,
				Title:       "Kontoverifizierung",
				Description: "Guten Tag,\nSie befinden sich auf einem Discord Server für Studierende der Hochschule für Wirtschaft und Recht. Damit wir überprüfen können ob Sie Studierender der Hochschule sind, benötigen wir Ihren Namen und Ihre E-Mail Adresse.",
				Fields:      []*discordgo.MessageEmbedField{},
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Verifizieren",
						Style:    discordgo.PrimaryButton,
						CustomID: "verify",
					},
					discordgo.Button{
						Label:    "Verlassen",
						Style:    discordgo.DangerButton,
						CustomID: "leave-discord",
					},
				},
			},
		},
	}

	message, err := session.ChannelMessageSendComplex(channelId, &messageSend)
	if err != nil {
		panic(err)
	}

	fmt.Println("Message: " + message.ID)

	fmt.Println("Waiting on interruption...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	err = session.Close()
	if err != nil {
		panic(err)
	}
}
