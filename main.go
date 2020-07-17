package main

import (
	"log"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var commandPrefix = "!"
var latestMessages = map[string]*discordgo.Message{}
var latestMessagesLock = sync.RWMutex{}
var messageIDRegex = regexp.MustCompile(`^[\d]+$`)
var channelIDRegex = regexp.MustCompile(`[\d]+`)

var session *discordgo.Session

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Failed to load .env: %v\n", err)
	}

	token := os.Getenv("ONBROID_TOKEN")
	if token == "" {
		log.Fatalf("ONBROID_TOKEN is missing!")
	}
	if prefix := os.Getenv("ONBROID_COMMAND_PREFIX"); prefix != "" {
		commandPrefix = prefix
	}

	latestMessages = make(map[string]*discordgo.Message)

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Failed to initialize Discord session: %v\n", err)
	}
	session = s
	session.AddHandler(onMessage)

	if session.Open() != nil {
		log.Fatalf("Failed to open Discord session: %v\n", err)
	}

	<-make(chan struct{})
}

// onMessage invokes command handler if message starts with commandPrefix.
// else, stores latest command to latestMessages map.
func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if strings.HasPrefix(m.Content, commandPrefix) {
		m := m.Message
		// invoke command handler
		splitted := strings.SplitN(strings.TrimPrefix(m.Content, commandPrefix), " ", 2)
		// splitted[0] -> command
		// splitted[1] -> rest of message
		args := strings.TrimSpace(splitted[1])

		// it's entirely small code base, so don't need command dispatcher I think...
		switch splitted[0] {
		case "move", "mv":
			cmdMove(m, args)
		case "copy", "cp":
			cmdCopy(m, args)
		}
	} else {
		// store latest message
		latestMessagesLock.Lock()
		defer latestMessagesLock.Unlock()

		latestMessages[m.ChannelID] = m.Message
	}
}

// cmdMove moves message to another channel.
// if no message ID is given, move latest message in channel.
// !move [message ID, or URL] <channel>
func cmdMove(m *discordgo.Message, args string) {
	copied := doCopy(m, args)
	if copied != nil {
		session.ChannelMessageDelete(copied.ChannelID, copied.ID)
	}
}

// cmdCopy copies message to another channel.

func cmdCopy(m *discordgo.Message, args string) {
	doCopy(m, args)
}

// doCopy copies message, returns copied message struct.
func doCopy(m *discordgo.Message, args string) (copiedMessage *discordgo.Message) {
	sp := strings.Split(args, " ")
	var lmsg *discordgo.Message
	var targetChannel string

	switch splen := len(sp); {
	case splen == 1:
		// !move <channel>
		latestMessagesLock.RLock()
		defer latestMessagesLock.RUnlock()
		lmsg = latestMessages[m.ChannelID]

		targetChannel = sp[0]
	case splen >= 2:
		// !move [messageID] <channel>
		msg, err := session.ChannelMessage(m.ChannelID, getMessageID(sp[0]))
		if err != nil {
			log.Printf("Message not found for ID %s\n", sp[0])
			return
		}
		lmsg = msg

		targetChannel = sp[1]

	default:
		log.Printf("Invalid args: %v\n", args)
		return
	}

	if lmsg == nil {
		log.Printf("No message is cached for channel %s\n", m.ChannelID)
		return
	}
	// I'm not sure why this won't work.
	// if len(m.MentionChannels) == 0 {
	// 	log.Printf("No mentioned channel.")
	// 	return
	// }
	if targetChannel == "" {
		log.Printf("No channelID.")
		return
	}
	targetChannel = getChannelID(targetChannel)

	clone, err := getCloneEmbed(lmsg)
	if err != nil {
		log.Printf("Failed to clone message for %s (%s): %v\n", lmsg.ID, lmsg.Content, err)
		return
	}
	// TODO: support multiple embed!
	if _, err := session.ChannelMessageSendEmbed(targetChannel, clone); err != nil {
		log.Printf("Failed to send clone embed for message %s (%s): %v\n", lmsg.ID, lmsg.Content, err)
	}

	return lmsg
}

func getCloneEmbed(m *discordgo.Message) (*discordgo.MessageEmbed, error) {
	// TODO: support multiple embed!
	if len(m.Embeds) > 0 {
		return m.Embeds[0], nil
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    m.Author.Username,
			IconURL: m.Author.AvatarURL(""),
		},
		Description: m.Content,
		Timestamp:   string(m.Timestamp),
		Color:       rand.Intn(0x1000000),
	}

	// NEW! handle media.
	// TODO: is this work?
	if len(m.Attachments) > 0 {
		mediaURLs := []string{}
		for _, a := range m.Attachments {
			mediaURLs = append(mediaURLs, a.URL)
		}

		embed.Fields = []*discordgo.MessageEmbedField{
			{
				Name:   "Attachments",
				Value:  strings.Join(mediaURLs, "\n"),
				Inline: false,
			},
		}
	}

	return embed, nil
}

// getMessageID returns messageID from URL, or messageID.
func getMessageID(raw string) string {
	if messageIDRegex.MatchString(raw) {
		return raw
	}
	sp := strings.Split(raw, "/")
	maybeID := sp[len(sp)-1]
	if messageIDRegex.MatchString(maybeID) {
		return maybeID
	}

	return ""
}

func getChannelID(raw string) string {
	return channelIDRegex.FindString(raw)
}
