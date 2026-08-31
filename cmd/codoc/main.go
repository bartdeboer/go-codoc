package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/bartdeboer/codoc/internal/app"
	"github.com/bartdeboer/codoc/internal/load"
	"github.com/bartdeboer/codoc/internal/model"
	"github.com/bartdeboer/go-clir"
)

type commandContext struct {
	source  *load.Package
	doc     model.Package
	options app.Options
}
type globalOptions struct{ workDir, pattern, format string }

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
func run(ctx context.Context, args []string, out io.Writer) error {
	globals, args, err := parseGlobalOptions(args)
	if err != nil {
		return err
	}
	a := app.App{Out: out}
	router := buildRouter(a, globals)
	if clir.IsHelpRequest(args) {
		return router.FPrintHelp(ctx, out, clir.StripHelpToken(args))
	}
	if len(args) == 0 {
		source, doc, err := a.Load(ctx, globals.workDir, globals.pattern)
		if err != nil {
			return err
		}
		return a.Narrative(ctx, source, doc, app.Options{Format: globals.format})
	}
	return router.Run(ctx, args)
}
func buildRouter(a app.App, g globalOptions) *clir.Router {
	r := clir.New()
	r.Routes(func(b *clir.Builder) {
		b.Describe("", "Inspect the current Go package. Global options: --json, --format text|json, -C <dir>, --package <pattern>.")
		commands := clir.WithContext(b, func(req *clir.Request) (commandContext, error) {
			source, doc, err := a.Load(req.Context(), g.workDir, g.pattern)
			return commandContext{source, doc, app.Options{Format: g.format}}, err
		})
		commands.Handle("package", "Show package orientation and API map.", func(_ *clir.Request, c commandContext) error { return a.Package(c.doc, c.options) })
		commands.Handle("workflows", "List workflows.", func(_ *clir.Request, c commandContext) error { return a.Workflows(c.doc, c.options) })
		commands.Handle("workflow <workflow>", "Show one workflow.", func(req *clir.Request, c commandContext) error {
			return a.Workflow(c.doc, req.Params["workflow"], c.options)
		})
		commands.Handle("contracts", "List documented contracts.", func(_ *clir.Request, c commandContext) error { return a.Contracts(c.doc, c.options) })
		commands.Handle("contract <contract>", "Show one documented contract.", func(req *clir.Request, c commandContext) error {
			return a.Contract(c.doc, req.Params["contract"], c.options)
		})
		commands.Handle("symbol <symbol>", "Show one public API symbol.", func(req *clir.Request, c commandContext) error {
			return a.Symbol(c.doc, req.Params["symbol"], c.options)
		})
		search := func(req *clir.Request, c commandContext) error { return a.Search(c.doc, req.Params["text"], c.options) }
		commands.Handle("search <text>", "Search documentation records.", search)
		commands.Handle("query <text>", "Alias for search.", search)
		commands.Handle("verify", "Run package tests explicitly.", func(req *clir.Request, c commandContext) error {
			return a.Verify(req.Context(), c.source, c.doc, "package", "", c.options)
		})
		commands.Handle("verify workflow <workflow>", "Run one workflow example.", func(req *clir.Request, c commandContext) error {
			return a.Verify(req.Context(), c.source, c.doc, "workflow", req.Params["workflow"], c.options)
		})
		commands.Handle("verify contract <contract>", "Run one contract test.", func(req *clir.Request, c commandContext) error {
			return a.Verify(req.Context(), c.source, c.doc, "contract", req.Params["contract"], c.options)
		})
	})
	return r
}
func parseGlobalOptions(args []string) (globalOptions, []string, error) {
	g := globalOptions{pattern: ".", format: "text"}
	rest := []string{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			g.format = "json"
		case "--format":
			if i+1 >= len(args) {
				return g, nil, fmt.Errorf("--format requires text or json")
			}
			i++
			g.format = args[i]
		case "-C":
			if i+1 >= len(args) {
				return g, nil, fmt.Errorf("-C requires a directory")
			}
			i++
			g.workDir = args[i]
		case "--package":
			if i+1 >= len(args) {
				return g, nil, fmt.Errorf("--package requires a pattern")
			}
			i++
			g.pattern = args[i]
		default:
			rest = append(rest, args[i])
		}
	}
	if g.format != "text" && g.format != "json" {
		return g, nil, fmt.Errorf("unsupported format %q (want text or json)", g.format)
	}
	return g, rest, nil
}
