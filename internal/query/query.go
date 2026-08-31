// Package query ranks existing documentation records using deterministic lexical matching.
package query

import (
	"sort"
	"strings"
	"unicode"

	"github.com/bartdeboer/go-codoc/internal/model"
)

type Match struct {
	Kind    string         `json:"kind"`
	ID      string         `json:"id"`
	Summary string         `json:"summary,omitempty"`
	Score   int            `json:"score"`
	Source  model.Position `json:"source"`
}

func Search(pkg model.Package, text string, limit int) []Match {
	terms := words(text)
	matches := []Match{}
	for _, workflow := range pkg.Workflows {
		document := workflow.ID + " " + workflow.Summary + " " + workflow.Code + " " + strings.Join(workflow.RelatedSymbols, " ")
		matches = addMatch(matches, terms, "workflow", workflow.ID, workflow.Summary, document, workflow.Source)
	}
	for _, contract := range pkg.Contracts {
		document := contract.ID + " " + contract.Summary + " " + strings.Join(contract.RelatedSymbols, " ")
		matches = addMatch(matches, terms, "contract", contract.ID, contract.Summary, document, contract.Source)
	}
	for _, symbol := range pkg.Symbols {
		summary := symbol.Doc
		if summary == "" {
			summary = symbol.Signature
		}
		document := symbol.ID + " " + symbol.Doc + " " + symbol.Signature
		matches = addMatch(matches, terms, "symbol", symbol.ID, summary, document, symbol.Source)
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
func addMatch(matches []Match, terms []string, kind, id, summary, document string, source model.Position) []Match {
	score := rank(terms, document)
	if score == 0 {
		return matches
	}
	return append(matches, Match{Kind: kind, ID: id, Summary: summary, Score: score, Source: source})
}
func rank(terms []string, document string) int {
	lower := strings.ToLower(document)
	fields := strings.Fields(lower)
	score := 0
	for _, term := range terms {
		if !strings.Contains(lower, term) {
			continue
		}
		score++
		if len(fields) > 0 && strings.Contains(fields[0], term) {
			score += 2
		}
	}
	return score
}
func words(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}
