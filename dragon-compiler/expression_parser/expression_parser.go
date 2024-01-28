package expression_parser

import (
	"fmt"
	"io"
	"lexer"
	"math"
	"strconv"
)

type Symbol struct {
	token lexer.Token
	value int
}

type ExpressionParser struct {
	parserLexer lexer.Lexer
	//用于存储一个算术表达式的所有标签
	symbols []Symbol
}

func NewExpressionParser(parserLexer lexer.Lexer) *ExpressionParser {
	return &ExpressionParser{
		parserLexer: parserLexer,
		symbols:     make([]Symbol, 0),
	}
}

func (e *ExpressionParser) makeSymbol(token lexer.Token) {
	val, err := strconv.Atoi(e.parserLexer.Lexeme)
	if err != nil {
		val = math.MaxInt
	}
	symbol := Symbol{
		token: token,
		value: val,
	}

	e.symbols = append(e.symbols, symbol)
}

func (e *ExpressionParser) getExprTokens() error {
	sawSemi := false
	for true {
		//读取算术表达式对应的标签，结束标志是遇到分号 s
		token, err := e.parserLexer.Scan()
		if err != nil && token.Tag != lexer.EOF {
			errStr := fmt.Sprintf("error: %v\n", err)
			panic(errStr)
		}

		if err == io.EOF {
			return err
		}

		e.makeSymbol(token)

		if token.Tag == lexer.SEMI {
			sawSemi = true
			break
		}
	}

	if sawSemi != true {
		//算术表达式没有 1️⃣ 分号结尾
		errStr := fmt.Sprintf("err: expression missing semi")
		panic(errStr)
	}

	return nil
}

func (e *ExpressionParser) Parse() {
	//第一个表达式左边是 stmt 所以从调用函数 stmt 开始
	e.stmt()
}

func (e *ExpressionParser) ioEnd() bool {
	e.symbols = make([]Symbol, 0)
	token, err := e.parserLexer.Scan()
	if err != nil && err != io.EOF {
		strErr := fmt.Sprintf("err: %v\n", err)
		panic(strErr)
	}
	if err == io.EOF {
		return true
	}

	e.makeSymbol(token)

	return false
}

func (e *ExpressionParser) stmt() {
	//stmt -> expr SEMI | expr SEMI stmt
	e.getExprTokens()
	val := e.expr(e.symbols[:len(e.symbols)-1])
	if e.symbols[len(e.symbols)-1].token.Tag != lexer.SEMI {
		panic("parsing error, expression not end with semi")
	}
	fmt.Printf("%d;", val)

	if e.ioEnd() {
		//所有标签读取完毕,这里采用 stmt -> expr SEMI
		return
	}
	//这里采用 stmt -> expr SEMI stmt
	e.stmt()
}

func (e *ExpressionParser) expr(symbols []Symbol) int {
	if len(symbols) == 0 || symbols == nil {
		panic("error token begin for expr parsing")
	}

	/*
		读取 PLUS 或 MINUS 标签，读取到，那么标签前面部分继续用 expr 分析
		后面部分用 term 分析
	*/
	sawOperator := false
	operatorPos := 0
	inPara := false
	for i := 0; i < len(symbols); i++ {
		/*
			在将 expr 通过+,-分割成两部分时，如果遇到左括号，那么在括号内部的
			+,-不作为分割的依据
		*/
		if symbols[i].token.Tag == lexer.LEFT_BRACKET {
			inPara = true
		}

		if symbols[i].token.Tag == lexer.RIGHT_BRACKET {
			if !inPara {
				panic("expr parsing err, missing left ")
			}
			inPara = false
		}
		if inPara {
			continue
		}

		if symbols[i].token.Tag == lexer.PLUS || symbols[i].token.Tag == lexer.MINUS {
			//必须找到表达式中最后一个加号或减号
			sawOperator = true
			operatorPos = i
		}
	}

	if sawOperator {
		//expr -> expr PLUS term | expr MINUS term
		left := e.expr(symbols[0:operatorPos])
		right := e.term(symbols[operatorPos+1:])
		res := 0
		if symbols[operatorPos].token.Tag == lexer.PLUS {
			res = left + right
		} else {
			res = left - right
		}

		return res
	} else {
		//expr -> term
		return e.term(symbols)
	}

	panic("expr parsing error: should not go here")
}

func (e *ExpressionParser) term(symbols []Symbol) int {
	if len(symbols) == 0 || symbols == nil {
		panic("error token begin for term parsing")
	}
	/*
		遍历标签,如果找到 MUL 或者 DIV，那么使用
		term -> term MUL factor | term DIV factor
		如果找不到使用
		term -> factor
	*/
	sawOperator := false
	operatorPos := 0
	inPara := false
	for i := 0; i < len(symbols); i++ {
		/*
			在将 expr 通过+,-分割成两部分时，如果遇到左括号，那么在括号内部的
			+,-不作为分割的依据
		*/
		if symbols[i].token.Tag == lexer.LEFT_BRACKET {
			inPara = true
		}

		if symbols[i].token.Tag == lexer.RIGHT_BRACKET {
			if !inPara {
				panic("expr parsing err, missing left ")
			}
			inPara = false
		}
		if inPara {
			continue
		}

		if symbols[i].token.Tag == lexer.MUL || symbols[i].token.Tag == lexer.DIV {
			//必须是表达式中最后一个乘号或除号
			sawOperator = true
			operatorPos = i
		}
	}

	if sawOperator {
		//term -> term MUL factor | term DIV factor
		left := e.term(symbols[0:operatorPos])
		right := e.factor(symbols[operatorPos+1:])
		if symbols[operatorPos].token.Tag == lexer.MUL {
			return left * right
		} else {
			return left / right
		}
	} else {
		return e.factor(symbols)
	}

	panic("term parsing err, should not go here")
}

func (e *ExpressionParser) factor(symbols []Symbol) int {
	if len(symbols) == 0 || symbols == nil {
		panic("error token begin for factor parsing")
	}

	sawLeftPara := false
	if symbols[0].token.Tag == lexer.LEFT_BRACKET {
		sawLeftPara = true
		symbols = symbols[1:]
	}

	sawRightPara := false
	if symbols[len(symbols)-1].token.Tag == lexer.RIGHT_BRACKET {
		sawRightPara = true
		symbols = symbols[:len(symbols)-1]
	}

	if sawLeftPara && !sawRightPara {
		panic("parsing factor err: missing right para")
	}

	if !sawLeftPara && sawRightPara {
		panic("parsing factor err: missing left para")
	}

	if sawLeftPara && sawRightPara {
		return e.expr(symbols)
	}

	//factor -> NUM
	if len(symbols) == 0 || len(symbols) > 1 {
		panic("factor->num but we have zero or more than 1 tokens")
	}

	if symbols[0].value == math.MaxInt {
		panic("parsing factor->num error: not a number")
	}
	return symbols[0].value
}
