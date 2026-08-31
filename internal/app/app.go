// Package app defines codoc's typed command operations independently of CLI routing.
package app

import (
	"fmt"
	"io"

	"github.com/bartdeboer/codoc/internal/index"
	"github.com/bartdeboer/codoc/internal/load"
	"github.com/bartdeboer/codoc/internal/model"
	"github.com/bartdeboer/codoc/internal/query"
	"github.com/bartdeboer/codoc/internal/render"
)

type App struct{ Out io.Writer }
type Options struct{ Format string }

func (a App) Load(pattern string) (model.Package, error) {
	p, err := load.PackageAt(pattern)
	if err != nil {
		return model.Package{}, err
	}
	return index.Build(p)
}
func (a App) Package(p model.Package, o Options) error {
	if o.Format == "json" {
		view := struct {
			Kind       string   `json:"kind"`
			ImportPath string   `json:"import_path"`
			Name       string   `json:"name"`
			Overview   string   `json:"overview"`
			Workflows  []string `json:"workflows"`
			Contracts  []string `json:"contracts"`
		}{Kind: p.Kind, ImportPath: p.ImportPath, Name: p.Name, Overview: p.Overview}
		for _, workflow := range p.Workflows {
			view.Workflows = append(view.Workflows, workflow.ID)
		}
		for _, contract := range p.Contracts {
			view.Contracts = append(view.Contracts, contract.ID)
		}
		return render.JSON(a.Out, view)
	}
	render.Package(a.Out, p)
	return nil
}
func (a App) Workflows(p model.Package, o Options) error {
	if o.Format == "json" {
		return render.JSON(a.Out, p.Workflows)
	}
	render.Workflows(a.Out, p.Workflows)
	return nil
}
func (a App) Workflow(p model.Package, id string, o Options) error {
	for _, x := range p.Workflows {
		if x.ID == id {
			if o.Format == "json" {
				return render.JSON(a.Out, x)
			}
			render.Workflow(a.Out, x)
			return nil
		}
	}
	return fmt.Errorf("workflow %q not found", id)
}
func (a App) Contracts(p model.Package, o Options) error {
	if o.Format == "json" {
		return render.JSON(a.Out, p.Contracts)
	}
	render.Contracts(a.Out, p.Contracts)
	return nil
}
func (a App) Contract(p model.Package, id string, o Options) error {
	for _, x := range p.Contracts {
		if x.ID == id {
			if o.Format == "json" {
				return render.JSON(a.Out, x)
			}
			render.Contract(a.Out, x)
			return nil
		}
	}
	return fmt.Errorf("contract %q not found", id)
}
func (a App) Symbol(p model.Package, id string, o Options) error {
	for _, x := range p.Symbols {
		if x.ID == id {
			if o.Format == "json" {
				return render.JSON(a.Out, x)
			}
			render.Symbol(a.Out, x)
			return nil
		}
	}
	return fmt.Errorf("symbol %q not found", id)
}
func (a App) Query(p model.Package, text string, o Options) error {
	x := query.Search(p, text, 5)
	if o.Format == "json" {
		return render.JSON(a.Out, x)
	}
	render.Matches(a.Out, x)
	return nil
}
