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
	Typ   rune
	Start int
	End   int
	Buf   []rune
	Toks  []*Token
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

func (tok *Token) Name() string {
	a := fmt.Sprintf("%c ", tok.Typ)
	var b string
	if len(tok.Buf) > 0 {
	   switch tok.Typ {
		case 'R':
			b = fmt.Sprintf("REG[%d]", tok.Buf[0])
		case 'K':
		   b = fmt.Sprintf("{%s}", KEYWORDS[tok.Buf[0]])
		case 'S':
			b = fmt.Sprintf("{%c}", tok.Buf[0])
			if tok.Buf[1] >= 0 {
			   b = fmt.Sprintf("%s -> %d", b, tok.Buf[1])
			}
		default:
			b = fmt.Sprintf("'%s'", string(tok.Buf))
	   }
	}
	c := fmt.Sprintf(" - %d:%d", tok.Start, tok.End)
	return a + b + c
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
      fmt.Println(tok.Name())
	} else {
	   fmt.Printf("@NIL\n")
	   return
	}

	newPrefix := make([]bool, len(prefix)+1)
	copy(newPrefix, prefix)
	for i, child := range tok.Toks {
		last := i == len(tok.Toks)-1
		newPrefix[len(newPrefix)-1] = !last
		child.dump_node(newPrefix, last)
	}
}



func (tok *Token) BuildNodes(printer func(*Token,string,string)) {
	tok.build_node([]bool{}, false, printer)
}

func (tok *Token) build_node(prefix []bool, isLast bool, printer func(*Token,string,string)) {
   var p, n string

	for i, hasNext := range prefix {
		if i == len(prefix)-1 {
			if isLast {
				p += fmt.Sprint("└── ")
			} else {
				p += fmt.Sprint("├── ")
			}
		} else {
			if hasNext {
				p += fmt.Sprint("│   ")
			} else {
				p += fmt.Sprint("    ")
			}
		}
	}

   if tok != nil {
      n = tok.Name()
   	printer(tok, p, n)
	} else {
	   n = "@NIL"
   	printer(tok, p, n)
   	return
	}


	newPrefix := make([]bool, len(prefix)+1)
	copy(newPrefix, prefix)
	for i, child := range tok.Toks {
		last := i == len(tok.Toks)-1
		newPrefix[len(newPrefix)-1] = !last
		child.build_node(newPrefix, last, printer)
	}
}
