package augmented_parser

import (
	"fmt"
	"lexer"
)

type AugmentedParser struct {
	parserLexer lexer.Lexer
	//用于存储虚拟寄存器的名字
	registerNames []string
	//存储当前已分配寄存器的名字
	regiserStack []string
	//当前可用寄存器名字的下标
	registerNameIdx int
	//存储读取后又放回去的 token
	reverseToken []lexer.Token
}

func NewAugmentedParser(parserLexer lexer.Lexer) *AugmentedParser {
	return &AugmentedParser{
		parserLexer:     parserLexer,
		registerNames:   []string{"t0", "t1", "t2", "t3", "t4", "t5", "t6", "t7"},
		regiserStack:    make([]string, 0),
		registerNameIdx: 0,
		reverseToken:    make([]lexer.Token, 0),
	}
}

func (a *AugmentedParser) putbackToken(token lexer.Token) {
	a.reverseToken = append(a.reverseToken, token)
}

func (a *AugmentedParser) newName() string {
	//返回一个寄存器的名字
	if a.registerNameIdx >= len(a.registerNames) {
		//没有寄存器可用
		panic("register name running out")
	}
	name := a.registerNames[a.registerNameIdx]
	a.registerNameIdx += 1
	return name
}

func (a *AugmentedParser) freeName(name string) {
	//释放当前寄存器名字
	if a.registerNameIdx > len(a.registerNames) {
		panic("register name index out of bound")
	}

	if a.registerNameIdx == 0 {
		panic("register name is full")
	}

	a.registerNameIdx -= 1
	a.registerNames[a.registerNameIdx] = name
}

func (a *AugmentedParser) createTmp(str string) {
	//创建一条寄存器赋值指令并
	name := a.newName()
	//生成一条寄存器赋值指令
	fmt.Printf("%s=%s\n", name, str)
	//将当前使用的寄存器压入堆栈
	a.regiserStack = append(a.regiserStack, name)
}

func (a *AugmentedParser) op(what string) {
	/*
		将寄存器堆栈顶部两个寄存器取出，生成一条计算指令，
		并赋值给第二个寄存器，然后释放第一个寄存器，第二个寄存器依然保持在堆栈上
	*/
	right := a.regiserStack[len(a.regiserStack)-1]
	a.regiserStack = a.regiserStack[0 : len(a.regiserStack)-1]
	left := a.regiserStack[len(a.regiserStack)-1]
	fmt.Printf("%s %s= %s\n", left, what, right)
	a.freeName(right)
}

func (a *AugmentedParser) getToken() lexer.Token {
	//先看看有没有上次退回去的 token
	if len(a.reverseToken) > 0 {
		token := a.reverseToken[len(a.reverseToken)-1]
		a.reverseToken = a.reverseToken[0 : len(a.reverseToken)-1]
		return token
	}

	token, err := a.parserLexer.Scan()
	if err != nil && token.Tag != lexer.EOF {
		sErr := fmt.Sprintf("get token with err:%s\n", err)
		panic(sErr)
	}

	return token
}

func (a *AugmentedParser) match(tag lexer.Tag) bool {
	token := a.getToken()
	if token.Tag != tag {
		a.putbackToken(token)
		return false
	}

	return true
}

func (a *AugmentedParser) Parse() {
	a.stmt()
}

func (a *AugmentedParser) isEOF() bool {
	token := a.getToken()
	if token.Tag == lexer.EOF {
		return true
	} else {
		a.putbackToken(token)
	}
	return false
}

func (a *AugmentedParser) stmt() {
	//stmt-> epsilon
	if a.isEOF() {
		return
	}
	//stmt -> expr ; stmt
	a.expr()
	if a.match(lexer.SEMI) != true {
		panic("mismatch token, expect semi")
	}
	a.stmt()
}

func (a *AugmentedParser) expr() {
	//expr -> term expr_prime
	a.term()
	a.expr_prime()
}

func (a *AugmentedParser) expr_prime() {
	//expr_prime -> + term {op('+')} expr_prime
	if a.match(lexer.PLUS) == true {
		a.term()
		a.op("+")
		a.expr_prime()
	}

	//expr -> epsilon
	return
}

func (a *AugmentedParser) term() {
	//term -> factor term_prime
	a.factor()
	a.term_prime()
}

func (a *AugmentedParser) term_prime() {
	//term_prime -> * factor {op('*')} term_prime
	if a.match(lexer.MUL) == true {
		a.factor()
		a.op("*")
		a.term_prime()
	}
	//term_prime -> epsilon
	return
}

func (a *AugmentedParser) factor() {
	// factor -> NUM {create_tmp(lexer.lexeme)}
	if a.match(lexer.NUM) == true {
		a.createTmp(a.parserLexer.Lexeme)
		return
	} else if a.match(lexer.LEFT_BRACKET) == true {
		a.expr()
		if a.match(lexer.RIGHT_BRACKET) != true {
			panic("mismatch token, expect right_paren")
		}
		return
	}

	//should not come here
	panic("factor parsing error")
}
