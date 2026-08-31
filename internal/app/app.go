// Package app defines codoc's typed operations independently of CLI routing.
package app

import (
	"context"
	"fmt"
	"io"

	"github.com/bartdeboer/codoc/internal/index"
	"github.com/bartdeboer/codoc/internal/load"
	"github.com/bartdeboer/codoc/internal/model"
	"github.com/bartdeboer/codoc/internal/query"
	"github.com/bartdeboer/codoc/internal/render"
	"github.com/bartdeboer/codoc/internal/verify"
)

type App struct{ Out io.Writer }
type Options struct{ Format string }

type packageOverview struct {
	Kind       string    `json:"kind"`
	ImportPath string    `json:"import_path"`
	Name       string    `json:"name"`
	Overview   string    `json:"overview"`
	API        []apiItem `json:"api"`
	Workflows  []string  `json:"workflows"`
	Contracts  []string  `json:"contracts"`
	Gaps       []string  `json:"gaps"`
}
type apiItem struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Summary string `json:"summary,omitempty"`
}

func (a App) Load(ctx context.Context, workDir, pattern string) (*load.Package, model.Package, error) {
	source, err := load.PackageAtContext(ctx, workDir, pattern)
	if err != nil {
		return nil, model.Package{}, err
	}
	doc, err := index.Build(source)
	return source, doc, err
}
func (a App) Package(p model.Package, options Options) error {
	if options.Format == "json" {
		return render.JSON(a.Out, overviewOf(p))
	}
	render.Package(a.Out, p)
	return nil
}
func (a App) Workflows(p model.Package, options Options) error {
	if options.Format == "json" {
		return render.JSON(a.Out, nonNil(p.Workflows))
	}
	render.Workflows(a.Out, p.Workflows)
	return nil
}
func (a App) Workflow(p model.Package, id string, options Options) error {
	for _, workflow := range p.Workflows {
		if workflow.ID != id {
			continue
		}
		if options.Format == "json" {
			workflow.RelatedSymbols = nonNil(workflow.RelatedSymbols)
			workflow.RelatedContracts = nonNil(workflow.RelatedContracts)
			return render.JSON(a.Out, workflow)
		}
		render.Workflow(a.Out, workflow)
		return nil
	}
	return fmt.Errorf("workflow %q not found", id)
}
func (a App) Contracts(p model.Package, options Options) error {
	if options.Format == "json" {
		return render.JSON(a.Out, nonNil(p.Contracts))
	}
	render.Contracts(a.Out, p.Contracts)
	return nil
}
func (a App) Contract(p model.Package, id string, options Options) error {
	for _, contract := range p.Contracts {
		if contract.ID != id {
			continue
		}
		if options.Format == "json" {
			contract.RelatedSymbols = nonNil(contract.RelatedSymbols)
			return render.JSON(a.Out, contract)
		}
		render.Contract(a.Out, contract)
		return nil
	}
	return fmt.Errorf("contract %q not found", id)
}
func (a App) Symbol(p model.Package, id string, options Options) error {
	for _, symbol := range p.Symbols {
		if symbol.ID != id {
			continue
		}
		if options.Format == "json" {
			symbol.RelatedWorkflows = nonNil(symbol.RelatedWorkflows)
			symbol.RelatedContracts = nonNil(symbol.RelatedContracts)
			return render.JSON(a.Out, symbol)
		}
		render.Symbol(a.Out, symbol)
		return nil
	}
	return fmt.Errorf("symbol %q not found", id)
}
func (a App) Search(p model.Package, text string, options Options) error {
	matches := query.Search(p, text, 5)
	if options.Format == "json" {
		return render.JSON(a.Out, nonNil(matches))
	}
	render.Matches(a.Out, matches)
	return nil
}
func (a App) Verify(ctx context.Context, source *load.Package, doc model.Package, kind, id string, options Options) error {
	testName, err := verificationTest(doc, kind, id)
	if err != nil {
		return err
	}
	result := verify.Run(ctx, source, kind, id, testName)
	if options.Format == "json" {
		_ = render.JSON(a.Out, result)
	} else {
		render.Verification(a.Out, result)
	}
	if !result.Passed {
		return fmt.Errorf("verification failed")
	}
	return nil
}

func overviewOf(p model.Package) packageOverview {
	view := packageOverview{Kind: p.Kind, ImportPath: p.ImportPath, Name: p.Name, Overview: p.Overview, API: []apiItem{}, Workflows: []string{}, Contracts: []string{}, Gaps: nonNil(p.Gaps)}
	for _, symbol := range p.OrientationSymbols(12) {
		view.API = append(view.API, apiItem{symbol.ID, symbol.DeclarationKind, symbol.Doc})
	}
	for _, workflow := range p.Workflows {
		view.Workflows = append(view.Workflows, workflow.ID)
	}
	for _, contract := range p.Contracts {
		view.Contracts = append(view.Contracts, contract.ID)
	}
	return view
}
func verificationTest(doc model.Package, kind, id string) (string, error) {
	switch kind {
	case "package":
		return "", nil
	case "workflow":
		for _, workflow := range doc.Workflows {
			if workflow.ID == id {
				return workflow.ExampleName, nil
			}
		}
		return "", fmt.Errorf("workflow %q not found", id)
	case "contract":
		for _, contract := range doc.Contracts {
			if contract.ID == id {
				return contract.TestName, nil
			}
		}
		return "", fmt.Errorf("contract %q not found", id)
	default:
		return "", fmt.Errorf("verification kind must be workflow or contract")
	}
}
func nonNil[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}
