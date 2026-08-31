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
	router := buildRouter(a)
	if len(args) == 0 || clir.IsHelpRequest(args) {
		return router.FPrintHelp(ctx, out, clir.StripHelpToken(args))
	}
	return router.Run(ctx, args)
}

func buildRouter(a app.App) *clir.Router {
	r := clir.New()
	r.Routes(func(b *clir.Builder) {
		b.Describe("", "Retrieve compact, code-derived documentation for Go packages.")
		packages := clir.WithContext(b, func(req *clir.Request) (model.Package, error) { return a.Load(req.Params["package"]) })
		packages.Handle("package <package>", "Show package orientation.", func(req *clir.Request, p model.Package) error {
			o, e := options(req.Extra)
			if e != nil {
				return e
			}
			return a.Package(p, o)
		})
		packages.Handle("workflows <package>", "List workflows.", func(req *clir.Request, p model.Package) error {
			o, e := options(req.Extra)
			if e != nil {
				return e
			}
			return a.Workflows(p, o)
		})
		packages.Handle("workflow <package> <workflow>", "Show one workflow.", func(req *clir.Request, p model.Package) error {
			o, e := options(req.Extra)
			if e != nil {
				return e
			}
			return a.Workflow(p, req.Params["workflow"], o)
		})
		packages.Handle("contracts <package>", "List documented contracts.", func(req *clir.Request, p model.Package) error {
			o, e := options(req.Extra)
			if e != nil {
				return e
			}
			return a.Contracts(p, o)
		})
		packages.Handle("contract <package> <contract>", "Show one documented contract.", func(req *clir.Request, p model.Package) error {
			o, e := options(req.Extra)
			if e != nil {
				return e
			}
			return a.Contract(p, req.Params["contract"], o)
		})
		packages.Handle("symbol <package> <symbol>", "Show one API symbol.", func(req *clir.Request, p model.Package) error {
			o, e := options(req.Extra)
			if e != nil {
				return e
			}
			return a.Symbol(p, req.Params["symbol"], o)
		})
		packages.Handle("query <package> <query>", "Lexically rank documentation records.", func(req *clir.Request, p model.Package) error {
			o, e := options(req.Extra)
			if e != nil {
				return e
			}
			return a.Query(p, req.Params["query"], o)
		})
	})
	return r
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
