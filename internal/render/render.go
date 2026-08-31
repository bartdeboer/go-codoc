// Package render writes codoc records as stable JSON or compact text.
package render

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bartdeboer/codoc/internal/model"
	"github.com/bartdeboer/codoc/internal/query"
)

func JSON(w io.Writer, value any) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(value)
}
func Package(w io.Writer, p model.Package) {
	fmt.Fprintf(w, "Package %s\n\n%s\n", p.Name, p.Overview)
	if len(p.Workflows) > 0 {
		fmt.Fprintln(w, "\nWorkflows:")
		for _, x := range p.Workflows {
			fmt.Fprintf(w, "- %s\n", x.ID)
		}
	}
	if len(p.Contracts) > 0 {
		fmt.Fprintln(w, "\nContracts:")
		for _, x := range p.Contracts {
			fmt.Fprintf(w, "- %s\n", x.ID)
		}
	}
}
func Workflows(w io.Writer, xs []model.Workflow) {
	for i, x := range xs {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, x.ID)
		indent(w, x.Summary)
	}
}
func Workflow(w io.Writer, x model.Workflow) {
	fmt.Fprintf(w, "Workflow: %s\n\n%s\n\nExample:\n", x.ID, x.Summary)
	indent(w, x.Code)
	list(w, "Related symbols", x.RelatedSymbols)
	list(w, "Related contracts", x.RelatedContracts)
}
func Contracts(w io.Writer, xs []model.Contract) {
	for i, x := range xs {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, x.ID)
		indent(w, x.Summary)
	}
}
func Contract(w io.Writer, x model.Contract) {
	fmt.Fprintf(w, "Contract: %s\n\n%s\n\nTest:\n    %s\n", x.ID, x.Summary, x.TestName)
	list(w, "Related symbols", x.RelatedSymbols)
}
func Symbol(w io.Writer, x model.Symbol) {
	fmt.Fprintf(w, "%s\n\n%s\n", x.Signature, x.Doc)
	list(w, "Related workflows", x.RelatedWorkflows)
	list(w, "Related contracts", x.RelatedContracts)
}
func Matches(w io.Writer, xs []query.Match) {
	for i, x := range xs {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s: %s\n", x.Kind, x.ID)
		indent(w, x.Summary)
	}
}
func indent(w io.Writer, s string) {
	for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
		fmt.Fprintf(w, "    %s\n", line)
	}
}
func list(w io.Writer, title string, xs []string) {
	if len(xs) == 0 {
		return
	}
	fmt.Fprintf(w, "\n%s:\n", title)
	for _, x := range xs {
		fmt.Fprintf(w, "- %s\n", x)
	}
}
