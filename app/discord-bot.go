package app

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	expiremap "github.com/nursik/go-expire-map"
	"log"
	"net/mail"
	"strings"
	"time"
)

type DiscordBot struct {
	IsStarted        bool
	Session          *discordgo.Session
	config           BotConfig
	expirationMap    *expiremap.ExpireMap
	userVerification UserVerification
}

func (b *DiscordBot) Start(c Config) {
	b.config = *c.Bot
	b.expirationMap = expiremap.New()
	b.userVerification = &MapUserVerification{
		tokens: expiremap.New(),
	}

	if !b.IsStarted {
		fmt.Println("Starting discord bot.")

		session, err := discordgo.New(c.Bot.Token)

		if err != nil {
			err = fmt.Errorf("an error has occurred while initializing the discord bot: %v", err)
			panic(err)
		}

		session.Identify.Intents = discordgo.IntentsGuildMessages

		b.Session = session

		err = b.Session.Open()

		if err != nil {
			err = fmt.Errorf("an error has occurred while starting the discord bot: %v", err)
			panic(err)
		}

		b.addListener()
		b.sendMessage()
	}
}

func (b *DiscordBot) sendMessage() {
	buildMessage := b.buildMessage()

	message, err := b.Session.ChannelMessageSendComplex(b.config.Channel, &buildMessage)

	if err != nil {
		err = fmt.Errorf("an error has occurred while sending messages in channel: %v", err)
		log.Fatal(err)
		return
	}

	//defer b.Session.ChannelMessageDelete(b.config.Channel, message.ID)
	b.deleteAllMessagesInChannel(message.ID)
}

func (b *DiscordBot) buildMessage() discordgo.MessageSend {
	return discordgo.MessageSend{
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
}

func (b *DiscordBot) addListener() {
	b.Session.AddHandler(b.listenButtons)
}

func (b *DiscordBot) listenButtons(s *discordgo.Session, m *discordgo.InteractionCreate) {
	if m.Type != discordgo.InteractionMessageComponent {
		return
	}

	data := m.MessageComponentData()

	if data.CustomID == "verify" {
		if b.memberHasRole(m) {
			return
		}

		if b.checkMemberExpiration(m) {
			return
		}

		b.openVerificationModal(m)
	} else {
		fmt.Println("TODO: Kick User")
	}
}

func (b *DiscordBot) listenModal(s *discordgo.Session, m *discordgo.InteractionCreate) {
	if m.Type != discordgo.InteractionModalSubmit {
		return
	}

	data := m.ModalSubmitData()

	if data.CustomID == "verification-modal" {

	}
}

func (b *DiscordBot) handleVerificationModal(m *discordgo.InteractionCreate, messageComponent []discordgo.MessageComponent) {
	// Provides user input
	fullname := messageComponent[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	email := messageComponent[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

	// Check email address
	if !strings.HasSuffix(email, ".hwr-berlin.de") && !strings.HasSuffix(email, "@hwr-berlin.de") {
		b.Session.InteractionRespond(m.Interaction, &InvalidEmailAddressResponse)
		return
	}

	_, err := mail.ParseAddress(email)

	if err != nil {
		b.Session.InteractionRespond(m.Interaction, &InvalidEmailAddressResponse)
		return
	}

	token := b.userVerification.GenerateToken(&VerficationRequest{
		guildId:    m.GuildID,
		userId:     m.Member.User.ID,
		expiringAt: time.Now().Add(time.Duration(5) * time.Minute),
	})
	vURLBuilder := new(strings.Builder)

	/*
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
		b.expirationMap.Set(m.Member.User.ID, true, time.Duration(1)*time.Minute)
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
		err = b.Session.InteractionRespond(m.Interaction, &SuccessEmailAddressResponse)
		if err != nil {
			fmt.Println(err)
		}*/

}

func (b *DiscordBot) memberHasRole(ic *discordgo.InteractionCreate) bool {
	m := ic.Member

	if contains(m.Roles, b.config.VerificationRole) {
		err := b.Session.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Du bist bereits verifiziert.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

		if err != nil {
			err = fmt.Errorf("an error has occurred while send reaction: %v", err)
			log.Fatal(err)
		}

		return true
	}

	return false
}

func (b *DiscordBot) checkMemberExpiration(ic *discordgo.InteractionCreate) bool {
	m := ic.Member

	_, ok := b.expirationMap.Get(m.User.ID)

	if ok {
		err := b.Session.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Probiere es bitte später erneut.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

		if err != nil {
			err = fmt.Errorf("an error has occurred while send reaction: %v", err)
			log.Fatal(err)
		}

		return true
	}

	return false
}

func (b *DiscordBot) openVerificationModal(ic *discordgo.InteractionCreate) {
	response := b.buildVerificationModal()

	err := b.Session.InteractionRespond(ic.Interaction, &response)

	if err != nil {
		err = fmt.Errorf("an error has occurred while open verification modal: %v", err)
		log.Fatal(err)
	}
}

func (b *DiscordBot) buildVerificationModal() discordgo.InteractionResponse {
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

	return discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &data,
	}
}

func (b *DiscordBot) deleteAllMessagesInChannel(messageId string) {
	messages, err := b.Session.ChannelMessages(b.config.Channel, 100, messageId, "", "")

	if err != nil {
		err = fmt.Errorf("an error has occurred while getting old messages in channel: %v", err)
		log.Fatal(err)
		return
	}

	messageIDs := make([]string, len(messages))
	for i, message := range messages {
		messageIDs[i] = message.ID
	}

	if err := b.Session.ChannelMessagesBulkDelete(b.config.Channel, messageIDs); err != nil {
		err = fmt.Errorf("an error has occurred while deleting old messages in channel: %v", err)
		log.Fatal(err)
		return
	}
}

func contains[C comparable](s []C, e C) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
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
