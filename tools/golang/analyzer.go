package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

type Node struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Args       []string `json:"args"`
	ReturnType []string `json:"returns"`
	Endpoint   string   `json:"endpoint,omitempty"` // Add endpoint field
}

type Edge struct {
	Source     string `json:"source"`
	Target     string `json:"target"`
	LineNumber int    `json:"line"`
	FilePath   string `json:"file"`
	Endpoint   string `json:"endpoint"` // Add endpoint field
}

type CallGraph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

var anonFuncCounter int

func main() {
	if len(os.Args) < 2 {
		panic("Please provide at least one file path")
	}

	callGraph := &CallGraph{
		Nodes: []Node{},
		Edges: []Edge{},
	}

	// Track unique nodes
	nodeMap := make(map[string]bool)

	// Track current function
	var currentFunc string

	// Iterate over all provided file paths
	for _, filePath := range os.Args[1:] {
		// Skip test files
		if strings.Contains(filePath, "_test") {
			continue
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			panic(err)
		}

		// First pass: collect all function declarations
		ast.Inspect(node, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				name := funcDecl.Name.Name
				currentFunc = name

				// Collect arguments
				args := []string{}
				if funcDecl.Type.Params != nil {
					for _, param := range funcDecl.Type.Params.List {
						// Get parameter type as string
						paramType := typeToString(param.Type)
						// Handle multiple names for the same type
						for _, name := range param.Names {
							args = append(args, name.Name+" "+paramType)
						}
					}
				}

				// Collect return types
				returns := []string{}
				if funcDecl.Type.Results != nil {
					for _, result := range funcDecl.Type.Results.List {
						returns = append(returns, typeToString(result.Type))
					}
				}

				if !nodeMap[name] {
					callGraph.Nodes = append(callGraph.Nodes, Node{
						ID:         name,
						Name:       name,
						Args:       args,
						ReturnType: returns,
					})
					nodeMap[name] = true
				}
			}
			return true
		})

		// Second pass: collect all function calls and route registrations
		ast.Inspect(node, func(n ast.Node) bool {
			if funcDecl, ok := n.(*ast.FuncDecl); ok {
				currentFunc = funcDecl.Name.Name
			}
			if callExpr, ok := n.(*ast.CallExpr); ok {
				pos := fset.Position(callExpr.Pos())

				// Collect arguments for the function call
				callArgs := []string{}
				for _, arg := range callExpr.Args {
					callArgs = append(callArgs, exprToString(arg))
				}

				switch fun := callExpr.Fun.(type) {
				case *ast.Ident:
					target := fun.Name
					if !nodeMap[target] {
						callGraph.Nodes = append(callGraph.Nodes, Node{
							ID:         target,
							Name:       target,
							Args:       callArgs,
							ReturnType: []string{},
						})
						nodeMap[target] = true
					}
					callGraph.Edges = append(callGraph.Edges, Edge{
						Source:     currentFunc,
						Target:     target,
						LineNumber: pos.Line,
						FilePath:   pos.Filename,
					})
				case *ast.SelectorExpr:
					target := fun.Sel.Name
					if !nodeMap[target] {
						callGraph.Nodes = append(callGraph.Nodes, Node{
							ID:         target,
							Name:       target,
							Args:       callArgs,
							ReturnType: []string{},
						})
						nodeMap[target] = true
					}
					callGraph.Edges = append(callGraph.Edges, Edge{
						Source:     currentFunc,
						Target:     target,
						LineNumber: pos.Line,
						FilePath:   pos.Filename,
					})

					// Detect route registrations
					method := strings.ToLower(target)
					if method == "get" || method == "post" || method == "put" || method == "delete" || method == "patch" {
						if len(callExpr.Args) > 1 {
							endpoint := exprToString(callExpr.Args[0])
							if funcLit, ok := callExpr.Args[1].(*ast.SelectorExpr); ok {
								target := funcLit.Sel.Name

								// Create a node for the endpoint
								if !nodeMap[endpoint] {
									callGraph.Nodes = append(callGraph.Nodes, Node{
										ID:         endpoint,
										Name:       endpoint,
										Args:       []string{},
										ReturnType: []string{},
										Endpoint:   endpoint, // Set endpoint
									})
									nodeMap[endpoint] = true
								}

								// Create a node for the target function if it doesn't exist
								if !nodeMap[target] {
									callGraph.Nodes = append(callGraph.Nodes, Node{
										ID:         target,
										Name:       target,
										Args:       []string{},
										ReturnType: []string{},
									})
									nodeMap[target] = true
								}

								// Add edge from endpoint to target function
								callGraph.Edges = append(callGraph.Edges, Edge{
									Source:     endpoint,
									Target:     target,
									LineNumber: pos.Line,
									FilePath:   pos.Filename,
									Endpoint:   endpoint, // Set endpoint in edge
								})
							} else if funcLit, ok := callExpr.Args[1].(*ast.FuncLit); ok {
								// Handle anonymous functions
								anonFuncName := "AnonFunc" + strconv.Itoa(anonFuncCounter)
								anonFuncCounter++
								pos := fset.Position(callExpr.Pos())

								// Create a node for the endpoint
								if !nodeMap[endpoint] {
									callGraph.Nodes = append(callGraph.Nodes, Node{
										ID:         endpoint,
										Name:       endpoint,
										Args:       []string{},
										ReturnType: []string{},
										Endpoint:   endpoint, // Set endpoint
									})
									nodeMap[endpoint] = true
								}

								// Add edge from endpoint to anonymous function
								callGraph.Edges = append(callGraph.Edges, Edge{
									Source:     endpoint,
									Target:     anonFuncName,
									LineNumber: pos.Line,
									FilePath:   pos.Filename,
									Endpoint:   endpoint, // Set endpoint in edge
								})

								// Create a node for the anonymous function
								callGraph.Nodes = append(callGraph.Nodes, Node{
									ID:         anonFuncName,
									Name:       anonFuncName,
									Args:       []string{},
									ReturnType: []string{},
								})

								// Inspect the body of the anonymous function to find internal calls
								ast.Inspect(funcLit.Body, func(n ast.Node) bool {
									if callExpr, ok := n.(*ast.CallExpr); ok {
										pos := fset.Position(callExpr.Pos())
										if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
											target := selExpr.Sel.Name

											// Create a node for the target function if it doesn't exist
											if !nodeMap[target] {
												callGraph.Nodes = append(callGraph.Nodes, Node{
													ID:         target,
													Name:       target,
													Args:       []string{},
													ReturnType: []string{},
												})
												nodeMap[target] = true
											}

											// Add edge from anonymous function to target function
											callGraph.Edges = append(callGraph.Edges, Edge{
												Source:     anonFuncName,
												Target:     target,
												LineNumber: pos.Line,
												FilePath:   pos.Filename,
											})
										}
									}
									return true
								})
							}
						}
					}
				}
			}
			return true
		})
	}

	// Ensure all nodes referenced in edges are present in the nodes list
	for _, edge := range callGraph.Edges {
		if !nodeMap[edge.Source] {
			callGraph.Nodes = append(callGraph.Nodes, Node{
				ID:         edge.Source,
				Name:       edge.Source,
				Args:       []string{},
				ReturnType: []string{},
			})
			nodeMap[edge.Source] = true
		}
		if !nodeMap[edge.Target] {
			callGraph.Nodes = append(callGraph.Nodes, Node{
				ID:         edge.Target,
				Name:       edge.Target,
				Args:       []string{},
				ReturnType: []string{},
			})
			nodeMap[edge.Target] = true
		}
	}

	// Convert call graph to JSON
	jsonData, err := json.Marshal(callGraph)
	if err != nil {
		panic(err)
	}

	// Print JSON data
	os.Stdout.Write(jsonData)
}

// Convert AST expression to string representation
func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	default:
		return "unknown"
	}
}

// Convert AST expression to string representation for function arguments
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.BasicLit:
		return t.Value
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.CallExpr:
		return exprToString(t.Fun) + "()"
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.UnaryExpr:
		return t.Op.String() + exprToString(t.X)
	case *ast.BinaryExpr:
		return exprToString(t.X) + " " + t.Op.String() + " " + exprToString(t.Y)
	case *ast.ParenExpr:
		return "(" + exprToString(t.X) + ")"
	case *ast.IndexExpr:
		return exprToString(t.X) + "[" + exprToString(t.Index) + "]"
	case *ast.SliceExpr:
		return exprToString(t.X) + "[" + exprToString(t.Low) + ":" + exprToString(t.High) + "]"
	case *ast.CompositeLit:
		if ident, ok := t.Type.(*ast.Ident); ok {
			return ident.Name + "{}"
		}
		return "composite{}"
	default:
		return "unknown"
	}
}
