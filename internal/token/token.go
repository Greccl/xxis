package token

import (
	"fmt"
)

const (
   IF rune = iota
   IFZ
   IFN
   FOR
)

var KEYWORDS = []string{
   "if",
   "ifz",
   "ifn",
   "for",
}


type Token struct {
	Typ  rune
	Buf  []rune
	Toks []*Token
}

func (tok *Token) Repr() string {
	switch tok.Typ {
	case 'T':
		return fmt.Sprintf("%s", string(tok.Buf))
	case 'V':
		return fmt.Sprintf("$%s", string(tok.Buf))
	case 'R':
		return fmt.Sprintf("{R%d}", tok.Buf[0])
	case 'K':
		name := KEYWORDS[tok.Buf[0]]
		if len(tok.Toks) == 3 {
			return fmt.Sprintf("%s {%c}:{%c}:{%c}", name, tok.Toks[0].Typ, tok.Toks[1].Typ, tok.Toks[2].Typ)
		} else if len(tok.Toks) == 2 {
			return fmt.Sprintf("%s {%c}:{%c}", name, tok.Toks[0].Typ, tok.Toks[1].Typ)
		} else if len(tok.Toks) == 1 {
			return fmt.Sprintf("%s {%c}", name, tok.Toks[0].Typ)
		} else {
			return name
		}
	}
	var items string
	for _, t := range tok.Toks {
		items += t.Repr()
	}
	switch tok.Typ {
	case 'S':
		return fmt.Sprintf("CMD R%d (%c) %s", tok.Buf[1], tok.Buf[0], items)
	case 'C':
		return fmt.Sprintf("CMD -- %s", items)
	case 'Q':
		return fmt.Sprintf("'%s'", items)
	}
	return "???"
}

func (tok *Token) Dump() {
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
   	fmt.Printf("%c", tok.Typ)
   	if len(tok.Buf) > 0 {
   	   switch tok.Typ {
   		case 'R':
   			fmt.Printf(" REG[%d]", tok.Buf[0])
   		case 'K':
   		   fmt.Printf(" {%s}", KEYWORDS[tok.Buf[0]])
   		case 'S':
   			fmt.Printf(" {%c}", tok.Buf[0])
   			if tok.Buf[1] >= 0 {
   			   fmt.Printf(" -> %d", tok.Buf[1])
   			}
   		default:
   			fmt.Printf(" '%s'", string(tok.Buf))
   	   }
   	}
	   fmt.Println()
	} else {
	   fmt.Printf("@NIL\n")
	   return
	}

	newPrefix := make([]bool, len(prefix)+1)
	copy(newPrefix, prefix)
	for i, child := range tok.Toks {
		last := i == len(tok.Toks)-1
		newPrefix[len(newPrefix)-1] = !last
		// newPrefix[len(newPrefix)-1] = !isLast && !last
		child.dump_node(newPrefix, last)
	}
}
