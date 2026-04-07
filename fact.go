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

// resolveVars resolves variable placeholders in the fact's tidbit.
// Returns false if a variable could not be resolved (caller should treat as no result).
func (f *Fact) resolveVars(db *sqlite.DB, author string) (bool, error) {
	if !varRegex.MatchString(f.Tidbit) {
		return true, nil
	}
	f.ReplacedTidbit = f.Tidbit
	vars := varRegex.FindAllString(f.Tidbit, -1)
	log.Debug(fmt.Sprintf("Fact var match: %v", vars))
	availVars := []string{}
	if err := db.Select(`select name from var`, nil, &availVars); err != nil {
		return false, err
	}
	for _, origVar := range vars {
		if author != "" {
			if s := whoRegex.FindString(origVar); s != "" {
				f.ReplacedTidbit = strings.Replace(f.ReplacedTidbit, s, author, 1)
				continue
			}
		}
		for _, v := range availVars {
			r := regexp.MustCompile(fmt.Sprintf(`\$(%s)`, regexp.QuoteMeta(v)))
			if found := r.FindString(origVar); found > "" {
				origVar = found
				break
			}
		}
		origVar = origVar[1:]
		varValue, err := findVarValue(db, origVar)
		if err != nil {
			return false, err
		}
		if varValue == nil {
			return false, nil
		}
		f.ReplacedTidbit = strings.Replace(f.ReplacedTidbit, varValue.Name, varValue.Value, 1)
	}
	return true, nil
}

func findFact(db *sqlite.DB, msg, author string) (*Fact, error) {
	fact := &Fact{}
	msg = strings.ToLower(punctuationRegex.ReplaceAllString(msg, ""))
	log.Debug("Fact lookup: ", msg)
	if err := db.Get(`select id, fact, tidbit, verb from fact where fact=:fact order by random() limit 1`, map[string]interface{}{"fact": msg}, fact); err != nil {
		if err == sqlite.ErrNoRows {
			log.Debug("No fact found")
			return nil, nil
		}
		return nil, err
	}
	if ok, err := fact.resolveVars(db, author); err != nil || !ok {
		return nil, err
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
	if ok, err := fact.resolveVars(db, ""); err != nil || !ok {
		return nil, err
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
