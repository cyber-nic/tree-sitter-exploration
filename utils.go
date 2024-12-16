package main

import (
	"bytes"
	"strings"
)

func prettyPrintSexp(sexp string) string {
	var buffer bytes.Buffer
	indent := 0
	for _, char := range sexp {
		switch char {
		case '(':
			buffer.WriteString("\n" + strings.Repeat("  ", indent) + string(char))
			indent++
		case ')':
			indent--
			buffer.WriteString(string(char))
		case ' ':
			buffer.WriteString(string(char))
		default:
			buffer.WriteString(string(char))
		}
	}
	return buffer.String()
}
