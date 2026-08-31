// Package model defines the independently addressable records exposed by codoc.
package model

type Position struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

type Package struct {
	Kind       string     `json:"kind"`
	ImportPath string     `json:"import_path"`
	Name       string     `json:"name"`
	Overview   string     `json:"overview"`
	Workflows  []Workflow `json:"workflows,omitempty"`
	Contracts  []Contract `json:"contracts,omitempty"`
	Symbols    []Symbol   `json:"symbols,omitempty"`
}

type Workflow struct {
	Kind             string   `json:"kind"`
	ID               string   `json:"id"`
	Summary          string   `json:"summary,omitempty"`
	ExampleName      string   `json:"example_name"`
	Code             string   `json:"code"`
	RelatedSymbols   []string `json:"related_symbols,omitempty"`
	RelatedContracts []string `json:"related_contracts,omitempty"`
	Source           Position `json:"source"`
}

type Contract struct {
	Kind           string   `json:"kind"`
	ID             string   `json:"id"`
	Summary        string   `json:"summary"`
	TestName       string   `json:"test_name"`
	RelatedSymbols []string `json:"related_symbols,omitempty"`
	Source         Position `json:"source"`
}

type Symbol struct {
	Kind             string   `json:"kind"`
	ID               string   `json:"id"`
	DeclarationKind  string   `json:"declaration_kind"`
	Signature        string   `json:"signature"`
	Doc              string   `json:"doc,omitempty"`
	RelatedWorkflows []string `json:"related_workflows,omitempty"`
	RelatedContracts []string `json:"related_contracts,omitempty"`
	Source           Position `json:"source"`
}
