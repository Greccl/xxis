package main

const (
	CMD rune = iota + 128
	IF
)

type Operation func(*Metatoken)

type Node struct {
	typ string
	op Operation
	tok *Metatoken
}

func (n *Node) dump() {
	n.dump_node([]bool{}, false)
}

func (n *Node) dump_node(prefix []bool, isLast bool) {
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

	fmt.Printf(" '%s'", n.typ)
	fmt.Println()

    newPrefix := make([]bool, len(prefix)+1)
    copy(newPrefix, prefix)
    newPrefix[len(prefix)] = !isLast
    for i, child := range n.toks {
        last := i == len(n.toks)-1
        child.dump_node(newPrefix, last)
    }
}









func do_subcommand(tok *Metatoken) {
	
}

func do_if(tok *Metatoken) {
	
}



func compare_runes(a, b []rune) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func lstrip(s []rune) []rune {
	i := 0
	for i < len(s) && is_space(s[i]) {
		i++
	}
	out := make([]rune, len(s)-i)
	copy(out, s[i:])
	return out
}

func rstrip(s []rune) []rune {
	i := len(s) - 1
	for i > 0 && is_space(s[i]) { i-- }
	out := make([]rune, i)
	copy(out, s[:i])
	return out
}

func is_space(r rune) bool {
    return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func strip_prefix(tok *Metatoken, i int) {
   tok.buf = lstrip(tok.buf[i:])
}

func find(s []rune, target rune) int {
   for i, r := range s {
		if r == target { return i }
	}
	return -1
}

func find_colon(toks []*Metatoken) (int, int) {
	for i := 0; i < len(toks); i++ {
		t := toks[i]
      if t.typ != 'T' { continue }
      pos := find(t.buf, ':')
      if pos != -1 { return i, pos }
   }
   return -1, 0
}

func split_by_colon(tokens []*Metatoken) ([]*Metatoken, []*Metatoken) {
   index, pos := find_colon(tokens)

   if index == -1 {
   	return tokens, nil
	}

   text := tokens[index].buf
    
   var cond []*Metatoken
   left := rstrip(text[:pos])
   if len(left) > 0 {
   	cond = make([]*Metatoken, index + 1)
   	cond[index-1] = &Metatoken{typ:'T', buf:left}
   } else {
   	cond = make([]*Metatoken, index)
   }
   copy(cond, tokens[:index])

   var body []*Metatoken
   right := lstrip(text[pos+1:])
   if len(right) > 0 {
   	body = make([]*Metatoken, len(tokens) - index)
   	body[0] = &Metatoken{typ:'T', buf:right}
   } else {
   	body = make([]*Metatoken, len(tokens) - index - 1)
   }
   copy(body[1:], tokens[index+1:])

   return cond, body
}
















func parse_if(tok *Metatoken) *Node {
   strip_prefix(tok.toks[0], 3)
   cond, body := split_by_colon(tok.toks)

   tok.typ = 'K'
   if body != nil {
   	tok.toks = make([]*Metatoken, 2)
   	tok.toks[0] = &Metatoken{toks:cond}
   	tok.toks[1] = &Metatoken{toks:body}
   } else {
   	tok.toks = make([]*Metatoken, 1)
   	tok.toks[0] = &Metatoken{toks:cond}
   }
   
   node := &Node{}
	node.typ = "if"
   node.op = do_if
   node.tok = tok
   return node
}

func is_if(tok *Metatoken) bool {
	start := []rune("if ")
   return compare_runes(start, tok.buf)
}



func parse_expression(tok *Metatoken) *Node {
	node := &Node{}
	node.typ = "expr"
	node.op = do_subcommand
	node.tok = tok
	tok.typ = 'S'
	return node
}




func process_token(base *Metatoken) *Node {
	if base.typ != 'B' {
		return nil
	}
   if len(base.toks) == 0 {
   	return nil
   }
   first := base.toks[0]
   if first.typ == 'T' {
      if is_if(first) {
         return parse_if(base)
      }
   }
   return parse_expression(base)
}
