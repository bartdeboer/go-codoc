// Package index turns Go's documentation model and syntax into compact records.
package index

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/doc/comment"
	"go/printer"
	"go/token"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bartdeboer/go-codoc/internal/load"
	"github.com/bartdeboer/go-codoc/internal/model"
)

var contractID = regexp.MustCompile(`^[a-z0-9]+(?:[/-][a-z0-9]+)*$`)

func Build(pkg *load.Package) (model.Package, error) {
	out := model.Package{Kind: "package", ImportPath: pkg.ImportPath, Name: pkg.Name, Workflows: []model.Workflow{}, Contracts: []model.Contract{}, Symbols: []model.Symbol{}, DocumentedTests: []model.DocumentedTest{}, Gaps: []string{}}
	for _, file := range pkg.Files {
		if file.Name.Name == pkg.Name && out.Overview == "" && file.Doc != nil {
			out.Overview = clean(file.Doc.Text())
		}
	}
	out.Workflows = examples(pkg)
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			if err := inspectDecl(pkg, file, decl, &out); err != nil {
				return model.Package{}, err
			}
		}
	}
	if err := validateContracts(out.Contracts); err != nil {
		return model.Package{}, err
	}
	if err := validateDocumentedTests(out.DocumentedTests); err != nil {
		return model.Package{}, err
	}
	link(&out)
	sortRecords(&out)
	out.Gaps = documentationGaps(out)
	return out, nil
}

func examples(pkg *load.Package) []model.Workflow {
	var out []model.Workflow
	for _, ex := range doc.Examples(pkg.Files...) {
		if ex.Output == "" && !ex.EmptyOutput {
			continue
		}
		base, suffix := ex.Name, ex.Suffix
		if suffix == "" {
			if i := strings.LastIndexByte(base, '_'); i >= 0 && i+1 < len(base) && unicode.IsLower(rune(base[i+1])) {
				suffix, base = base[i+1:], base[:i]
			}
		}
		primary := strings.ReplaceAll(base, "_", ".")
		idSource := suffix
		if idSource == "" {
			idSource = base
			if i := strings.LastIndex(idSource, "_"); i >= 0 {
				idSource = idSource[i+1:]
			}
		}
		name := "Example" + ex.Name
		pos := findFunction(pkg, name)
		out = append(out, model.Workflow{Kind: "workflow", ID: kebab(idSource), Summary: clean(ex.Doc), ExampleName: name, PrimarySymbol: primary, Code: nodeText(pkg, ex.Code), ExpectedOutput: ex.Output, EmptyOutput: ex.EmptyOutput, RelatedSymbols: []string{}, RelatedContracts: []string{}, Verification: "not run", Source: pos})
	}
	return out
}

func inspectDecl(pkg *load.Package, file *ast.File, decl ast.Decl, out *model.Package) error {
	d, ok := decl.(*ast.FuncDecl)
	if ok && strings.HasSuffix(pkg.FSet.Position(d.Pos()).Filename, "_test.go") {
		annotated, omitCode, err := documentedDirectives(d.Doc)
		if err != nil {
			return fmt.Errorf("%s: %w", position(pkg, d.Pos()).File, err)
		}
		if annotated {
			if !isGoTest(file, d) {
				return fmt.Errorf("%s is marked codoc:doc but is not a valid Go test", d.Name.Name)
			}
			path := documentedTest(pkg, file, d, omitCode, out.Symbols)
			if path.Summary == "" {
				return fmt.Errorf("documented test %s requires a prose comment", d.Name.Name)
			}
			out.DocumentedTests = append(out.DocumentedTests, path)
			return nil
		}
		if strings.HasPrefix(d.Name.Name, "Test") {
			contract, found, err := parseContract(pkg, d)
			if err != nil {
				return err
			}
			if found {
				out.Contracts = append(out.Contracts, contract)
			}
		}
		return nil
	}
	if strings.HasSuffix(pkg.FSet.Position(decl.Pos()).Filename, "_test.go") {
		return nil
	}
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if d.Name.IsExported() && (d.Recv == nil || ast.IsExported(receiverName(d.Recv.List[0].Type))) {
			out.Symbols = append(out.Symbols, funcSymbol(pkg, d))
		}
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			out.Symbols = append(out.Symbols, genSymbols(pkg, d, spec)...)
		}
	}
	return nil
}

func documentedDirectives(doc *ast.CommentGroup) (annotated, omitCode bool, err error) {
	if doc == nil {
		return false, false, nil
	}
	for _, comment := range doc.List {
		if !strings.HasPrefix(comment.Text, "//codoc:") {
			continue
		}
		directive := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		switch directive {
		case "codoc:doc":
			annotated = true
		case "codoc:code omit":
			omitCode = true
		default:
			if strings.HasPrefix(directive, "codoc:contract ") {
				continue
			}
			return false, false, fmt.Errorf("unsupported directive %q", directive)
		}
	}
	if omitCode && !annotated {
		return false, false, fmt.Errorf("codoc:code omit requires codoc:doc")
	}
	return annotated, omitCode, nil
}

func isGoTest(file *ast.File, d *ast.FuncDecl) bool {
	if d.Recv != nil || !strings.HasPrefix(d.Name.Name, "Test") || len(d.Name.Name) == len("Test") {
		return false
	}
	first, _ := utf8.DecodeRuneInString(strings.TrimPrefix(d.Name.Name, "Test"))
	if unicode.IsLower(first) {
		return false
	}
	if d.Type.TypeParams != nil || d.Type.Results != nil || d.Type.Params == nil || len(d.Type.Params.List) != 1 {
		return false
	}
	parameter := d.Type.Params.List[0]
	if len(parameter.Names) != 1 {
		return false
	}
	pointer, ok := parameter.Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	qualifiers, dotImport := testingImports(file)
	if identifier, ok := pointer.X.(*ast.Ident); ok {
		return dotImport && identifier.Name == "T"
	}
	selector, ok := pointer.X.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "T" {
		return false
	}
	qualifier, ok := selector.X.(*ast.Ident)
	return ok && qualifiers[qualifier.Name]
}

func testingImports(file *ast.File) (map[string]bool, bool) {
	qualifiers := map[string]bool{}
	dotImport := false
	for _, spec := range file.Imports {
		if strings.Trim(spec.Path.Value, "\"") != "testing" {
			continue
		}
		if spec.Name == nil {
			qualifiers["testing"] = true
			continue
		}
		switch spec.Name.Name {
		case ".":
			dotImport = true
		case "_":
		default:
			qualifiers[spec.Name.Name] = true
		}
	}
	return qualifiers, dotImport
}

func documentedTest(pkg *load.Package, file *ast.File, d *ast.FuncDecl, omitCode bool, symbols []model.Symbol) model.DocumentedTest {
	name := strings.TrimPrefix(d.Name.Name, "Test")
	code := ""
	if !omitCode {
		code = nodeText(pkg, d.Body)
	}
	return model.DocumentedTest{Kind: "documented_test", ID: kebab(name), Title: title(name), Summary: docText(d.Doc), TestName: d.Name.Name, Code: code, CodeOmitted: omitCode, Source: position(pkg, d.Pos()), Status: "not run", RelatedSymbols: relatedDocSymbols(pkg, file, d.Doc, symbols)}
}

func relatedDocSymbols(pkg *load.Package, file *ast.File, group *ast.CommentGroup, symbols []model.Symbol) []model.RelatedSymbol {
	if group == nil {
		return []model.RelatedSymbol{}
	}
	available := make(map[string]model.Symbol, len(symbols))
	for _, symbol := range symbols {
		available[symbol.ID] = symbol
	}
	aliases := currentPackageAliases(pkg, file)
	parser := comment.Parser{
		LookupPackage: func(name string) (string, bool) {
			if aliases[name] {
				return pkg.ImportPath, true
			}
			return "", false
		},
		LookupSym: func(recv, name string) bool {
			if file.Name.Name != pkg.Name {
				return false
			}
			id := name
			if recv != "" {
				id = recv + "." + name
			}
			_, ok := available[id]
			return ok
		},
	}
	parsed := parser.Parse(group.Text())
	var links []*comment.DocLink
	walkCommentBlocks(parsed.Content, func(link *comment.DocLink) { links = append(links, link) })
	result := []model.RelatedSymbol{}
	seen := map[string]bool{}
	for _, link := range links {
		if link.Name == "" || link.ImportPath != "" && link.ImportPath != pkg.ImportPath {
			continue
		}
		id := link.Name
		if link.Recv != "" {
			id = link.Recv + "." + link.Name
		}
		symbol, ok := available[id]
		if !ok || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, model.RelatedSymbol{Package: pkg.ImportPath, Symbol: id, Source: symbol.Source})
	}
	return result
}

func currentPackageAliases(pkg *load.Package, file *ast.File) map[string]bool {
	aliases := map[string]bool{}
	if file.Name.Name == pkg.Name {
		aliases[pkg.Name] = true
	}
	for _, spec := range file.Imports {
		if strings.Trim(spec.Path.Value, "\"") != pkg.ImportPath {
			continue
		}
		if spec.Name == nil {
			aliases[pkg.Name] = true
			continue
		}
		if spec.Name.Name != "_" && spec.Name.Name != "." {
			aliases[spec.Name.Name] = true
		}
	}
	return aliases
}

func walkCommentBlocks(blocks []comment.Block, visit func(*comment.DocLink)) {
	for _, block := range blocks {
		switch value := block.(type) {
		case *comment.Paragraph:
			walkCommentText(value.Text, visit)
		case *comment.Heading:
			walkCommentText(value.Text, visit)
		case *comment.List:
			for _, item := range value.Items {
				walkCommentBlocks(item.Content, visit)
			}
		}
	}
}
func walkCommentText(text []comment.Text, visit func(*comment.DocLink)) {
	for _, item := range text {
		if link, ok := item.(*comment.DocLink); ok {
			visit(link)
		}
	}
}

func title(name string) string {
	var words []string
	start := 0
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) {
			words = append(words, name[start:i])
			start = i
		}
	}
	if start < len(name) {
		words = append(words, name[start:])
	}
	return strings.Join(words, " ")
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
	return model.Contract{Kind: "contract", ID: id, Summary: strings.Join(summary, "\n"), TestName: d.Name.Name, RelatedSymbols: []string{}, Verification: "not run", Source: position(pkg, d.Pos())}, true, nil
}

func funcSymbol(pkg *load.Package, d *ast.FuncDecl) model.Symbol {
	id := d.Name.Name
	if d.Recv != nil {
		id = receiverName(d.Recv.List[0].Type) + "." + id
	}
	return model.Symbol{Kind: "symbol", ID: id, DeclarationKind: "func", Signature: funcSignature(pkg, d), Doc: docText(d.Doc), RelatedWorkflows: []string{}, RelatedContracts: []string{}, Source: position(pkg, d.Pos())}
}
func genSymbols(pkg *load.Package, decl *ast.GenDecl, spec ast.Spec) []model.Symbol {
	var out []model.Symbol
	switch s := spec.(type) {
	case *ast.TypeSpec:
		if s.Name.IsExported() {
			out = append(out, model.Symbol{Kind: "symbol", ID: s.Name.Name, DeclarationKind: "type", Signature: "type " + nodeText(pkg, s), Doc: firstDoc(s.Doc, decl.Doc), RelatedWorkflows: []string{}, RelatedContracts: []string{}, Source: position(pkg, s.Pos())})
		}
	case *ast.ValueSpec:
		kind := strings.ToLower(decl.Tok.String())
		for _, name := range s.Names {
			if name.IsExported() {
				out = append(out, model.Symbol{Kind: "symbol", ID: name.Name, DeclarationKind: kind, Signature: kind + " " + nodeText(pkg, s), Doc: firstDoc(s.Doc, decl.Doc), RelatedWorkflows: []string{}, RelatedContracts: []string{}, Source: position(pkg, s.Pos())})
			}
		}
	}
	return out
}

func link(pkg *model.Package) {
	symbols := map[string]*model.Symbol{}
	for i := range pkg.Symbols {
		symbols[pkg.Symbols[i].ID] = &pkg.Symbols[i]
	}
	for i := range pkg.Workflows {
		workflow := &pkg.Workflows[i]
		primary := symbols[workflow.PrimarySymbol]
		if primary == nil {
			workflow.RelatedSymbols = []string{}
			continue
		}
		workflow.RelatedSymbols = []string{primary.ID}
		primary.RelatedWorkflows = appendUnique(primary.RelatedWorkflows, workflow.ID)
	}
}

func documentationGaps(p model.Package) []string {
	var gaps []string
	if p.Overview == "" {
		gaps = append(gaps, "package overview missing")
	}
	undocumented := 0
	for _, symbol := range p.OrientationSymbols(len(p.Symbols)) {
		if symbol.Doc == "" {
			undocumented++
		}
	}
	if undocumented > 0 {
		gaps = append(gaps, fmt.Sprintf("%d public entry points undocumented", undocumented))
	}
	return gaps
}
func sortRecords(p *model.Package) {
	sort.Slice(p.Workflows, func(i, j int) bool { return p.Workflows[i].ID < p.Workflows[j].ID })
	sort.Slice(p.Contracts, func(i, j int) bool { return p.Contracts[i].ID < p.Contracts[j].ID })
	sort.Slice(p.Symbols, func(i, j int) bool { return p.Symbols[i].ID < p.Symbols[j].ID })
}
func validateDocumentedTests(paths []model.DocumentedTest) error {
	ids, names := map[string]bool{}, map[string]bool{}
	for _, path := range paths {
		if ids[path.ID] {
			return fmt.Errorf("duplicate documented test ID %q", path.ID)
		}
		if names[path.TestName] {
			return fmt.Errorf("duplicate documented test %q", path.TestName)
		}
		ids[path.ID], names[path.TestName] = true, true
	}
	return nil
}

func validateContracts(xs []model.Contract) error {
	seen := map[string]bool{}
	for _, x := range xs {
		if seen[x.ID] {
			return fmt.Errorf("duplicate contract ID %q", x.ID)
		}
		seen[x.ID] = true
	}
	return nil
}
func findFunction(pkg *load.Package, name string) model.Position {
	for _, f := range pkg.Files {
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == name {
				return position(pkg, fn.Pos())
			}
		}
	}
	return model.Position{}
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
func kebab(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r == '_' || r == '/' {
			b.WriteByte('-')
			continue
		}
		if unicode.IsUpper(r) && i > 0 {
			b.WriteByte('-')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	id := strings.Trim(b.String(), "-")
	for strings.Contains(id, "--") {
		id = strings.ReplaceAll(id, "--", "-")
	}
	return id
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
func appendUnique(xs []string, x string) []string {
	for _, v := range xs {
		if v == x {
			return xs
		}
	}
	return append(xs, x)
}
func sorted(xs []string) []string {
	sort.Strings(xs)
	if xs == nil {
		return []string{}
	}
	return xs
}
