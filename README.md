我们前面章节看到的语法规则中，语法只给出了代码字符串组合规则是否符合规定，实际上我们可以在语法解析过程中增加一些特定的属性或者操作，使得语法解析流程中就能完成中间代码生成，或者是创建好特定的元信息，以便在后续处理流程中辅助代码生成。例如我们看看如何在语法解析规则中附加特定操作，使得语法解析过程就能生成中间代码，我们看一个例子，给定如下语法规则：

expr_prime -> + term {op('+');} expr_prime
其中{op(‘+’)}就是对语法的增强，它表示在解析完 term 这个符号后，执行 op(‘+’)这个操作，对应到代码实现上就如下所示：

expr_prime() {
    if (match(PLUS)) {
        term()
        op('+')
        expr_prime()
    }
}
要想理解增强语法的特性，我们还是需要去实现一个具体实例，我们现给出一个能解析算术表达式的增强语法规则：

stmt -> epsilon | expr ; stmt
expr -> term expr_prime 
expr_prime -> + term {op('+')} expr_prime | epsilon 
term -> factor term_prime 
term_prime -> * factor {op('*')} term_prime
factor -> NUM {create_tmp(lexer.lexeme)} | ( expr )
上面的语法用于识别加法，乘法，以及带括号的算术表达式，他跟我们前面用于识别表达式的语法有所不同，它这里主要是进行了“左递归消除”，在后续章节我们会详细讨论这个话题，那么我们怎么用上面语法来解析表达式呢，解析完毕后会有什么效果呢，我们看具体实现你就会明白了。首先在前dragon-compiler 项目中创建一个文件夹叫augmented_parser,在该目录下创建新文件叫 augmented_parser.go,添加代码如下：

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
在代码实现中有几处需要留意，一是代码存储了多个虚拟寄存器的名称，在读取表达式时，一旦读取到数字字符，那么就会将其数值赋值给某个寄存器，例如“1+2”，当代码读取字符1 时就会取出寄存器 t0,然后生成语句 t0=1，这个功能是由 createTmp 函数实现，调用该接口时输入的参数就对应当前读取到的数字。

在前面的语法规则中有{op(‘+’)}这样的指令，它在代码中对应函数 op，该函数从当前指令堆栈中取出顶部两个寄存处，然后执行加法指令，假设当前栈顶两个寄存器是 t0,t1,那么 op(‘+’)执行后就会创建指令 t1+=t0，然后它会把 t0 从堆栈去除，但是会保留 t1 在堆栈顶部。

在 main.go 中调用上面实现的代码测试一下效果：

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
上面代码执行后所得结果如下：

t0=1
t1=2
t2=4
t3=3
t2 += t3
t1 *= t2
t0 += t1
可以看成生成的虚拟指令确实能对应得上给定的算术表达式，更详细的调试演示过程请在 b 站搜索 coding 迪斯尼。代码下载：

