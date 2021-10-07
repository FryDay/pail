package pail

import (
	"regexp"
)

var (
	varRegex         = regexp.MustCompile(`\$(\w+)`)
	punctuationRegex = regexp.MustCompile(`[^\w\s]`)
	whoRegex         = regexp.MustCompile(`\$(who)`)
)
