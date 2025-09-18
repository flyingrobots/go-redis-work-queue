package requestidlint

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
)

// Analyzer ensures admin HTTP handlers emit X-Request-ID by funneling error responses through helpers.
var Analyzer = &analysis.Analyzer{
	Name: "requestidlint",
	Doc:  "reports error responses that bypass writeError and would miss X-Request-ID headers",
	Run:  run,
}

var allowedWriteHeaderFuncs = map[string]struct{}{
	"writeError": {},
	"writeJSON":  {},
}

func run(pass *analysis.Pass) (interface{}, error) {
	pkgPath := pass.Pkg.Path()
	if !strings.Contains(pkgPath, "internal/admin-api") && !strings.Contains(pkgPath, "internal/adminapi") {
		return nil, nil
	}

	for _, file := range pass.Files {
		filename := pass.Fset.File(file.Pos()).Name()
		if strings.HasSuffix(filename, "_test.go") {
			continue
		}

		var funcStack []string
		astutil.Apply(file, func(c *astutil.Cursor) bool {
			switch node := c.Node().(type) {
			case *ast.FuncDecl:
				funcStack = append(funcStack, node.Name.Name)
			case *ast.CallExpr:
				currentFunc := ""
				if len(funcStack) > 0 {
					currentFunc = funcStack[len(funcStack)-1]
				}
				inspectCall(pass, node, currentFunc)
			}
			return true
		}, func(c *astutil.Cursor) bool {
			if _, ok := c.Node().(*ast.FuncDecl); ok {
				if len(funcStack) > 0 {
					funcStack = funcStack[:len(funcStack)-1]
				}
			}
			return true
		})
	}

	return nil, nil
}

func inspectCall(pass *analysis.Pass, call *ast.CallExpr, currentFunc string) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	// Handle package selector (e.g., http.Error)
	if obj := pass.TypesInfo.Uses[sel.Sel]; obj != nil {
		if pkg := obj.Pkg(); pkg != nil && pkg.Path() == "net/http" && obj.Name() == "Error" {
			pass.Reportf(sel.Sel.Pos(), "use writeError helper to ensure X-Request-ID header is set instead of http.Error")
			return
		}
	}

	if sel.Sel.Name != "WriteHeader" {
		return
	}

	if !isHTTPResponseWriter(pass.TypesInfo.TypeOf(sel.X)) {
		return
	}

	if len(call.Args) != 1 {
		return
	}

	info, ok := pass.TypesInfo.Types[call.Args[0]]
	if !ok || info.Value == nil {
		return
	}

	value := info.Value
	if value.Kind() != constant.Int {
		return
	}

	if v, ok := constant.Int64Val(value); ok && v >= 400 {
		if _, allowed := allowedWriteHeaderFuncs[currentFunc]; !allowed {
			pass.Reportf(sel.Sel.Pos(), "use writeError helper to ensure X-Request-ID header is set instead of calling WriteHeader directly")
		}
	}
}

func isHTTPResponseWriter(t types.Type) bool {
	if t == nil {
		return false
	}

	if named, ok := t.(*types.Named); ok {
		if obj := named.Obj(); obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == "net/http" && obj.Name() == "ResponseWriter" {
			return true
		}
	}

	if iface, ok := t.(*types.Interface); ok {
		// Direct interface type imported with type alias (rare)
		if iface.NumMethods() == 3 {
			hasHeader := false
			hasWrite := false
			hasWriteHeader := false
			for i := 0; i < iface.NumMethods(); i++ {
				m := iface.Method(i)
				switch m.Name() {
				case "Header":
					hasHeader = true
				case "Write":
					hasWrite = true
				case "WriteHeader":
					hasWriteHeader = true
				}
			}
			if hasHeader && hasWrite && hasWriteHeader {
				return true
			}
		}
	}

	if pointer, ok := t.(*types.Pointer); ok {
		return isHTTPResponseWriter(pointer.Elem())
	}

	return false
}
