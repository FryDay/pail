package pail

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

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
		log.Debug(fmt.Sprintf("Mention: %+v. Mention ID: %s. My ID: %s. Part 0: %s", mention, mention.ID, s.State.User.ID, parts[0]))
		if mention.ID == s.State.User.ID && (fmt.Sprintf("<@%s>", s.State.User.ID) == parts[0] || fmt.Sprintf("<@!%s>", s.State.User.ID) == parts[0]) {
			msg = strings.Join(parts[1:], " ")
			mentioned = true
			log.Debug("My mention: ", msg)
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
			log.Error(err)
			return
		}
		p.Say(m.ChannelID, reply)
		return
	}

	if mentioned {
		regex := getAllRegex(p.db, true)
		for _, rxp := range regex {
			if rxp.Compiled.MatchString(msg) {
				log.Debug("Mention regex found: ", rxp.Expression)
				reply, err := rxp.handle(p, msg, m.Author.Mention())
				if err != nil {
					log.Error(err)
					return
				}
				p.Say(m.ChannelID, reply)
				return
			}
		}
		log.Debug("No mention regex found")
	}

	regex := getAllRegex(p.db, false)
	for _, rxp := range regex {
		if rxp.Compiled.MatchString(msg) {
			log.Debug("Regex found: ", rxp.Expression)
			reply, err := rxp.handle(p, msg, m.Author.Mention())
			if err != nil {
				log.Error(err)
				return
			}
			p.Say(m.ChannelID, reply)
			return
		}
	}
	log.Debug("No regex found")
}
