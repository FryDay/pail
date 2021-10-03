package pail

import (
	"regexp"

	"github.com/FryDay/pail/sqlite"
)

var (
	db               *sqlite.DB
	varRegex         = regexp.MustCompile(`\$(\w+)`)
	punctuationRegex = regexp.MustCompile(`[^\w\s]`)
)
