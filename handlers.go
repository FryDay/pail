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

	fact, err := findFact(p.db, msg, m.Author.Mention())
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
		p.Say(m.ChannelID, reply)
		return
	}

	if mentioned {
		regex := getAllRegex(p.db, true)
		for _, rxp := range regex {
			if rxp.Compiled.MatchString(msg) {
				reply, err := rxp.handle(p, msg, m.Author.Mention())
				if err != nil {
					log.Println(err)
					return
				}
				p.Say(m.ChannelID, reply)
				return
			}
		}
	}

	regex := getAllRegex(p.db, false)
	for _, rxp := range regex {
		if rxp.Compiled.MatchString(msg) {
			reply, err := rxp.handle(p, msg, m.Author.Mention())
			if err != nil {
				log.Println(err)
				return
			}
			p.Say(m.ChannelID, reply)
			return
		}
	}
}
