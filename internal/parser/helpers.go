package parser





func trim_buffer(buf []rune, left, right bool) []rune {
	start := 0
	end := len(buf)

	if left {
		for start < end && unicode.IsSpace(buf[start]) {
			start++
		}
	}

	if right {
		for end > start && unicode.IsSpace(buf[end-1]) {
			end--
		}
	}

	return buf[start:end]
}

func compare_runes(a, b []rune) bool {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equal_runes_str(s []rune, word string) bool {
   w := []rune(word)
	if len(s) != len(w) {
		return false
	}
	for i, r := range w {
		if s[i] != r {
			return false
		}
	}
	return true
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





func is_space(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func is_iden_start(r rune) bool {
	if 'a' <= r && r <= 'z' {
		return true
	}
	return 'A' <= r && r <= 'Z'
}

func is_iden_char(r rune) bool {
	if 'a' <= r && r <= 'z' {
		return true
	}
	if 'A' <= r && r <= 'Z' {
		return true
	}
	if '0' <= r && r <= '9' {
		return true
	}
	return r == '_'
}





func strip_prefix(tok *Token, i int) {
	tok.buf = trim_buffer(tok.buf[i:], true, false)
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
	left := trim_buffer(text[:pos], false, true)
	if len(left) > 0 {
		cond = make([]*Token, index+1)
		cond[index] = &Token{typ: 'T', buf: left}
	} else {
		cond = make([]*Token, index)
	}
	copy(cond, tokens[:index])

	var body []*Token
	right := trim_buffer(text[pos+1:], true, false)
	if len(right) > 0 {
		body = make([]*Token, len(tokens)-index)
		body[0] = &Token{typ: 'T', buf: right}
	} else {
		body = make([]*Token, len(tokens)-index-1)
	}
	copy(body[1:], tokens[index+1:])

	return cond, body
}





func is_single_word(tok *Token, word string) bool {
	if tok.typ != 'I' {
		return false
	}
	if len(tok.toks) != 1 || tok.toks[0].typ != 'T' {
		return false
	}
	return equal_runes_str(trim_buffer(tok.toks[0].buf, true, true), word)
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

func parse_if(tok *Token) (*Token, *Token) {
	strip_prefix(tok.toks[0], 2)
	cond, body := split_by_colon(tok.toks)
	if body == nil {
		return &Token{typ: 'C', toks: cond}, nil
	} else {
		return &Token{typ: 'C', toks: cond}, &Token{typ: 'C', toks: body}
	}
}