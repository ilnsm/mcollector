package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// DirectExitAnalyzer is an analyzer that checks for explicit exit calls in main func in main package.
// It is defined as a variable of type *analysis.Analyzer.
var DirectExitAnalyzer = &analysis.Analyzer{
	Name: "directexit",
	Doc:  "check for explicit exit calls",
	Run:  run,
}

// run is the function that performs the analysis. It takes a *analysis.Pass as an argument,
// which contains information about the package being analyzed. The function iterates over all
// the files in the package and checks if the file name is "main". If it is, it inspects all
// the nodes of the Abstract Syntax Tree (AST) of the file. It specifically looks for function
// declarations with the name "main" and checks each statement in the function body. If a
// statement is an expression statement, it checks if the expression represents a function call
// with an explicit exit.
func run(pass *analysis.Pass) (interface{}, error) {
	expr := func(x *ast.ExprStmt) {
		// check that the expression represents a function call
		// with an explicit exit
		if call, ok := x.X.(*ast.CallExpr); ok {
			if isExplitExit(call) {
				pass.Reportf(x.Pos(), "direct exit call")
			}
		}
	}
	for _, file := range pass.Files {
		if file.Name.String() != "main" {
			continue
		}
		// using the ast.Inspect function, we go through all the nodes of the AST
		ast.Inspect(file, func(node ast.Node) bool {
			if f, ok := node.(*ast.FuncDecl); ok {
				if f.Name.Name == "main" {
					for _, stmt := range f.Body.List {
						if x, ok := stmt.(*ast.ExprStmt); ok {
							expr(x)
						}
					}
				}
			}
			return true
		})
	}
	//nolint // we don't need to return anything
	return nil, nil
}

// isExplitExit is a helper function that checks if a function call is an explicit exit.
// It takes an *ast.CallExpr as an argument, which represents a function call in the AST.
// It checks if the function being called is named "exit". If it is, the function returns true;
// otherwise, it returns false.
func isExplitExit(call *ast.CallExpr) bool {
	// check that the function call is exit
	if pkg, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := pkg.X.(*ast.Ident); ok {
			if ident.Name == "os" && pkg.Sel.Name == "Exit" {
				return true
			}
		}
	}
	return false
}
