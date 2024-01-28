package main

import (
	"augmented_parser"
	"lexer"
)

func main() {
	exprLexer := lexer.NewLexer("1+2*(4+3);")
	augmentedParser := augmented_parser.NewAugmentedParser(exprLexer)
	augmentedParser.Parse()
}
