module dragon-compiler

go 1.17

require (
	augmented_parser v0.0.0-00010101000000-000000000000
	lexer v0.0.0-00010101000000-000000000000
)

replace lexer => ./lexer

replace simple_parser => ./parser

replace expression_parser => ./expression_parser

replace pda => ./pda

replace augmented_parser => ./augmented_parser
