package parser

import "unicode"



func Trim_buffer(buf []rune, left, right bool) []rune {
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





func Is_space(r rune) bool {
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
	tok.Buf = Trim_buffer(tok.Buf[i:], true, false)
}

func find_rune_in_T(toks []*Token, r rune, index, pos int) (int, int) {
   if index >= len(toks) {
      return -1, 0
   }
   if pos >= len(toks[index].Buf) {
      return -1, 0
   }
   p := pos
	for i := index; i < len(toks); i++ {
		t := toks[i]
		if t.Typ != 'T' {
			continue
		}
		found := find(t.Buf, r, p)
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
	text := tokens[index].Buf

	var cond []*Token
	left := Trim_buffer(text[:pos], false, true)
	if len(left) > 0 {
		cond = make([]*Token, index+1)
		cond[index] = &Token{Typ: 'T', Buf: left}
	} else {
		cond = make([]*Token, index)
	}
	copy(cond, tokens[:index])

	var body []*Token
	right := Trim_buffer(text[pos+1:], true, false)
	if len(right) > 0 {
		body = make([]*Token, len(tokens)-index)
		body[0] = &Token{Typ: 'T', Buf: right}
	} else {
		body = make([]*Token, len(tokens)-index-1)
	}
	copy(body[1:], tokens[index+1:])

	return cond, body
}





func is_single_word(tok *Token, word string) bool {
	if tok.Typ != 'I' {
		return false
	}
	if len(tok.Toks) != 1 || tok.Toks[0].Typ != 'T' {
		return false
	}
	return equal_runes_str(Trim_buffer(tok.Toks[0].Buf, true, true), word)
}

func is_if_cmd(tok *Token) bool {
	if tok.Typ != 'I' {
   	return false
	}
	if len(tok.Toks) == 0 || tok.Toks[0].Typ != 'T' {
		return false
	}
	start := []rune("if")
	return compare_runes(start, tok.Toks[0].Buf)
}

func parse_if(tok *Token) (*Token, *Token) {
	strip_prefix(tok.Toks[0], 2)
	cond, body := split_by_colon(tok.Toks)
	if body == nil {
		// return &Token{Typ: 'C', Toks: cond}, nil
		return split_and_or(&Token{Typ: 'C', Toks: cond}), nil
	} else {
		// return &Token{Typ: 'C', Toks: cond}, &Token{Typ: 'C', Toks: body}
		return split_and_or(&Token{Typ: 'C', Toks: cond}), split_and_or(&Token{Typ: 'C', Toks: body})
	}
}

func split_and_or(tok *Token) *Token {
	if tok == nil || tok.Typ != 'C' {
		return tok
	}

	findOp := func(buf []rune, from int) (int, rune) {
		for i := from; i < len(buf)-1; i++ {
			if buf[i] == '&' && buf[i+1] == '&' {
				return i, '&'
			}
			if buf[i] == '|' && buf[i+1] == '|' {
				return i, '|'
			}
		}
		return -1, 0
	}

	parts := make([]*Token, 0, 2)
	ops := make([]rune, 0, 1)
	curr := make([]*Token, 0)
	stripLeft := false
	found := false

	for _, t := range tok.Toks {
		if t.Typ != 'T' {
			if stripLeft {
				stripLeft = false
			}
			curr = append(curr, t)
			continue
		}

		buf := t.Buf
		if stripLeft {
			buf = Trim_buffer(buf, true, false)
			stripLeft = false
		}

		pos := 0
		for {
			opPos, op := findOp(buf, pos)
			if opPos == -1 {
				if pos == 0 {
					if len(buf) > 0 {
						curr = append(curr, &Token{Typ: 'T', Buf: buf})
					}
				} else if pos < len(buf) {
					frag := buf[pos:]
					if len(frag) > 0 {
						curr = append(curr, &Token{Typ: 'T', Buf: frag})
					}
				}
				break
			}

			found = true
			left := Trim_buffer(buf[pos:opPos], false, true)
			if len(left) > 0 {
				curr = append(curr, &Token{Typ: 'T', Buf: left})
			}

			if len(curr) > 0 {
				parts = append(parts, &Token{Typ: 'C', Toks: curr})
				ops = append(ops, op)
			}
			curr = make([]*Token, 0)

			pos = opPos + 2
			for pos < len(buf) && Is_space(buf[pos]) {
				pos++
			}
			if pos >= len(buf) {
				stripLeft = true
				break
			}
		}
	}

	if !found {
		return tok
	}
	if len(curr) > 0 {
		parts = append(parts, &Token{Typ: 'C', Toks: curr})
	}
	if len(parts) == 0 || len(ops) == 0 || len(parts) != len(ops)+1 {
		return tok
	}

	block := &Token{Typ: 'B', Toks: make([]*Token, len(parts))}
	for i, part := range parts {
		if i > 0 {
			part.Typ = ops[i-1]
		}
		block.Toks[i] = part
	}
	return block
}