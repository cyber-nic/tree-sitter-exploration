package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	// sitter "github.com/smacker/go-tree-sitter"
	sitter "github.com/tree-sitter/go-tree-sitter"

	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_typescript "github.com/tree-sitter/tree-sitter-typescript/bindings/go"
)

func getLanguageByExtension(ext string) *sitter.Language {
	switch ext {
	case ".go":
		log.Trace().Msg("file language: go")
		return sitter.NewLanguage(tree_sitter_go.Language())
	case ".js":
		log.Trace().Msg("file language: javascript")
		return sitter.NewLanguage(tree_sitter_javascript.Language())
	case ".py":
		log.Trace().Msg("file language: python")
		return sitter.NewLanguage(tree_sitter_python.Language())
	case ".ts":
		log.Trace().Msg("file language: typescript")
		return sitter.NewLanguage(tree_sitter_typescript.LanguageTypescript())
	default:
		return nil
	}
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// Ensure a file path is provided as a command-line argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <file-path>")
		os.Exit(1)
	}

	filePath := os.Args[1]

	// Get the file extension
	ext := filepath.Ext(filePath)

	// Get the appropriate language for the file extension
	language := getLanguageByExtension(ext)
	if language == nil {
		fmt.Printf("Unsupported file extension: %s\n", ext)
		os.Exit(1)
	}

	// Initialize the parser
	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(language)

	// Read the file contents
	code, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		os.Exit(1)
	}

	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	tree := parser.Parse(code, nil)
	defer tree.Close()

	root := tree.RootNode()
	runtime.ReadMemStats(&memAfter)
	cstBytes := memAfter.HeapAlloc - memBefore.HeapAlloc

	sexp := root.ToSexp()
	sexpBytes := len(sexp)
	codeBytes := float64(len(code))

	// Compute mem increase from code to tree
	cstIncrease := (float64(cstBytes) - codeBytes) / codeBytes * 100
	sexpIncrease := (float64(sexpBytes) - codeBytes) / codeBytes * 100

	printTreeWithNamesRecursive(root, code, 0)
	log.Info().
		Int("code", int(codeBytes)).
		Str("tree", fmt.Sprintf("%d (+%.f%%)", cstBytes, cstIncrease)).
		Str("sexp", fmt.Sprintf("%d (+%.f%%)", sexpBytes, sexpIncrease)).
		// Msg(filePath)
		Msg("mem")
}

// structural view of the CST but omits the textual content
func printTreeWithNamesRecursive(node *sitter.Node, code []byte, level int) {
	if node == nil {
		return
	}

	// Indentation for nested levels
	indent := strings.Repeat("  ", level)

	// Extract and print relevant information
	switch node.Kind() {
	case "function_declaration":
		// Extract function name
		if nameNode := node.ChildByFieldName("name"); nameNode != nil {
			fmt.Printf("%sFunction: %s\n", indent, nameNode.Utf8Text(code))
		}
	case "variable_declaration", "short_var_declaration":
		// Extract variable names
		if leftNode := node.ChildByFieldName("left"); leftNode != nil {
			for i := uint(0); i < leftNode.ChildCount(); i++ {
				child := leftNode.Child(i)
				if child.Kind() == "identifier" {
					fmt.Printf("%sVariable: %s\n", indent, child.Utf8Text(code))
				}
			}
		}
	case "identifier":
		// Print identifiers in general context (e.g., parameters or expressions)
		fmt.Printf("%sIdentifier: %s\n", indent, node.Utf8Text(code))
	}

	// Recursively process child nodes
	for i := uint(0); i < node.ChildCount(); i++ {
		printTreeWithNamesRecursive(node.Child(i), code, level+1)
	}
}

// func printTreeWithNamesIterative(node *sitter.Node, code []byte) {
// 	if node == nil {
// 		return
// 	}

// 	// Stack to hold nodes for iterative traversal
// 	type stackEntry struct {
// 		node  *sitter.Node
// 		level int
// 	}
// 	stack := []stackEntry{{node, 0}}

// 	// Iteratively process the CST
// 	for len(stack) > 0 {
// 		// Pop the top node from the stack
// 		entry := stack[len(stack)-1]
// 		stack = stack[:len(stack)-1]

// 		currentNode := entry.node
// 		level := entry.level

// 		// Indentation for nested levels
// 		indent := strings.Repeat("  ", level)

// 		// Extract and print relevant information
// 		switch currentNode.Type() {
// 		case "function_declaration":
// 			// Extract function name
// 			if nameNode := currentNode.ChildByFieldName("name"); nameNode != nil {
// 				fmt.Printf("%sFunction: %s\n", indent, nameNode.Content(code))
// 			}
// 		case "variable_declaration", "short_var_declaration":
// 			// Extract variable names
// 			if leftNode := currentNode.ChildByFieldName("left"); leftNode != nil {
// 				for i := 0; i < int(leftNode.ChildCount()); i++ {
// 					child := leftNode.Child(i)
// 					if child.Type() == "identifier" {
// 						fmt.Printf("%sVariable: %s\n", indent, child.Content(code))
// 					}
// 				}
// 			}
// 		case "identifier":
// 			// Print identifiers in general context (e.g., parameters or expressions)
// 			fmt.Printf("%sIdentifier: %s\n", indent, currentNode.Content(code))
// 		}

// 		// Push child nodes onto the stack
// 		for i := int(currentNode.ChildCount()) - 1; i >= 0; i-- {
// 			stack = append(stack, stackEntry{
// 				node:  currentNode.Child(i),
// 				level: level + 1,
// 			})
// 		}
// 	}
// }
