package pail

import (
	"fmt"
	"strings"

	"github.com/FryDay/pail/sqlite"
	"github.com/bwmarrin/discordgo"
)

type Fact struct {
	ID     int    `db:"id"`
	Fact   string `db:"fact"`
	Tidbit string `db:"tidbit"`
	Verb   string `db:"verb"`
}

func NewFact(fact, tidbit, verb string) *Fact {
	return &Fact{Fact: fact, Tidbit: tidbit, Verb: verb}
}

func getFact(msg string) (*Fact, error) {
	msg = punctuationRegex.ReplaceAllString(msg, "")
	fact := &Fact{}
	if err := db.Get(`select id, fact, tidbit, verb from fact where fact=:fact order by random() limit 1`, map[string]interface{}{"fact": msg}, fact); err != nil {
		if err == sqlite.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if varRegex.MatchString(fact.Tidbit) {
		vars := varRegex.FindAllString(fact.Tidbit, -1)
		for _, origVar := range vars {
			replaceVar := origVar[1:]
			db.Get(`select val.value from value val join var v on v.id = val.var_id where v.name=:name order by random() limit 1`, map[string]interface{}{"name": replaceVar}, &replaceVar)
			fact.Tidbit = strings.Replace(fact.Tidbit, origVar, replaceVar, 1)
		}
	}
	return fact, nil
}

func (f *Fact) insert() error {
	return db.NamedExec(`insert into fact (fact, tidbit, verb) values (:fact, :tidbit, :verb)`, f)
}

func (f *Fact) handle(s *discordgo.Session, channelID string) {
	switch f.Verb {
	case "<action>":
		s.ChannelMessageSend(channelID, fmt.Sprintf("_%s_", f.Tidbit))
	case "<reply>":
		s.ChannelMessageSend(channelID, f.Tidbit)
	}
}
