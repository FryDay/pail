package pail

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/FryDay/pail/sqlite"
)

type Regex struct {
	ID         int64          `db:"id"`
	Expression string         `db:"expression"`
	Action     string         `db:"action"`
	Sub        string         `db:"sub"`
	Compiled   *regexp.Regexp `db:"-"`
}

func getAllRegex(db *sqlite.DB, mention bool) []*Regex {
	regex := []*Regex{}
	db.Select(`select id, expression, action, sub from regex where mention=:mention`, map[string]interface{}{"mention": mention}, &regex)
	for _, r := range regex {
		r.Compiled = regexp.MustCompile(r.Expression)
	}
	return regex
}

func (r *Regex) handle(p *Pail, msg, author, channelID, messageID string) (string, error) {
	log.Debug("Regex action: ", r.Action)
	switch r.Action {
	case "add":
		parts := r.Compiled.FindStringSubmatch(msg)
		if len(parts) != 4 {
			return "", fmt.Errorf("wrong syntax")
		}
		fact := NewFact(strings.TrimSpace(parts[1]), strings.TrimSpace(parts[3]), strings.TrimSpace(parts[2]))
		if err := fact.insert(p.db); err == nil {
			p.lastFact = fact
			return fmt.Sprintf("Okay %s", author), nil
		} else {
			return fmt.Sprintf("I'm sorry %s, I can't let you do that...", author), nil
		}

	case "add_var":
		parts := r.Compiled.FindStringSubmatch(msg)
		if len(parts) != 3 {
			return "", fmt.Errorf("wrong syntax")
		}

		v, err := getVar(p.db, parts[1])
		if err != nil {
			return r.handleError(err), nil
		}
		if v == nil {
			v = NewVar(parts[1])
			v.insert(p.db)
		}

		value, err := getValue(p.db, v.ID, parts[2])
		if err != nil {
			return r.handleError(err), nil
		}
		if value != nil {
			return fmt.Sprintf("I already know that one %s!", author), nil
		}

		value = NewValue(v.ID, parts[2])
		if err := value.insert(p.db); err != nil {
			return r.handleError(err), nil
		}
		return fmt.Sprintf("Okay %s", author), nil

	case "remove_var":
		parts := r.Compiled.FindStringSubmatch(msg)
		if len(parts) != 3 {
			return "", fmt.Errorf("wrong syntax")
		}

		v, err := getVar(p.db, parts[1])
		if err != nil {
			return r.handleError(err), nil
		}
		if v == nil {
			return fmt.Sprintf("I don't know that one %s!", author), nil
		}

		value, err := getValue(p.db, v.ID, parts[2])
		if err != nil {
			return r.handleError(err), nil
		}
		if value == nil {
			return fmt.Sprintf("I don't know that one %s!", author), nil
		}

		if err := value.delete(p.db); err != nil {
			return r.handleError(err), nil
		}
		return fmt.Sprintf("Okay %s", author), nil

	case "forget":
		if p.lastFact != nil {
			err := p.lastFact.delete(p.db)
			if err != nil {
				return r.handleError(err), nil
			}
			response := fmt.Sprintf("Okay %s, I forgot \"%s _%s_ %s\"", author, p.lastFact.Fact, p.lastFact.Verb, p.lastFact.Tidbit)
			p.lastFact = nil
			return response, nil
		}
		log.Debug("Forget: Don't remember last fact.")
		return fmt.Sprintf("I'm sorry %s, I can't let you do that...", author), nil

	case "inquiry":
		if p.lastFact != nil {
			return fmt.Sprintf("That was \"%s _%s_ %s\"", p.lastFact.Fact, p.lastFact.Verb, p.lastFact.Tidbit), nil
		}
		log.Debug("Inquiry: Don't remember last fact.")
		return "BZZZZZZZZZT!", nil

	case "replace":
		chance := rand.Intn(99) + 1
		log.Debug(fmt.Sprintf("Replace chance: %d, configured chance: %d", chance, p.config.ReplaceChance))
		if chance <= p.config.ReplaceChance {
			return r.Compiled.ReplaceAllString(msg, r.Sub), nil
		}
		return "", nil

	case "reply":
		return r.Sub, nil

	case "react":
		return "", p.session.MessageReactionAdd(channelID, messageID, r.Sub)
	}

	return "", fmt.Errorf("action %s not found", r.Action)
}

func (r *Regex) handleError(err error) string {
	log.Error(err)
	return "BZZZZZZZZZT!"
}
