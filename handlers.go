package pail

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	msg := strings.TrimSpace(m.Content)
	parts := strings.Split(msg, " ")

	mentioned := false
	for _, mention := range m.Mentions {
		if mention.ID == s.State.User.ID && fmt.Sprintf("<@!%s>", s.State.User.ID) == parts[0] {
			msg = strings.Join(parts[1:], " ")
			mentioned = true
			break
		}
	}

	fact, err := getFact(msg)
	if err != nil {
		return
	}
	if fact != nil {
		fact.handle(s, m.ChannelID)
		return
	}

	if mentioned {
		regex := loadRegex(true)
		for _, rxp := range regex {
			if rxp.Compiled.MatchString(msg) {
				rxp.handle(s, m.ChannelID, msg, m.Author.Mention())
				return
			}
		}
	}

	regex := loadRegex(false)
	for _, rxp := range regex {
		if rxp.Compiled.MatchString(msg) {
			rxp.handle(s, m.ChannelID, msg, m.Author.Mention())
			return
		}
	}
}
