package pail

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/FryDay/pail/sqlite"
	log "github.com/sirupsen/logrus"
)

type Fact struct {
	ID             int64  `db:"id"`
	Fact           string `db:"fact"`
	Verb           string `db:"verb"`
	Tidbit         string `db:"tidbit"`
	ReplacedTidbit string `db:"-"`
}

func NewFact(fact, tidbit, verb string) *Fact {
	return &Fact{Fact: fact, Tidbit: tidbit, Verb: verb}
}

func findFact(db *sqlite.DB, msg, author string) (fact *Fact, err error) {
	fact = &Fact{}
	msg = strings.ToLower(punctuationRegex.ReplaceAllString(msg, ""))
	log.Debug("Fact lookup: ", msg)
	if err = db.Get(`select id, fact, tidbit, verb from fact where fact=:fact order by random() limit 1`, map[string]interface{}{"fact": msg}, fact); err != nil {
		if err == sqlite.ErrNoRows {
			log.Debug("No fact found")
			return nil, nil
		}
		return nil, err
	}
	if varRegex.MatchString(fact.Tidbit) {
		fact.ReplacedTidbit = fact.Tidbit
		vars := varRegex.FindAllString(fact.Tidbit, -1)
		log.Debug(fmt.Sprintf("Fact var match: %v", vars))
		availVars := []string{}
		db.Select(`select name from var`, nil, &availVars)
		for _, origVar := range vars {
			varValue := &VarValue{Var: &Var{}, Value: &Value{}}
			if s := whoRegex.FindString(origVar); s != "" {
				varValue.Name = s
				varValue.Value.Value = author
			} else {
				for _, v := range availVars {
					r := regexp.MustCompile(fmt.Sprintf(`\$(%s)`, v))
					if found := r.FindString(origVar); found > "" {
						origVar = found
						break
					}
				}
				origVar = origVar[1:]
				varValue, err = findVarValue(db, origVar)
				if err != nil {
					return nil, err
				}
				if varValue == nil {
					return nil, nil
				}
			}
			fact.ReplacedTidbit = strings.Replace(fact.ReplacedTidbit, varValue.Name, varValue.Value.Value, 1)
		}
	}
	return fact, nil
}

func getRandomFact(db *sqlite.DB) (fact *Fact, err error) {
	fact = &Fact{}
	if err = db.Get(`select id, fact, tidbit, verb from fact where tidbit not like '%$who%' order by random() limit 1`, nil, fact); err != nil {
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
			varValue := &VarValue{}
			for _, v := range availVars {
				r := regexp.MustCompile(fmt.Sprintf(`\$(%s)`, v))
				if found := r.FindString(origVar); found > "" {
					origVar = found
					break
				}
			}
			origVar = origVar[1:]
			varValue, err = findVarValue(db, origVar)
			if err != nil {
				return nil, err
			}
			if varValue == nil {
				return nil, nil
			}
			fact.ReplacedTidbit = strings.Replace(fact.ReplacedTidbit, varValue.Name, varValue.Value.Value, 1)
		}
	}
	return fact, nil
}

func (f *Fact) insert(db *sqlite.DB) (err error) {
	log.Debug(fmt.Sprintf("New fact: '%s' '%s' '%s'", f.Fact, f.Tidbit, f.Verb))
	f.ID, err = db.Insert(`insert into fact (fact, tidbit, verb) values (lower(:fact), :tidbit, :verb)`, f)
	return err
}

func (f *Fact) delete(db *sqlite.DB) error {
	return db.Delete(`delete from fact where id=:id`, f)
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
