package pail

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/FryDay/pail/sqlite"
)

type Fact struct {
	ID             int    `db:"id"`
	Fact           string `db:"fact"`
	Tidbit         string `db:"tidbit"`
	ReplacedTidbit string `db:"-"`
	Verb           string `db:"verb"`
}

type Value struct {
	Var   string `db:"var"`
	Value string `db:"value"`
}

func NewFact(fact, tidbit, verb string) *Fact {
	return &Fact{Fact: fact, Tidbit: tidbit, Verb: verb}
}

func getFact(db *sqlite.DB, msg, author string) (*Fact, error) {
	msg = strings.ToLower(punctuationRegex.ReplaceAllString(msg, ""))
	fact := &Fact{}
	if err := db.Get(`select id, fact, tidbit, verb from fact where fact=:fact order by random() limit 1`, map[string]interface{}{"fact": msg}, fact); err != nil {
		if err == sqlite.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if varRegex.MatchString(fact.Tidbit) {
		fact.ReplacedTidbit = fact.Tidbit
		vars := varRegex.FindAllString(fact.Tidbit, -1)
		availVars := []string{}
		db.Select(`select name from var`, nil, &availVars)
		for _, origVar := range vars {
			val := &Value{}
			if s := whoRegex.FindString(origVar); s != "" {
				val.Var = s
				val.Value = author
			} else {
				for _, v := range availVars {
					r := regexp.MustCompile(fmt.Sprintf(`\$(%s)`, v))
					if found := r.FindString(origVar); found > "" {
						origVar = found
						break
					}
				}
				origVar = origVar[1:]
				db.Get(`select v.name var, val.value from value val join var v on v.id = val.var_id where v.name=:name order by random() limit 1`, map[string]interface{}{"name": origVar}, val)
				if val.Var > "" {
					val.Var = fmt.Sprintf("$%s", val.Var)
				}
			}
			fact.ReplacedTidbit = strings.Replace(fact.ReplacedTidbit, val.Var, val.Value, 1)
		}
	}
	return fact, nil
}

func getRandomFact(db *sqlite.DB) (*Fact, error) {
	fact := &Fact{}
	if err := db.Get(`select id, fact, tidbit, verb from fact where tidbit not like '%$who%' order by random() limit 1`, nil, fact); err != nil {
		if err == sqlite.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if varRegex.MatchString(fact.Tidbit) {
		fact.ReplacedTidbit = fact.Tidbit
		vars := varRegex.FindAllString(fact.Tidbit, -1)
		availVars := []string{}
		db.Select(`select name from var`, nil, &availVars)
		for _, origVar := range vars {
			val := &Value{}
			for _, v := range availVars {
				r := regexp.MustCompile(fmt.Sprintf(`\$(%s)`, v))
				if found := r.FindString(origVar); found > "" {
					origVar = found
					break
				}
			}
			origVar = origVar[1:]
			db.Get(`select v.name var, val.value from value val join var v on v.id = val.var_id where v.name=:name order by random() limit 1`, map[string]interface{}{"name": origVar}, val)
			if val.Var > "" {
				val.Var = fmt.Sprintf("$%s", val.Var)
			}
			fact.ReplacedTidbit = strings.Replace(fact.ReplacedTidbit, val.Var, val.Value, 1)
		}
	}
	return fact, nil
}

func (f *Fact) insert(db *sqlite.DB) error {
	return db.NamedExec(`insert into fact (fact, tidbit, verb) values (lower(:fact), :tidbit, :verb)`, f)
}

func (f *Fact) delete(db *sqlite.DB) error {
	return db.NamedExec(`delete from fact where id=:id`, f)
}

func (f *Fact) handle() (string, error) {
	reply := f.Tidbit
	if f.ReplacedTidbit != "" {
		reply = f.ReplacedTidbit
	}
	switch f.Verb {
	case "<action>":
		return fmt.Sprintf("_%s_", reply), nil
	case "<reply>":
		return reply, nil
	}

	return "", fmt.Errorf("verb %s unknown", f.Verb)
}
