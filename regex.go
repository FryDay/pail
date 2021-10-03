package pail

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Regex struct {
	Expression string         `db:"expression"`
	Action     string         `db:"action"`
	Sub        string         `db:"sub"`
	Compiled   *regexp.Regexp `db:"-"`
}

func loadRegex(mention bool) []*Regex {
	regex := []*Regex{}
	db.Select(`select expression, action, sub from regex where mention=:mention`, map[string]interface{}{"mention": mention}, &regex)
	for _, r := range regex {
		r.Compiled = regexp.MustCompile(r.Expression)
	}
	return regex
}

func (r *Regex) handle(s *discordgo.Session, channelID, msg, author string) {
	switch r.Action {
	case "add":
		parts := r.Compiled.FindStringSubmatch(msg)
		if len(parts) != 4 {
			return
		}
		fact := NewFact(strings.TrimSpace(parts[1]), strings.TrimSpace(parts[3]), strings.TrimSpace(parts[2]))
		if err := fact.insert(); err == nil {
			s.ChannelMessageSend(channelID, fmt.Sprintf("Okay %s", author))
		}
	case "replace":
		s.ChannelMessageSend(channelID, r.Compiled.ReplaceAllString(msg, r.Sub))
	case "reply":
		s.ChannelMessageSend(channelID, r.Sub)
	}
}
