package pail

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (p *Pail) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
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

	fact, err := getFact(p.db, msg)
	if err != nil {
		return
	}
	if fact != nil {
		p.lastFact = fact
		fact.handle(s, m.ChannelID)
		return
	}

	if mentioned {
		regex := loadRegex(p.db, true)
		for _, rxp := range regex {
			if rxp.Compiled.MatchString(msg) {
				rxp.handle(p, s, m.ChannelID, msg, m.Author.Mention())
				return
			}
		}
	}

	regex := loadRegex(p.db, false)
	for _, rxp := range regex {
		if rxp.Compiled.MatchString(msg) {
			rxp.handle(p, s, m.ChannelID, msg, m.Author.Mention())
			return
		}
	}
}
