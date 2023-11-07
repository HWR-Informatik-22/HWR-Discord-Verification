package app

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
)

type DiscordBot struct {
	IsStarted bool
	Session   *discordgo.Session
	config    BotConfig
}

func (b *DiscordBot) Start(c Config) {
	b.config = *c.Bot

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

	} else {

	}

	/*

		switch data.CustomID {
		case "verify":
			if contains(m.Member.Roles, b.config.VerificationRole) {
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
		}*/
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
