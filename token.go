package main

import (
	"fmt"
)

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
		name := KEYWORDS[tok.buf[0]]
		if len(tok.toks) == 3 {
			return fmt.Sprintf("%s {%c}:{%c}:{%c}", name, tok.toks[0].typ, tok.toks[1].typ, tok.toks[2].typ)
		} else if len(tok.toks) == 2 {
			return fmt.Sprintf("%s {%c}:{%c}", name, tok.toks[0].typ, tok.toks[1].typ)
		} else if len(tok.toks) == 1 {
			return fmt.Sprintf("%s {%c}", name, tok.toks[0].typ)
		} else {
			return name
		}
	}
	var items string
	for _, t := range tok.toks {
		items += t.repr()
	}
	switch tok.typ {
	case 'S', 'C':
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

   if tok != nil {
   	fmt.Printf("%c", tok.typ)
   	if len(tok.buf) > 0 {
   		if tok.typ == 'R' {
   			fmt.Printf(" REG[%d]", tok.buf[0])
   		} else if tok.typ == 'K' {
   		   fmt.Printf(" < %s >", KEYWORDS[tok.buf[0]])
   		} else {
   			fmt.Printf(" '%s'", string(tok.buf))
   		}
   	}
	   fmt.Println()
	} else {
	   fmt.Printf("@NIL\n")
	   return
	}

	newPrefix := make([]bool, len(prefix)+1)
	copy(newPrefix, prefix)
	for i, child := range tok.toks {
		last := i == len(tok.toks)-1
		newPrefix[len(newPrefix)-1] = !last
		// newPrefix[len(newPrefix)-1] = !isLast && !last
		child.dump_node(newPrefix, last)
	}
}
