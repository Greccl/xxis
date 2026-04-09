package main

import (
	"fmt"
)

// DataToken TVR
// BoxToken SQ

type Token struct {
	typ  rune
	buf  []rune
	toks []*Token
}

func (tok *Token) repr() string {
	switch tok.typ {
	case 'T':
		return fmt.Sprintf("%s", string(tok.buf))
	case 'V':
		return fmt.Sprintf("$%s", string(tok.buf))
	case 'R':
		return fmt.Sprintf("{R%d}", tok.buf[0])
	case 'K':
		name := string(tok.buf)
		if len(tok.toks) >= 2 {
			return fmt.Sprintf("%s {%s}:{%s}", name, tok.toks[0].repr(), tok.toks[1].repr())
		} else if len(tok.toks) == 1 {
			return fmt.Sprintf("%s {%s}", name, tok.toks[0].repr())
		} else {
			return name
		}
	case 'C':
		if len(tok.toks) == 1 {
			return fmt.Sprintf("CMD   %s", tok.toks[0].repr())
		}
		return "CMD   "
	}
	var items string
	for _, t := range tok.toks {
		items += t.repr()
	}
	switch tok.typ {
	case 'S':
		if len(tok.buf) > 0 {
			return fmt.Sprintf("CMD R%d %s", tok.buf[0], items)
		} else {
			return fmt.Sprintf("CMD -- %s", items)
		}
	case 'Q':
		return fmt.Sprintf("'%s'", items)
	}
	return "???"
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
		newPrefix[len(newPrefix)-1] = !last
		// newPrefix[len(newPrefix)-1] = !isLast && !last
		child.dump_node(newPrefix, last)
	}
}
