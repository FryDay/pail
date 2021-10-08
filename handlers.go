package pail

import (
	"fmt"
	"log"
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

	fact, err := getFact(p.db, msg, m.Author.Mention())
	if err != nil {
		return
	}
	if fact != nil {
		p.lastFact = fact
		reply, err := fact.handle()
		if err != nil {
			log.Println(err)
			return
		}
		s.ChannelMessageSend(m.ChannelID, reply)
		p.randomReset <- true
		return
	}

	if mentioned {
		regex := loadRegex(p.db, true)
		for _, rxp := range regex {
			if rxp.Compiled.MatchString(msg) {
				reply, err := rxp.handle(p, msg, m.Author.Mention())
				if err != nil {
					log.Println(err)
					return
				}
				s.ChannelMessageSend(m.ChannelID, reply)
				p.randomReset <- true
				return
			}
		}
	}

	regex := loadRegex(p.db, false)
	for _, rxp := range regex {
		if rxp.Compiled.MatchString(msg) {
			reply, err := rxp.handle(p, msg, m.Author.Mention())
			if err != nil {
				log.Println(err)
				return
			}
			s.ChannelMessageSend(m.ChannelID, reply)
			p.randomReset <- true
			return
		}
	}
}
