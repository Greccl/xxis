package main

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
	if i == 0 { return s }
	// out := make([]rune, len(s)-i)
	// copy(out, s[i:])
	out := s[i:]
	return out
}

func rstrip(s []rune) []rune {
	i := len(s) - 1
	for i > 0 && is_space(s[i]) {
		i--
	}
	if i == len(s) - 1 { return s }
	// out := make([]rune, i)
	// copy(out, s[:i+1])
	out := s[:i+1]
	return out
}

func is_space(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func trim_spaces(s []rune) []rune {
	start := 0
	end := len(s)
	for start < end && is_space(s[start]) {
		start++
	}
	for end > start && is_space(s[end-1]) {
		end--
	}
	return s[start:end]
}

func equal_runes_str(s []rune, word string) bool {
	if len(s) != len(word) {
		return false
	}
	for i, r := range word {
		if s[i] != r {
			return false
		}
	}
	return true
}

func strip_prefix(tok *Token, i int) {
	tok.buf = lstrip(tok.buf[i:])
}

func find(s []rune, target rune, from int) int {
	for i, r := range s {
	   if i < from { continue }
		if r == target {
			return i
		}
	}
	return -1
}

func find_rune_in_T(toks []*Token, r rune, index, pos int) (int, int) {
   if index >= len(toks) {
      return -1, 0
   }
   if pos >= len(toks[index].buf) {
      return -1, 0
   }
   p := pos
	for i := index; i < len(toks); i++ {
		t := toks[i]
		if t.typ != 'T' {
			continue
		}
		found := find(t.buf, r, p)
		if found != -1 {
			return i, found
		}
		p = 0
	}
	return -1, 0
}

func split_by_colon(tokens []*Token) ([]*Token, []*Token) {
	index, pos := find_rune_in_T(tokens, ':', 0, 0)

	if index == -1 {
		return tokens, nil
	}

   return split_tokens_at(tokens, index, pos)
}

func split_tokens_at(tokens []*Token, index, pos int) ([]*Token, []*Token) {
	text := tokens[index].buf

	var cond []*Token
	left := rstrip(text[:pos+1])
	if len(left) > 0 {
		cond = make([]*Token, index+1)
		cond[index] = &Token{typ: 'T', buf: left}
	} else {
		cond = make([]*Token, index)
	}
	copy(cond, tokens[:index])

	var body []*Token
	right := lstrip(text[pos+1:])
	if len(right) > 0 {
		body = make([]*Token, len(tokens)-index)
		body[0] = &Token{typ: 'T', buf: right}
	} else {
		body = make([]*Token, len(tokens)-index-1)
	}
	copy(body[1:], tokens[index+1:])

	return cond, body
}



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


func parse_if(tok *Token) (*Token, *Token) {
	strip_prefix(tok.toks[0], 2)
	cond, body := split_by_colon(tok.toks)
	if body == nil {
		return &Token{typ: 'C', toks: cond}, nil
	} else {
		return &Token{typ: 'C', toks: cond}, &Token{typ: 'C', toks: body}
	}
}









func is_single_word(tok *Token, word string) bool {
	if tok.typ != 'I' {
		return false
	}
	if len(tok.toks) != 1 || tok.toks[0].typ != 'T' {
		return false
	}
	return equal_runes_str(trim_spaces(tok.toks[0].buf), word)
}

func is_if_cmd(tok *Token) bool {
	if tok.typ != 'I' {
   	return false
	}
	if len(tok.toks) == 0 || tok.toks[0].typ != 'T' {
		return false
	}
	start := []rune("if")
	return compare_runes(start, tok.toks[0].buf)
}
