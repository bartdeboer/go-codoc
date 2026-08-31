// Package query ranks existing documentation records using deterministic lexical matching.
package query

import (
	"sort"
	"strings"
	"unicode"

	"github.com/bartdeboer/codoc/internal/model"
)

type Match struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Summary string `json:"summary,omitempty"`
	Score   int    `json:"score"`
}

func Search(pkg model.Package, text string, limit int) []Match {
	terms := words(text)
	var matches []Match
	for _, w := range pkg.Workflows {
		if score := rank(terms, w.ID+" "+w.Summary+" "+strings.Join(w.RelatedSymbols, " ")); score > 0 {
			matches = append(matches, Match{"workflow", w.ID, w.Summary, score})
		}
	}
	for _, c := range pkg.Contracts {
		if score := rank(terms, c.ID+" "+c.Summary+" "+strings.Join(c.RelatedSymbols, " ")); score > 0 {
			matches = append(matches, Match{"contract", c.ID, c.Summary, score})
		}
	}
	for _, s := range pkg.Symbols {
		if score := rank(terms, s.ID+" "+s.Doc); score > 0 {
			matches = append(matches, Match{"symbol", s.ID, s.Doc, score})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score == matches[j].Score {
			return matches[i].ID < matches[j].ID
		}
		return matches[i].Score > matches[j].Score
	})
	if limit > 0 && len(matches) > limit {
		return matches[:limit]
	}
	return matches
}
func rank(terms []string, document string) int {
	doc := strings.ToLower(document)
	score := 0
	for _, term := range terms {
		if strings.Contains(doc, term) {
			score++
			if strings.Contains(strings.ToLower(strings.Fields(document)[0]), term) {
				score += 2
			}
		}
	}
	return score
}
func words(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}
