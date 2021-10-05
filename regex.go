package pail

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/FryDay/pail/sqlite"
	"github.com/bwmarrin/discordgo"
)

type Regex struct {
	ID         int            `db:"id"`
	Expression string         `db:"expression"`
	Action     string         `db:"action"`
	Sub        string         `db:"sub"`
	Compiled   *regexp.Regexp `db:"-"`
}

func loadRegex(db *sqlite.DB, mention bool) []*Regex {
	regex := []*Regex{}
	db.Select(`select id, expression, action, sub from regex where mention=:mention`, map[string]interface{}{"mention": mention}, &regex)
	for _, r := range regex {
		r.Compiled = regexp.MustCompile(r.Expression)
	}
	return regex
}

func (r *Regex) handle(p *Pail, s *discordgo.Session, channelID, msg, author string) {
	switch r.Action {
	case "add":
		parts := r.Compiled.FindStringSubmatch(msg)
		if len(parts) != 4 {
			return
		}
		fact := NewFact(strings.TrimSpace(parts[1]), strings.TrimSpace(parts[3]), strings.TrimSpace(parts[2]))
		if err := fact.insert(p.db); err == nil {
			s.ChannelMessageSend(channelID, fmt.Sprintf("Okay %s", author))
		}
	case "forget":
		if p.lastFact != nil {
			err := p.lastFact.delete(p.db)
			if err != nil {
				s.ChannelMessageSend(channelID, "BZZZZZZZZZT!")
				return
			}
			s.ChannelMessageSend(channelID, fmt.Sprintf("Okay %s, I forgot \"%s _%s_ %s\"", author, p.lastFact.Fact, p.lastFact.Verb, p.lastFact.Tidbit))
			return
		}
		s.ChannelMessageSend(channelID, fmt.Sprintf("I'm sorry %s, I can't let you do that...", author))
	case "inquiry":
		if p.lastFact != nil {
			s.ChannelMessageSend(channelID, fmt.Sprintf("That was \"%s _%s_ %s\"", p.lastFact.Fact, p.lastFact.Verb, p.lastFact.Tidbit))
			return
		}
		s.ChannelMessageSend(channelID, "BZZZZZZZZZT!")
	case "replace":
		chance := rand.Intn(99) + 1
		if chance <= 5 {
			s.ChannelMessageSend(channelID, r.Compiled.ReplaceAllString(msg, r.Sub))
		}
	case "reply":
		s.ChannelMessageSend(channelID, r.Sub)
	}
}
