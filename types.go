package main

type ASTNode struct {
	Type     string    `json:"type"`
	Children []ASTNode `json:"children,omitempty"`
}
