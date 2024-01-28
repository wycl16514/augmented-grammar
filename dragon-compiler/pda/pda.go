package pda

import (
	"fmt"
)

const (
	ERROR = iota
	ACCEPT
	PUSH1
	POP
	EOF
)

type StateTable struct{}

func (s *StateTable) get(state int8, symbol int) int8 {
	if state == 0 {
		switch symbol {
		case '(':
			return PUSH1
		case ')':
			return ERROR
		case EOF:
			return ACCEPT
		}
	}

	if state == 1 {
		switch symbol {
		case '(':
			return PUSH1
		case ')':
			return POP
		case EOF:
			return ERROR
		}
	}

	panic("state can only 0 or 1")
}

type BracketPDA struct {
	stateTable *StateTable
	stack      []int8
}

func NewBracketPDA() *BracketPDA {
	pda := &BracketPDA{
		stateTable: &StateTable{},
		stack:      make([]int8, 0),
	}

	pda.stack = append(pda.stack, 0)

	return pda
}

func (b *BracketPDA) Parse(str string) {
	pos := 0
	for true {
		symbol := EOF
		if pos < len(str) {
			symbol = int(str[pos])
		}
		state := b.stack[len(b.stack)-1]
		action := b.stateTable.get(state, symbol)
		switch action {
		case ERROR:
			fmt.Printf("str: %s, is rejected\n", str)
			return
		case ACCEPT:
			fmt.Printf("str: %s, is accept\n", str)
			return
		case PUSH1:
			b.stack = append(b.stack, 1)
		case POP:
			b.stack = b.stack[:len(b.stack)-1]
		}

		pos += 1
		if symbol == EOF {
			return
		}
	}
}
