// Package model defines the independently addressable records exposed by codoc.
package model

import "strings"

type Position struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type Package struct {
	Kind            string           `json:"kind"`
	ImportPath      string           `json:"import_path"`
	Name            string           `json:"name"`
	Overview        string           `json:"overview"`
	Workflows       []Workflow       `json:"workflows"`
	Contracts       []Contract       `json:"contracts"`
	Symbols         []Symbol         `json:"symbols"`
	DocumentedTests []DocumentedTest `json:"documented_tests"`
	Gaps            []string         `json:"gaps"`
}

type DocumentedTest struct {
	Kind        string   `json:"kind"`
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary,omitempty"`
	TestName    string   `json:"test_name"`
	Code        string   `json:"code,omitempty"`
	CodeOmitted bool     `json:"code_omitted,omitempty"`
	Source      Position `json:"source"`
	Status      string   `json:"status"`
	Output      string   `json:"output,omitempty"`
}

type Narrative struct {
	Kind            string           `json:"kind"`
	ImportPath      string           `json:"import_path"`
	Name            string           `json:"name"`
	Overview        string           `json:"overview"`
	DocumentedTests []DocumentedTest `json:"documented_tests"`
	Passed          bool             `json:"passed"`
	Output          string           `json:"output,omitempty"`
}

type Workflow struct {
	Kind             string   `json:"kind"`
	ID               string   `json:"id"`
	Summary          string   `json:"summary,omitempty"`
	ExampleName      string   `json:"example_name"`
	PrimarySymbol    string   `json:"primary_symbol,omitempty"`
	Code             string   `json:"code"`
	ExpectedOutput   string   `json:"expected_output,omitempty"`
	EmptyOutput      bool     `json:"empty_output,omitempty"`
	RelatedSymbols   []string `json:"related_symbols"`
	RelatedContracts []string `json:"related_contracts"`
	Verification     string   `json:"verification"`
	Source           Position `json:"source"`
}

type Contract struct {
	Kind           string   `json:"kind"`
	ID             string   `json:"id"`
	Summary        string   `json:"summary"`
	TestName       string   `json:"test_name"`
	RelatedSymbols []string `json:"related_symbols"`
	Verification   string   `json:"verification"`
	Source         Position `json:"source"`
}

type Symbol struct {
	Kind             string   `json:"kind"`
	ID               string   `json:"id"`
	DeclarationKind  string   `json:"declaration_kind"`
	Signature        string   `json:"signature"`
	Doc              string   `json:"doc,omitempty"`
	RelatedWorkflows []string `json:"related_workflows"`
	RelatedContracts []string `json:"related_contracts"`
	Source           Position `json:"source"`
}

type Verification struct {
	Kind   string `json:"kind"`
	Target string `json:"target"`
	Passed bool   `json:"passed"`
	Output string `json:"output,omitempty"`
}

// OrientationSymbols returns a capped map of top-level public API entry points.
// Methods remain independently retrievable but do not crowd the package overview.
func (p Package) OrientationSymbols(limit int) []Symbol {
	items := make([]Symbol, 0, min(len(p.Symbols), limit))
	for _, symbol := range p.Symbols {
		if strings.Contains(symbol.ID, ".") {
			continue
		}
		items = append(items, symbol)
		if len(items) == limit {
			break
		}
	}
	return items
}

// OrientationSymbolCount reports the number of top-level public API entries.
func (p Package) OrientationSymbolCount() int {
	count := 0
	for _, symbol := range p.Symbols {
		if !strings.Contains(symbol.ID, ".") {
			count++
		}
	}
	return count
}
