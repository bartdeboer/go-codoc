// Package index turns Go syntax into compact documentation records.
package index

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/bartdeboer/codoc/internal/load"
	"github.com/bartdeboer/codoc/internal/model"
)

var contractID = regexp.MustCompile(`^[a-z0-9]+(?:[/-][a-z0-9]+)*$`)

func Build(pkg *load.Package) (model.Package, error) {
	result := model.Package{Kind: "package", ImportPath: pkg.ImportPath, Name: pkg.Name}
	for _, file := range pkg.Files {
		if file.Name.Name == pkg.Name && result.Overview == "" && file.Doc != nil {
			result.Overview = clean(file.Doc.Text())
		}
		for _, decl := range file.Decls {
			if err := inspectDecl(pkg, file, decl, &result); err != nil {
				return model.Package{}, err
			}
		}
	}
	if err := validateContracts(result.Contracts); err != nil {
		return model.Package{}, err
	}
	link(&result)
	sort.Slice(result.Workflows, func(i, j int) bool { return result.Workflows[i].ID < result.Workflows[j].ID })
	sort.Slice(result.Contracts, func(i, j int) bool { return result.Contracts[i].ID < result.Contracts[j].ID })
	sort.Slice(result.Symbols, func(i, j int) bool { return result.Symbols[i].ID < result.Symbols[j].ID })
	return result, nil
}

func inspectDecl(pkg *load.Package, file *ast.File, decl ast.Decl, out *model.Package) error {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if strings.HasPrefix(d.Name.Name, "Example") {
			out.Workflows = append(out.Workflows, workflow(pkg, d))
			return nil
		}
		if strings.HasPrefix(d.Name.Name, "Test") {
			contract, ok, err := parseContract(pkg, d)
			if err != nil {
				return err
			}
			if ok {
				out.Contracts = append(out.Contracts, contract)
			}
			return nil
		}
		if d.Name.IsExported() {
			out.Symbols = append(out.Symbols, funcSymbol(pkg, d))
		}
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			out.Symbols = append(out.Symbols, genSymbols(pkg, d, spec)...)
		}
	}
	return nil
}

func workflow(pkg *load.Package, d *ast.FuncDecl) model.Workflow {
	id := workflowID(d.Name.Name)
	return model.Workflow{Kind: "workflow", ID: id, Summary: clean(docWithoutOutput(d.Doc)), ExampleName: d.Name.Name, Code: nodeText(pkg, d.Body), RelatedSymbols: referencedNames(d.Body), Source: position(pkg, d.Pos())}
}

func parseContract(pkg *load.Package, d *ast.FuncDecl) (model.Contract, bool, error) {
	if d.Doc == nil {
		return model.Contract{}, false, nil
	}
	var id string
	var summary []string
	for _, c := range d.Doc.List {
		line := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
		if strings.HasPrefix(line, "codoc:contract") {
			fields := strings.Fields(line)
			if len(fields) != 2 {
				return model.Contract{}, false, fmt.Errorf("%s: malformed codoc:contract directive", position(pkg, c.Pos()).File)
			}
			id = fields[1]
		} else if line != "" {
			summary = append(summary, line)
		}
	}
	if id == "" {
		return model.Contract{}, false, nil
	}
	if !contractID.MatchString(id) {
		return model.Contract{}, false, fmt.Errorf("invalid contract ID %q", id)
	}
	if len(summary) == 0 {
		return model.Contract{}, false, fmt.Errorf("contract %q has no summary", id)
	}
	return model.Contract{Kind: "contract", ID: id, Summary: strings.Join(summary, "\n"), TestName: d.Name.Name, RelatedSymbols: referencedNames(d.Body), Source: position(pkg, d.Pos())}, true, nil
}

func funcSymbol(pkg *load.Package, d *ast.FuncDecl) model.Symbol {
	id := d.Name.Name
	if d.Recv != nil && len(d.Recv.List) > 0 {
		id = receiverName(d.Recv.List[0].Type) + "." + id
	}
	return model.Symbol{Kind: "symbol", ID: id, DeclarationKind: "func", Signature: funcSignature(pkg, d), Doc: docText(d.Doc), Source: position(pkg, d.Pos())}
}

func genSymbols(pkg *load.Package, decl *ast.GenDecl, spec ast.Spec) []model.Symbol {
	var result []model.Symbol
	switch s := spec.(type) {
	case *ast.TypeSpec:
		if s.Name.IsExported() {
			result = append(result, model.Symbol{Kind: "symbol", ID: s.Name.Name, DeclarationKind: "type", Signature: "type " + nodeText(pkg, s), Doc: firstDoc(s.Doc, decl.Doc), Source: position(pkg, s.Pos())})
		}
	case *ast.ValueSpec:
		kind := strings.ToLower(decl.Tok.String())
		for _, name := range s.Names {
			if name.IsExported() {
				result = append(result, model.Symbol{Kind: "symbol", ID: name.Name, DeclarationKind: kind, Signature: kind + " " + nodeText(pkg, s), Doc: firstDoc(s.Doc, decl.Doc), Source: position(pkg, s.Pos())})
			}
		}
	}
	return result
}

func validateContracts(items []model.Contract) error {
	seen := map[string]bool{}
	for _, x := range items {
		if seen[x.ID] {
			return fmt.Errorf("duplicate contract ID %q", x.ID)
		}
		seen[x.ID] = true
	}
	return nil
}
func link(pkg *model.Package) {
	symbols := map[string]*model.Symbol{}
	for i := range pkg.Symbols {
		symbols[pkg.Symbols[i].ID] = &pkg.Symbols[i]
		symbols[last(pkg.Symbols[i].ID)] = &pkg.Symbols[i]
	}
	for i := range pkg.Workflows {
		pkg.Workflows[i].RelatedSymbols = known(pkg.Workflows[i].RelatedSymbols, symbols)
		for _, n := range pkg.Workflows[i].RelatedSymbols {
			s := symbols[n]
			s.RelatedWorkflows = appendUnique(s.RelatedWorkflows, pkg.Workflows[i].ID)
		}
	}
	for i := range pkg.Contracts {
		pkg.Contracts[i].RelatedSymbols = known(pkg.Contracts[i].RelatedSymbols, symbols)
		for _, n := range pkg.Contracts[i].RelatedSymbols {
			s := symbols[n]
			s.RelatedContracts = appendUnique(s.RelatedContracts, pkg.Contracts[i].ID)
		}
	}
	for i := range pkg.Workflows {
		for _, c := range pkg.Contracts {
			if overlaps(pkg.Workflows[i].RelatedSymbols, c.RelatedSymbols) {
				pkg.Workflows[i].RelatedContracts = append(pkg.Workflows[i].RelatedContracts, c.ID)
			}
		}
	}
}
func known(names []string, symbols map[string]*model.Symbol) []string {
	var out []string
	for _, n := range names {
		if s := symbols[n]; s != nil {
			out = appendUnique(out, s.ID)
		}
	}
	sort.Strings(out)
	return out
}
func referencedNames(n ast.Node) []string {
	var out []string
	ast.Inspect(n, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			if x.Sel.IsExported() {
				out = appendUnique(out, x.Sel.Name)
			}
		case *ast.Ident:
			if x.IsExported() {
				out = appendUnique(out, x.Name)
			}
		}
		return true
	})
	return out
}
func workflowID(name string) string {
	s := strings.TrimPrefix(name, "Example")
	if s == "" {
		return "example"
	}
	parts := strings.Split(s, "_")
	if len(parts) > 1 {
		s = strings.Join(parts[1:], "-")
	}
	return kebab(s)
}
func kebab(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r == '_' {
			b.WriteByte('-')
			continue
		}
		if unicode.IsUpper(r) && i > 0 {
			b.WriteByte('-')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return strings.Trim(b.String(), "-")
}
func funcSignature(pkg *load.Package, d *ast.FuncDecl) string {
	copy := *d
	copy.Doc = nil
	copy.Body = nil
	return nodeText(pkg, &copy)
}

func receiverName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.StarExpr:
		return receiverName(x.X)
	case *ast.IndexExpr:
		return receiverName(x.X)
	case *ast.IndexListExpr:
		return receiverName(x.X)
	}
	return ""
}
func nodeText(pkg *load.Package, n any) string {
	var b bytes.Buffer
	_ = printer.Fprint(&b, pkg.FSet, n)
	return b.String()
}
func position(pkg *load.Package, p token.Pos) model.Position {
	x := pkg.FSet.Position(p)
	name, err := filepath.Rel(pkg.Dir, x.Filename)
	if err != nil {
		name = x.Filename
	}
	return model.Position{File: filepath.ToSlash(name), Line: x.Line}
}
func clean(s string) string { return strings.TrimSpace(s) }
func docText(d *ast.CommentGroup) string {
	if d == nil {
		return ""
	}
	return clean(d.Text())
}
func firstDoc(a, b *ast.CommentGroup) string {
	if a != nil {
		return docText(a)
	}
	return docText(b)
}
func docWithoutOutput(d *ast.CommentGroup) string {
	if d == nil {
		return ""
	}
	text := d.Text()
	if i := strings.Index(text, "Output:"); i >= 0 {
		text = text[:i]
	}
	return text
}
func appendUnique(xs []string, x string) []string {
	for _, v := range xs {
		if v == x {
			return xs
		}
	}
	return append(xs, x)
}
func last(s string) string {
	if i := strings.LastIndex(s, "."); i >= 0 {
		return s[i+1:]
	}
	return s
}
func overlaps(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if x == y {
				return true
			}
		}
	}
	return false
}
