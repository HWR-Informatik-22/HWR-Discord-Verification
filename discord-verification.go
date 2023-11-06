package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/mail"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	expiremap "github.com/nursik/go-expire-map"
)

func contains[C comparable](s []C, e C) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func listenButtons(s *discordgo.Session, m *discordgo.InteractionCreate) {
	if m.Type != discordgo.InteractionMessageComponent {
		return
	}
	data := m.MessageComponentData()
	switch data.CustomID {
	case "verify":
		if contains(m.Member.Roles, os.Getenv("VERIFICATION_ROLE")) {
			s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Du bist bereits verifiziert.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		_, ok := mailExpirationMap.Get(m.Member.User.ID)
		if ok {
			s.InteractionRespond(m.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Probiere es bitte später erneut.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

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

func deleteAllNewMessages(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.ChannelID != os.Getenv("DISCORD_CHANNEL") {
		return
	}
	s.ChannelMessageDelete(m.ChannelID, m.ID)
}

var InvalidEmailAddressResponse = discordgo.InteractionResponse{
	Type: discordgo.InteractionResponseChannelMessageWithSource,
	Data: &discordgo.InteractionResponseData{
		Flags:   discordgo.MessageFlagsEphemeral,
		Content: "Die angegebene E-Mail Adresse ist nicht gültig.",
	},
}

var SuccessEmailAddressResponse = discordgo.InteractionResponse{
	Type: discordgo.InteractionResponseChannelMessageWithSource,
	Data: &discordgo.InteractionResponseData{
		Flags:   discordgo.MessageFlagsEphemeral,
		Content: "Bitte gucken Sie in Ihr E-Mail Postfach und bestätigen Sie die E-Mail.",
	},
}

type VerificationRequest struct {
	Fullname        string
	DiscordTag      string
	AvatarURL       string
	BannerURL       string
	AccentColor     string
	VerificationURL string
}

func handleVerificationModal(s *discordgo.Session, m *discordgo.InteractionCreate, messageComponent []discordgo.MessageComponent) {
	// Provides user input
	fullname := messageComponent[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	email := messageComponent[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	// Check email address
	if !strings.HasSuffix(email, ".hwr-berlin.de") && !strings.HasSuffix(email, "@hwr-berlin.de") {
		s.InteractionRespond(m.Interaction, &InvalidEmailAddressResponse)
		return
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		s.InteractionRespond(m.Interaction, &InvalidEmailAddressResponse)
		return
	}

	token := userVerification.GenerateToken(&VerficationRequest{
		guildId:    m.GuildID,
		userId:     m.Member.User.ID,
		expiringAt: time.Now().Add(time.Duration(5) * time.Minute),
	})
	vURLBuilder := new(strings.Builder)
	err = verificationURL.Execute(vURLBuilder, token)
	if err != nil {
		panic(err)
	}

	request := &VerificationRequest{Fullname: fullname, DiscordTag: m.Member.User.String(), AvatarURL: m.Member.User.AvatarURL("512"), BannerURL: m.Member.User.BannerURL("512"), VerificationURL: vURLBuilder.String()}

	// Executes templates
	emailSubject := new(strings.Builder)
	emailSubjectTemplate.Execute(emailSubject, request)

	emailBody := new(strings.Builder)
	emailTemplate.Execute(emailBody, request)

	// Sends email to the email address
	err = SendMail(smtpConfig, email, emailSubject.String(), emailBody.String())
	mailExpirationMap.Set(m.Member.User.ID, true, time.Duration(1)*time.Minute)
	fmt.Println("Sending email to", email, "...")
	if err != nil {
		fmt.Println("Failed to send email", err)
		err = s.InteractionRespond(m.Interaction, &InvalidEmailAddressResponse)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	fmt.Println("Email successfully sent.")
	err = s.InteractionRespond(m.Interaction, &SuccessEmailAddressResponse)
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

func deleteAllMessagesInChannel(s *discordgo.Session, channelId string, beforeId string) {
	messages, err := s.ChannelMessages(channelId, 100, beforeId, "", "")
	if err != nil {
		fmt.Println("Fehler beim Abrufen der Nachrichten: ", err)
		return
	}

	messageIDs := make([]string, len(messages))
	for i, message := range messages {
		messageIDs[i] = message.ID
	}

	if err := s.ChannelMessagesBulkDelete(channelId, messageIDs); err != nil {
		fmt.Println("Fehler beim Löschen der Nachrichten: ", err)
		return
	}
}

var smtpConfig *SmtpConfig
var emailTemplate *template.Template
var emailSubjectTemplate *template.Template
var verificationURL *template.Template
var userVerification UserVerification

var mailExpirationMap *expiremap.ExpireMap

func main() {
	godotenv.Load()

	rand.Seed(time.Now().Unix())

	// Loading smtp config
	{
		smtpPort, err := strconv.ParseInt(os.Getenv("SMTP_PORT"), 10, 16)
		if err != nil {
			panic(err)
		}

		smtpConfig = &SmtpConfig{
			Hostname:   os.Getenv("SMTP_HOST"),
			Port:       int(smtpPort),
			Username:   os.Getenv("SMTP_AUTH_USERNAME"),
			Password:   os.Getenv("SMTP_AUTH_PASSWORD"),
			SenderMail: os.Getenv("SMTP_SENDER_EMAIL"),
			ReplyMail:  os.Getenv("SMTP_REPLY_EMAIL"),
		}
	}
	// Loading email template
	{
		emailTemplate = template.Must(template.ParseFiles(os.Getenv("EMAIL_TEMPLATE")))

		emailSubjectTemplate = template.New("subject")
		emailSubjectTemplate = template.Must(emailSubjectTemplate.Parse(os.Getenv("EMAIL_SUBJECT")))

		verificationURL = template.New("verificationURL")
		verificationURL = template.Must(verificationURL.Parse(os.Getenv("VERIFICATION_URL")))
	}

	userVerification = NewMapUserVerification()
	mailExpirationMap = expiremap.New()

	session, err := discordgo.New(os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}
	session.Identify.Intents = discordgo.IntentsGuildMessages

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		q := r.URL.Query()
		token := q.Get("token")
		if len(token) != 0 {
			request, err := userVerification.VerifyToken(token)
			if err != nil {
				fmt.Println(err)
				return
			}
			session.GuildMemberRoleAdd(request.guildId, request.userId, os.Getenv("VERIFICATION_ROLE"))
		}
		if r.URL.Path == "/" {
			w.Header().Add("Location", "https://discord.gg/"+os.Getenv("DISCORD_INVITE"))
			w.WriteHeader(301)
			fmt.Println("Redirecting", token, "to "+"https://discord.gg/"+os.Getenv("DISCORD_INVITE"))
		}
	})

	go func() {
		err := http.ListenAndServe("0.0.0.0:8080", mux)
		if err != nil {
			panic(err)
		}
	}()

	session.AddHandler(listenButtons)
	session.AddHandler(listenSubmits)

	// enabled, err := strconv.ParseBool(os.Getenv("SILENT_CHANNEL"))
	// if err != nil {
	// 	panic(err)
	// }
	// if enabled {
	// 	session.AddHandler(deleteAllNewMessages)
	// }

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
	defer session.ChannelMessageDelete(channelId, message.ID)
	deleteAllMessagesInChannel(session, channelId, message.ID)

	fmt.Println("Waiting on interruption...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	err = session.Close()
	if err != nil {
		panic(err)
	}

	userVerification.Close()
	mailExpirationMap.Close()
}
