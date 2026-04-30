package parser

import "unicode"
import "fmt"

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

func trimBufferRange(buf []rune, start int, left, right bool) ([]rune, int) {
	from := 0
	to := len(buf)

	if left {
		for from < to && unicode.IsSpace(buf[from]) {
			from++
		}
	}

	if right {
		for to > from && unicode.IsSpace(buf[to-1]) {
			to--
		}
	}

	return buf[from:to], start + from
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
		if i < from {
			continue
		}
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
	tok.Start += i + 1
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
	textStart := tokens[index].Start

	var cond []*Token
	left, leftStart := trimBufferRange(text[:pos], textStart, false, true)
	if len(left) > 0 {
		cond = make([]*Token, index+1)
		cond[index] = &Token{Typ: 'T', Buf: left, Start: leftStart, End: leftStart + len(left)}
	} else {
		cond = make([]*Token, index)
	}
	copy(cond, tokens[:index])

	var body []*Token
	right, rightStart := trimBufferRange(text[pos+1:], textStart+pos+1, true, false)
	if len(right) > 0 {
		body = make([]*Token, len(tokens)-index)
		body[0] = &Token{Typ: 'T', Buf: right, Start: rightStart, End: rightStart + len(right)}
		copy(body[1:], tokens[index+1:])
	} else {
		body = make([]*Token, len(tokens)-index-1)
		copy(body, tokens[index+1:])
	}

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
		c := split_and_or(&Token{Typ: 'C', Toks: cond})
		return c, nil
	} else {
		c := split_and_or(&Token{Typ: 'C', Toks: cond})
		b := split_and_or(&Token{Typ: 'C', Toks: body})
		inheritRangeFromChildren(b)
		return c, b
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
				return i, '/'
			}
		}
		return -1, 0
	}

	parts := make([]*Token, 0, 2)
	ops := make([]rune, 0, 1)
	curr := make([]*Token, 0)
	stripLeft := false
	found := false

	appendText := func(dst []*Token, buf []rune, start int) []*Token {
		if len(buf) == 0 {
			return dst
		}
		return append(dst, &Token{
			Typ:   'T',
			Buf:   buf,
			Start: start,
			End:   start + len(buf),
		})
	}

	appendPart := func(part []*Token, op rune) {
		c := &Token{Typ: 'C', Toks: part}
		inheritRangeFromChildren(c)
		parts = append(parts, c)
		ops = append(ops, op)
	}

	for _, t := range tok.Toks {
		if t.Typ != 'T' {
			if stripLeft {
				stripLeft = false
			}
			curr = append(curr, t)
			continue
		}

		buf := t.Buf
		bufStart := t.Start
		if stripLeft {
			buf, bufStart = trimBufferRange(buf, bufStart, true, false)
			stripLeft = false
		}

		pos := 0
		for {
			opPos, op := findOp(buf, pos)
			if opPos == -1 {
				if pos == 0 {
					curr = appendText(curr, buf, bufStart)
				} else if pos < len(buf) {
					frag := buf[pos:]
					curr = appendText(curr, frag, bufStart+pos)
				}
				break
			}

			found = true
			left, leftStart := trimBufferRange(buf[pos:opPos], bufStart+pos, false, true)
			curr = appendText(curr, left, leftStart)

			if len(curr) > 0 {
				appendPart(curr, op)
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
		part := &Token{Typ: 'C', Toks: curr}
		inheritRangeFromChildren(part)
		parts = append(parts, part)
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
	inheritRangeFromChildren(block)
	return block
}

func inheritRangeFromChildren(tok *Token) {
	if len(tok.Toks) == 0 {
		return
	}
	first := tok.Toks[0]
	last := tok.Toks[len(tok.Toks)-1]
	if first != nil {
		tok.Start = first.Start
	}
	if last != nil {
		tok.End = last.End
	}
}

type ParseError struct {
	Msg   string
	Start int
	End   int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("ParseError: %s", e.Msg)
}
