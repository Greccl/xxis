package parser

import "unicode"
import xxisExpr "github.com/Greccl/xxis/internal/expr"
import xxisToken "github.com/Greccl/xxis/internal/token"

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
	if i > len(tok.Buf) {
		i = len(tok.Buf)
	}
	buf, start := trimBufferRange(tok.Buf[i:], tok.Start+i, true, false)
	tok.Buf = buf
	tok.Start = start
	tok.End = start + len(buf)
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
	return is_keyword_cmd(tok, "if")
}

func is_function_cmd(tok *Token) bool {
	return is_keyword_cmd(tok, "function")
}

func is_var_cmd(tok *Token) bool {
	return is_keyword_cmd(tok, "var")
}

func is_import_cmd(tok *Token) bool {
	return is_keyword_cmd(tok, "import")
}

func is_source_cmd(tok *Token) bool {
	return is_keyword_cmd(tok, "source")
}

func is_keyword_cmd(tok *Token, word string) bool {
	if tok.Typ != 'I' {
		return false
	}
	if len(tok.Toks) == 0 || tok.Toks[0].Typ != 'T' {
		return false
	}
	buf := tok.Toks[0].Buf
	prefix := []rune(word)
	if len(buf) < len(prefix) {
		return false
	}
	if !compare_runes(prefix, buf) {
		return false
	}
	if len(buf) == len(prefix) {
		return true
	}
	return Is_space(buf[len(prefix)])
}

func parse_if(tok *Token) (*Token, *Token, rune, int) {
	strip_prefix(tok.Toks[0], 2)
	cond, body := split_by_colon(tok.Toks)

	mode, cmdCond, isCmdCond := parse_if_cond_mode(cond)
	if isCmdCond {
		if !has_if_cond_tokens(cmdCond) {
			panic(ErrorWithRange{
				Msg:   "if command condition requires a command",
				Start: tok.Start,
				End:   tok.End,
			})
		}
		c := split_and_or(&Token{Typ: 'C', Toks: cmdCond})
		return c, parse_if_body(body), mode, -1
	}

	node, _ := xxisExpr.ParseTokens(cond)
	exprIndex := xxisExpr.Store(node)
	c := &Token{Typ: 'C', Toks: cond}
	inheritRangeFromChildren(c)
	return c, parse_if_body(body), xxisToken.IfCondExpr, exprIndex
}

func parse_if_body(body []*Token) *Token {
	if body == nil {
		return nil
	}
	b := split_and_or(&Token{Typ: 'C', Toks: body})
	inheritRangeFromChildren(b)
	return b
}

func parse_if_cond_mode(tokens []*Token) (rune, []*Token, bool) {
	for i, tok := range tokens {
		if tok == nil {
			continue
		}
		if tok.Typ != 'T' {
			return xxisToken.IfCondExpr, tokens, false
		}

		buf, start := trimBufferRange(tok.Buf, tok.Start, true, false)
		if len(buf) == 0 {
			continue
		}

		wordEnd := 0
		for wordEnd < len(buf) && !Is_space(buf[wordEnd]) {
			wordEnd++
		}

		mode := xxisToken.IfCondExpr
		switch string(buf[:wordEnd]) {
		case "success":
			mode = xxisToken.IfCondSuccess
		case "failure":
			mode = xxisToken.IfCondFailure
		default:
			return xxisToken.IfCondExpr, tokens, false
		}

		rest, restStart := trimBufferRange(buf[wordEnd:], start+wordEnd, true, false)
		cmdTokens := make([]*Token, 0, len(tokens)-i)
		if len(rest) > 0 {
			cmdTokens = append(cmdTokens, &Token{Typ: 'T', Buf: rest, Start: restStart, End: restStart + len(rest)})
		}
		cmdTokens = append(cmdTokens, tokens[i+1:]...)
		return mode, cmdTokens, true
	}

	return xxisToken.IfCondExpr, tokens, false
}

func has_if_cond_tokens(tokens []*Token) bool {
	for _, tok := range tokens {
		if tok == nil {
			continue
		}
		if tok.Typ != 'T' {
			return true
		}
		if len(Trim_buffer(tok.Buf, true, true)) > 0 {
			return true
		}
	}
	return false
}

func parse_function(tok *Token) *Token {
	strip_prefix(tok.Toks[0], 8)

	params := tok.Toks
	if len(params) > 0 && params[0].Typ == 'T' && len(params[0].Buf) == 0 {
		params = params[1:]
	}

	node := &Token{Typ: 'C', Toks: params}
	inheritRangeFromChildren(node)
	if len(node.Toks) == 0 {
		node.Start = tok.End
		node.End = tok.End
	}
	return node
}

func parse_var(tok *Token) *Token {
	strip_prefix(tok.Toks[0], 3)

	tokens := tok.Toks
	if len(tokens) > 0 && tokens[0].Typ == 'T' && len(tokens[0].Buf) == 0 {
		tokens = tokens[1:]
	}

	scope := xxisToken.VarScopeLocal
	kind := xxisToken.VarTypeString
	index := 0
	pos := 0

	nextWord := func() (string, int, int, bool) {
		for index < len(tokens) {
			t := tokens[index]
			if t.Typ != 'T' {
				return "", t.Start, t.End, false
			}
			for pos < len(t.Buf) && Is_space(t.Buf[pos]) {
				pos++
			}
			if pos >= len(t.Buf) {
				index++
				pos = 0
				continue
			}

			start := pos
			for pos < len(t.Buf) && !Is_space(t.Buf[pos]) {
				pos++
			}
			return string(t.Buf[start:pos]), t.Start + start, t.Start + pos, true
		}
		return "", tok.End, tok.End, false
	}

	for {
		word, wordStart, wordEnd, ok := nextWord()
		if !ok {
			panic(ErrorWithRange{
				Msg:   "var requires a variable name",
				Start: wordStart,
				End:   wordEnd,
			})
		}

		if parsedScope, ok := parse_var_scope(word); ok {
			scope = parsedScope
			continue
		}
		if parsedType, ok := parse_var_type(word); ok {
			kind = parsedType
			continue
		}

		if !is_var_name(word) {
			panic(ErrorWithRange{
				Msg:   "invalid var name",
				Start: wordStart,
				End:   wordEnd,
			})
		}
		name := &Token{Typ: 'T', Buf: []rune(word), Start: wordStart, End: wordEnd}

		opWord, opStart, opEnd, ok := nextWord()
		if !ok {
			panic(ErrorWithRange{
				Msg:   "var requires an operator",
				Start: opStart,
				End:   opEnd,
			})
		}
		op, ok := parse_var_operator(opWord)
		if !ok {
			panic(ErrorWithRange{
				Msg:   "invalid var operator",
				Start: opStart,
				End:   opEnd,
			})
		}

		args := var_args_from(tokens, index, pos, opEnd)
		if op == xxisToken.VarOpDelete {
			if args != nil && len(args.Toks) > 0 {
				panic(ErrorWithRange{
					Msg:   "var delete does not accept arguments",
					Start: args.Start,
					End:   args.End,
				})
			}
			args = nil
		} else if args == nil || len(args.Toks) == 0 {
			panic(ErrorWithRange{
				Msg:   "var operator requires arguments",
				Start: opStart,
				End:   opEnd,
			})
		}

		return &Token{
			Typ:   'K',
			Buf:   []rune{xxisToken.VAR, scope, kind, op},
			Toks:  []*Token{name, args},
			Start: tok.Start,
			End:   tok.End,
		}
	}
}

type keywordArgScanner struct {
	tokens []*Token
	index  int
	pos    int
	end    int
}

func new_keyword_arg_scanner(tok *Token, prefixLen int) *keywordArgScanner {
	strip_prefix(tok.Toks[0], prefixLen)

	tokens := tok.Toks
	if len(tokens) > 0 && tokens[0].Typ == 'T' && len(tokens[0].Buf) == 0 {
		tokens = tokens[1:]
	}

	return &keywordArgScanner{tokens: tokens, end: tok.End}
}

func (self *keywordArgScanner) nextWord() (string, int, int, bool) {
	for self.index < len(self.tokens) {
		t := self.tokens[self.index]
		if t.Typ != 'T' {
			return "", t.Start, t.End, false
		}
		for self.pos < len(t.Buf) && Is_space(t.Buf[self.pos]) {
			self.pos++
		}
		if self.pos >= len(t.Buf) {
			self.index++
			self.pos = 0
			continue
		}

		start := self.pos
		for self.pos < len(t.Buf) && !Is_space(t.Buf[self.pos]) {
			self.pos++
		}
		return string(t.Buf[start:self.pos]), t.Start + start, t.Start + self.pos, true
	}
	return "", self.end, self.end, false
}

func (self *keywordArgScanner) nextPath() (string, int, int, bool) {
	for self.index < len(self.tokens) {
		t := self.tokens[self.index]
		if t.Typ == 'T' {
			for self.pos < len(t.Buf) && Is_space(t.Buf[self.pos]) {
				self.pos++
			}
			if self.pos >= len(t.Buf) {
				self.index++
				self.pos = 0
				continue
			}
			return self.nextWord()
		}

		if t.Typ == 'Q' {
			if self.pos != 0 {
				return "", t.Start, t.End, false
			}
			self.index++
			return quoted_literal(t)
		}

		return "", t.Start, t.End, false
	}
	return "", self.end, self.end, false
}

func (self *keywordArgScanner) empty() (int, int, bool) {
	for self.index < len(self.tokens) {
		t := self.tokens[self.index]
		if t.Typ != 'T' {
			return t.Start, t.End, false
		}
		for self.pos < len(t.Buf) {
			if !Is_space(t.Buf[self.pos]) {
				return t.Start + self.pos, t.End, false
			}
			self.pos++
		}
		self.index++
		self.pos = 0
	}
	return self.end, self.end, true
}

func quoted_literal(tok *Token) (string, int, int, bool) {
	var buf []rune
	for _, child := range tok.Toks {
		if child.Typ != 'T' {
			return "", child.Start, child.End, false
		}
		buf = append(buf, child.Buf...)
	}
	if len(buf) == 0 {
		return "", tok.Start, tok.End, false
	}
	return string(buf), tok.Start, tok.End, true
}

func parse_import(tok *Token) *Token {
	scanner := new_keyword_arg_scanner(tok, 6)

	path, pathStart, pathEnd, ok := scanner.nextPath()
	if !ok || path == "" {
		panic(ErrorWithRange{
			Msg:   "import requires a path",
			Start: pathStart,
			End:   pathEnd,
		})
	}

	word, wordStart, wordEnd, ok := scanner.nextWord()
	if !ok || word != "as" {
		panic(ErrorWithRange{
			Msg:   "import requires 'as'",
			Start: wordStart,
			End:   wordEnd,
		})
	}

	name, nameStart, nameEnd, ok := scanner.nextWord()
	if !ok {
		panic(ErrorWithRange{
			Msg:   "import requires a name",
			Start: nameStart,
			End:   nameEnd,
		})
	}
	if !is_var_name(name) {
		panic(ErrorWithRange{
			Msg:   "invalid import name",
			Start: nameStart,
			End:   nameEnd,
		})
	}

	extraStart, extraEnd, empty := scanner.empty()
	if !empty {
		panic(ErrorWithRange{
			Msg:   "import does not accept extra arguments",
			Start: extraStart,
			End:   extraEnd,
		})
	}

	return &Token{
		Typ: 'K',
		Buf: []rune{xxisToken.IMPORT},
		Toks: []*Token{
			{Typ: 'T', Buf: []rune(path), Start: pathStart, End: pathEnd},
			{Typ: 'T', Buf: []rune(name), Start: nameStart, End: nameEnd},
		},
		Start: tok.Start,
		End:   tok.End,
	}
}

func parse_source(tok *Token) *Token {
	scanner := new_keyword_arg_scanner(tok, 6)

	path, pathStart, pathEnd, ok := scanner.nextPath()
	if !ok || path == "" {
		panic(ErrorWithRange{
			Msg:   "source requires a path",
			Start: pathStart,
			End:   pathEnd,
		})
	}

	extraStart, extraEnd, empty := scanner.empty()
	if !empty {
		panic(ErrorWithRange{
			Msg:   "source does not accept extra arguments",
			Start: extraStart,
			End:   extraEnd,
		})
	}

	return &Token{
		Typ:   'K',
		Buf:   []rune{xxisToken.SOURCE},
		Toks:  []*Token{{Typ: 'T', Buf: []rune(path), Start: pathStart, End: pathEnd}},
		Start: tok.Start,
		End:   tok.End,
	}
}

func parse_var_scope(word string) (rune, bool) {
	switch word {
	case "-l", "--local":
		return xxisToken.VarScopeLocal, true
	case "-g", "--global":
		return xxisToken.VarScopeGlobal, true
	case "-u", "--universal":
		return xxisToken.VarScopeUniversal, true
	}
	return 0, false
}

func parse_var_type(word string) (rune, bool) {
	switch word {
	case ":s", ":str", ":string":
		return xxisToken.VarTypeString, true
	case ":n", ":num", ":number", ":float":
		return xxisToken.VarTypeFloat, true
	case ":j", ":job", ":proc", ":process", ":proceso":
		return xxisToken.VarTypeProcess, true
	case ":p", ":path":
		return xxisToken.VarTypePath, true
	case ":l", ":list", ":lista":
		return xxisToken.VarTypeList, true
	case ":d", ":dict", ":dictionary", ":diccionario":
		return xxisToken.VarTypeDict, true
	}
	return 0, false
}

func parse_var_operator(word string) (rune, bool) {
	switch word {
	case "=":
		return xxisToken.VarOpAssign, true
	case "<":
		return xxisToken.VarOpAppend, true
	case ">":
		return xxisToken.VarOpPrepend, true
	case "delete", "--delete":
		return xxisToken.VarOpDelete, true
	}
	return 0, false
}

func is_var_name(word string) bool {
	buf := []rune(word)
	if len(buf) == 0 || !is_iden_start(buf[0]) {
		return false
	}
	for _, r := range buf[1:] {
		if !is_iden_char(r) {
			return false
		}
	}
	return true
}

func var_args_from(tokens []*Token, index, pos, emptyAt int) *Token {
	args := make([]*Token, 0)
	if index < len(tokens) {
		t := tokens[index]
		if t.Typ == 'T' {
			buf, start := trimBufferRange(t.Buf[pos:], t.Start+pos, true, false)
			if len(buf) > 0 {
				ctx := &ParseContext0{}
				ctx.init()
				ctx.new_segment(Segment{typ: 'R', offset: start, buf: buf})
				parsed := subcmd_by_text(ctx)
				args = append(args, parsed.Toks...)
			}
			args = append(args, tokens[index+1:]...)
		} else {
			args = append(args, tokens[index:]...)
		}
	}

	node := &Token{Typ: 'C', Toks: args, Start: emptyAt, End: emptyAt}
	inheritRangeFromChildren(node)
	return node
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
