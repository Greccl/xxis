package main

import (
	"fmt"
)


// DataToken TVR
// BoxToken SQ


type Token struct {
	typ rune
	buf []rune
	toks []*Token
}


func (tok *Token) repr() string {
	if len(tok.buf) > 0 {
		if tok.typ == 'R' {
			return fmt.Sprintf("%c REG[%d]", tok.typ, tok.buf[0])
		} else {
			return fmt.Sprintf("%c '%s'", tok.typ, string(tok.buf))
		}
	}
	return fmt.Sprintf("%c", tok.typ)
}

func (tok *Token) dump() {
	tok.dump_node([]bool{}, false)
}

func (tok *Token) dump_node(prefix []bool, isLast bool) {
    for i, hasNext := range prefix {
        if i == len(prefix)-1 {
            if isLast {
                fmt.Print("└── ")
            } else {
                fmt.Print("├── ")
            }
        } else {
            if hasNext {
                fmt.Print("│   ")
            } else {
                fmt.Print("    ")
            }
        }
    }

	fmt.Printf("%c", tok.typ)
	if len(tok.buf) > 0 {
		if tok.typ == 'R' {
			fmt.Printf(" REG[%d]", tok.buf[0])
		} else {
			fmt.Printf(" '%s'", string(tok.buf))
		}
	}
	fmt.Println()

	newPrefix := make([]bool, len(prefix)+1)
	copy(newPrefix, prefix)
   for i, child := range tok.toks {
   	last := i == len(tok.toks)-1
		newPrefix[len(newPrefix)-1] = !isLast && !last
      child.dump_node(newPrefix, last)
   }
}
