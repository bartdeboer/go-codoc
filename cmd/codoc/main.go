package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bartdeboer/codoc/internal/app"
	"github.com/bartdeboer/codoc/internal/model"
	"github.com/bartdeboer/go-clir"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out io.Writer) error {
	a := app.App{Out: out}
	if len(args) == 0 {
		pkg, err := a.Load(".")
		if err != nil {
			return err
		}
		return a.Package(pkg, app.Options{Format: "text"})
	}

	router := buildRouter(a)
	if clir.IsHelpRequest(args) {
		return router.FPrintHelp(ctx, out, clir.StripHelpToken(args))
	}
	return router.Run(ctx, args)
}

func buildRouter(a app.App) *clir.Router {
	r := clir.New()
	r.Routes(func(b *clir.Builder) {
		b.Describe("", "Inspect the current Go package by default; pass a package path when needed.")
		packages := clir.WithContext(b, func(req *clir.Request) (model.Package, error) {
			pattern := req.Params["package"]
			if pattern == "" {
				pattern = "."
			}
			return a.Load(pattern)
		})

		packages.Handle("package", "Show the current package.", packageHandler(a))
		packages.Handle("package <package>", "Show another package.", packageHandler(a))
		packages.Handle("workflows", "List workflows in the current package.", workflowsHandler(a))
		packages.Handle("workflows <package>", "List workflows in another package.", workflowsHandler(a))
		packages.Handle("workflow", "List workflows in the current package.", workflowsHandler(a))
		packages.Handle("workflow <workflow>", "Show a current-package workflow.", workflowHandler(a))
		packages.Handle("workflow <package> <workflow>", "Show a workflow from another package.", workflowHandler(a))
		packages.Handle("contracts", "List contracts in the current package.", contractsHandler(a))
		packages.Handle("contracts <package>", "List contracts in another package.", contractsHandler(a))
		packages.Handle("contract", "List contracts in the current package.", contractsHandler(a))
		packages.Handle("contract <contract>", "Show a current-package contract.", contractHandler(a))
		packages.Handle("contract <package> <contract>", "Show a contract from another package.", contractHandler(a))
		packages.Handle("symbol <symbol>", "Show a current-package API symbol.", symbolHandler(a))
		packages.Handle("symbol <package> <symbol>", "Show a symbol from another package.", symbolHandler(a))
		packages.Handle("query <query>", "Search the current package documentation.", queryHandler(a))
		packages.Handle("query <package> <query>", "Search another package documentation.", queryHandler(a))
	})
	return r
}

func packageHandler(a app.App) clir.ContextHandler[model.Package] {
	return func(req *clir.Request, p model.Package) error {
		o, err := options(req.Extra)
		if err != nil {
			return err
		}
		return a.Package(p, o)
	}
}
func workflowsHandler(a app.App) clir.ContextHandler[model.Package] {
	return func(req *clir.Request, p model.Package) error {
		o, err := options(req.Extra)
		if err != nil {
			return err
		}
		return a.Workflows(p, o)
	}
}
func workflowHandler(a app.App) clir.ContextHandler[model.Package] {
	return func(req *clir.Request, p model.Package) error {
		o, err := options(req.Extra)
		if err != nil {
			return err
		}
		return a.Workflow(p, req.Params["workflow"], o)
	}
}
func contractsHandler(a app.App) clir.ContextHandler[model.Package] {
	return func(req *clir.Request, p model.Package) error {
		o, err := options(req.Extra)
		if err != nil {
			return err
		}
		return a.Contracts(p, o)
	}
}
func contractHandler(a app.App) clir.ContextHandler[model.Package] {
	return func(req *clir.Request, p model.Package) error {
		o, err := options(req.Extra)
		if err != nil {
			return err
		}
		return a.Contract(p, req.Params["contract"], o)
	}
}
func symbolHandler(a app.App) clir.ContextHandler[model.Package] {
	return func(req *clir.Request, p model.Package) error {
		o, err := options(req.Extra)
		if err != nil {
			return err
		}
		return a.Symbol(p, req.Params["symbol"], o)
	}
}
func queryHandler(a app.App) clir.ContextHandler[model.Package] {
	return func(req *clir.Request, p model.Package) error {
		o, err := options(req.Extra)
		if err != nil {
			return err
		}
		return a.Query(p, req.Params["query"], o)
	}
}

func options(args []string) (app.Options, error) {
	o := app.Options{Format: "text"}
	for len(args) > 0 {
		if len(args) == 2 && args[0] == "--format" {
			o.Format = args[1]
			args = args[2:]
			continue
		}
		return o, fmt.Errorf("unexpected arguments: %s", strings.Join(args, " "))
	}
	if o.Format != "text" && o.Format != "json" {
		return o, fmt.Errorf("unsupported format %q (want text or json)", o.Format)
	}
	return o, nil
}
