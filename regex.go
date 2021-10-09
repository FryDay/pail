package pail

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"

	"github.com/FryDay/pail/sqlite"
)

type Regex struct {
	ID         int64          `db:"id"`
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

func (r *Regex) handle(p *Pail, msg, author string) (string, error) {
	switch r.Action {
	case "add":
		parts := r.Compiled.FindStringSubmatch(msg)
		if len(parts) != 4 {
			return "", fmt.Errorf("wrong syntax")
		}
		fact := NewFact(strings.TrimSpace(parts[1]), strings.TrimSpace(parts[3]), strings.TrimSpace(parts[2]))
		if err := fact.insert(p.db); err == nil {
			return fmt.Sprintf("Okay %s", author), nil
		}

	case "add_var":
		parts := r.Compiled.FindStringSubmatch(msg)
		if len(parts) != 3 {
			return "", fmt.Errorf("wrong syntax")
		}

		v, err := getVar(p.db, parts[1])
		if err != nil {
			log.Println(err)
			return "BZZZZZZZZZT!", nil
		}
		if v == nil {
			v = NewVar(parts[1])
			v.insert(p.db)
		}

		value, err := getValue(p.db, v.ID, parts[2])
		if err != nil {
			log.Println(err)
			return "BZZZZZZZZZT!", nil
		}
		if value != nil {
			return fmt.Sprintf("I already know that one %s!", author), nil
		}

		value = NewValue(v.ID, parts[2])
		if err := value.insert(p.db); err != nil {
			log.Println(err)
			return "BZZZZZZZZZT!", nil
		}
		return fmt.Sprintf("Okay %s", author), nil

	case "remove_var":
		parts := r.Compiled.FindStringSubmatch(msg)
		if len(parts) != 3 {
			return "", fmt.Errorf("wrong syntax")
		}

		v, err := getVar(p.db, parts[1])
		if err != nil {
			log.Println(err)
			return "BZZZZZZZZZT!", nil
		}
		if v == nil {
			return fmt.Sprintf("I don't know that one %s!", author), nil
		}

		value, err := getValue(p.db, v.ID, parts[2])
		if err != nil {
			log.Println(err)
			return "BZZZZZZZZZT!", nil
		}
		if value == nil {
			return fmt.Sprintf("I don't know that one %s!", author), nil
		}

		if err := value.delete(p.db); err != nil {
			log.Println(err)
			return "BZZZZZZZZZT!", nil
		}
		return fmt.Sprintf("Okay %s", author), nil

	case "forget":
		if p.lastFact != nil {
			err := p.lastFact.delete(p.db)
			if err != nil {
				log.Println(err)
				return "BZZZZZZZZZT!", nil
			}
			p.lastFact = nil
			return fmt.Sprintf("Okay %s, I forgot \"%s _%s_ %s\"", author, p.lastFact.Fact, p.lastFact.Verb, p.lastFact.Tidbit), nil
		}
		return fmt.Sprintf("I'm sorry %s, I can't let you do that...", author), nil

	case "inquiry":
		if p.lastFact != nil {
			return fmt.Sprintf("That was \"%s _%s_ %s\"", p.lastFact.Fact, p.lastFact.Verb, p.lastFact.Tidbit), nil
		}
		return "BZZZZZZZZZT!", nil

	case "replace":
		chance := rand.Intn(99) + 1
		if chance <= p.config.ReplaceChance {
			return r.Compiled.ReplaceAllString(msg, r.Sub), nil
		}
		return "", nil

	case "reply":
		return r.Sub, nil
	}

	return "", fmt.Errorf("action %s not found", r.Action)
}
