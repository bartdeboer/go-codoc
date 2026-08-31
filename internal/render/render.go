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
	fmt.Fprintf(w, "Package %s\n", p.Name)
	if p.Overview != "" {
		fmt.Fprintf(w, "\n%s\n", p.Overview)
	}
	if symbols := p.OrientationSymbols(12); len(symbols) > 0 {
		fmt.Fprintln(w, "\nAPI:")
		for _, kind := range []string{"type", "func", "const", "var"} {
			for _, symbol := range symbols {
				if symbol.DeclarationKind != kind {
					continue
				}
				summary := firstSentence(symbol.Doc)
				if summary == "" {
					summary = "undocumented"
				}
				fmt.Fprintf(w, "- %s (%s): %s\n", symbol.ID, kind, summary)
			}
		}
	}
	fmt.Fprintf(w, "\nWorkflows: %d\nContracts: %d\n", len(p.Workflows), len(p.Contracts))
	if len(p.Gaps) > 0 {
		fmt.Fprintln(w, "\nDocumentation gaps:")
		for _, gap := range p.Gaps {
			fmt.Fprintf(w, "- %s\n", gap)
		}
	}
}
func Workflows(w io.Writer, xs []model.Workflow) {
	if len(xs) == 0 {
		fmt.Fprintln(w, "No documented workflows.")
		return
	}
	for i, x := range xs {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, x.ID)
		indent(w, x.Summary)
	}
}
func Workflow(w io.Writer, x model.Workflow) {
	fmt.Fprintf(w, "Workflow: %s\n\n%s\n", x.ID, x.Summary)
	if x.PrimarySymbol != "" {
		fmt.Fprintf(w, "\nPrimary symbol: %s\n", x.PrimarySymbol)
	}
	fmt.Fprintln(w, "\nExample:")
	indent(w, x.Code)
	if x.ExpectedOutput != "" || x.EmptyOutput {
		fmt.Fprintln(w, "\nExpected output:")
		if x.EmptyOutput {
			indent(w, "(empty)")
		} else {
			indent(w, x.ExpectedOutput)
		}
	}
	fmt.Fprintf(w, "\nVerification: %s\nSource: %s\n", x.Verification, source(x.Source))
	list(w, "Related symbols", x.RelatedSymbols)
	list(w, "Related contracts", x.RelatedContracts)
}
func Contracts(w io.Writer, xs []model.Contract) {
	if len(xs) == 0 {
		fmt.Fprintln(w, "No documented contracts.")
		return
	}
	for i, x := range xs {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, x.ID)
		indent(w, x.Summary)
	}
}
func Contract(w io.Writer, x model.Contract) {
	fmt.Fprintf(w, "Contract: %s\n\n%s\n\nTest: %s\nVerification: %s\nSource: %s\n", x.ID, x.Summary, x.TestName, x.Verification, source(x.Source))
	list(w, "Related symbols", x.RelatedSymbols)
}
func Symbol(w io.Writer, x model.Symbol) {
	fmt.Fprintf(w, "%s\n", x.Signature)
	if x.Doc != "" {
		fmt.Fprintf(w, "\n%s\n", x.Doc)
	} else {
		fmt.Fprintln(w, "\nUndocumented.")
	}
	fmt.Fprintf(w, "\nSource: %s\n", source(x.Source))
	list(w, "Related workflows", x.RelatedWorkflows)
	list(w, "Related contracts", x.RelatedContracts)
}
func Matches(w io.Writer, xs []query.Match) {
	if len(xs) == 0 {
		fmt.Fprintln(w, "No matching documentation.")
		return
	}
	for i, x := range xs {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s: %s\n", x.Kind, x.ID)
		indent(w, x.Summary)
		fmt.Fprintf(w, "    Source: %s\n", source(x.Source))
	}
}
func Verification(w io.Writer, x model.Verification) {
	status := "FAIL"
	if x.Passed {
		status = "PASS"
	}
	fmt.Fprintf(w, "%s %s %s\n", status, x.Kind, x.Target)
	if x.Output != "" {
		fmt.Fprint(w, x.Output)
		if !strings.HasSuffix(x.Output, "\n") {
			fmt.Fprintln(w)
		}
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
func source(p model.Position) string {
	if p.File == "" {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", p.File, p.Line)
}
func firstSentence(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, ". "); i >= 0 {
		return s[:i+1]
	}
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
