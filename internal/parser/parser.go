package parser

import xxisToken "github.com/Greccl/xxis/internal/token"

type Token = xxisToken.Token

const (
	tokenIndent rune = '>'
	tokenDedent rune = '<'
)

type Segment struct {
	typ    rune
	offset int
	buf    []rune
}

type TokenSource func() *Token

func Enumerate_tokens(read IndexedRuneSource) TokenSource {
	m := 0
	escape := false
	comment := false
	buf := make([]rune, 0)
	typ := 'R'
	var quote rune = 0
	strip := true

	instr := make([]Segment, 0)
	queue := make([]*Token, 0)
	done := false
	lineStart := true
	lineIndent := 0
	lineIndentHasSpaces := false
	indentStack := []int{0}

	expr := func(final bool) Segment {
		exp := buf
		if final {
			exp = Trim_buffer(exp, false, true)
		}
		buf = make([]rune, 0)
		if len(exp) > 0 {
			return Segment{typ, m, exp}
		}
		return Segment{}
	}

	enqueue := func(tok *Token) {
		queue = append(queue, tok)
	}

	popQueue := func() *Token {
		tok := queue[0]
		queue = queue[1:]
		return tok
	}

	emit := func(seg Segment) {
		switch seg.typ {
		case 'R', 'Q':
			instr = append(instr, seg)
		case ';':
			strip = true
			if len(instr) > 0 {
				enqueue(subcmd_by_segment(instr))
				instr = make([]Segment, 0)
			}
		}
	}

	emitIndentEvents := func(indent, offset int) {
		curr := indentStack[len(indentStack)-1]
		if indent > curr {
			indentStack = append(indentStack, indent)
			enqueue(&Token{Typ: tokenIndent, Start: offset, End: offset})
			return
		}
		for indent < curr {
			if len(indentStack) == 1 {
				panic(ParseError{
					Msg:   "unmatched dedent",
					Start: offset,
					End:   offset,
				})
			}
			indentStack = indentStack[:len(indentStack)-1]
			enqueue(&Token{Typ: tokenDedent, Start: offset, End: offset})
			curr = indentStack[len(indentStack)-1]
		}
		if indent != curr {
			panic(ParseError{
				Msg:   "unmatched dedent",
				Start: offset,
				End:   offset,
			})
		}
	}

	emitFinalDedents := func(offset int) {
		for len(indentStack) > 1 {
			indentStack = indentStack[:len(indentStack)-1]
			enqueue(&Token{Typ: tokenDedent, Start: offset, End: offset})
		}
	}

	yield := func() *Token {
		if len(queue) > 0 {
			return popQueue()
		}

		if done {
			return nil
		}

		for {
			i, r, eof := read()

			if lineStart && quote == 0 && !comment {
				if eof {
					emitFinalDedents(i)
					done = true
					if len(queue) > 0 {
						return popQueue()
					}
					return nil
				}
				switch r {
				case '\t':
					lineIndent++
					continue
				case ' ':
					lineIndentHasSpaces = true
					continue
				case '\n':
					lineIndent = 0
					lineIndentHasSpaces = false
					strip = true
					continue
				case '#':
					lineStart = false
					comment = true
					typ = 'C'
					m = i + 1
					continue
				}
				if lineIndentHasSpaces {
					panic(ParseError{
						Msg:   "indentation must use tabs",
						Start: i,
						End:   i + 1,
					})
				}
				emitIndentEvents(lineIndent, i)
				lineStart = false
				strip = false
				m = i
			}

			if escape {
				escape = false
				buf = append(buf, r)
				if !eof {
					continue
				}
			}

			if strip {
				switch r {
				case ' ', '\t':
					if !eof {
						continue
					}
				}
				strip = false
				m = i
			}

			if comment {
				if r == '\n' || eof {
					buf = buf[:0]
					strip = true
					comment = false
					typ = 'R'
					lineStart = true
					lineIndent = 0
					lineIndentHasSpaces = false
				} else {
					buf = append(buf, r)
				}
				if eof {
					if quote != 0 {
						panic("unclosed quote")
					}
					emitFinalDedents(i)
					done = true
					if len(queue) > 0 {
						return popQueue()
					}
					return nil
				}
				continue
			}

			if quote == 0 && r == '#' {
				e := expr(true)
				if e.typ != 0 {
					emit(e)
					emit(Segment{';', i, nil})
				}
				strip = true
				comment = true
				typ = 'C'
				m = i + 1
				continue
			}

			if r == '"' || r == '\'' {
				if quote == 0 {
					e := expr(false)
					if e.typ != 0 {
						emit(e)
					}
					m = i + 1
					typ = 'Q'
					quote = r
				} else {
					if quote == r {
						e := expr(false)
						if e.typ != 0 {
							emit(e)
						}
						typ = 'R'
						quote = 0
						m = i + 1
					} else {
						buf = append(buf, r)
					}
				}
				if eof {
					if quote != 0 {
						panic("unclosed quote")
					}
					emitFinalDedents(i)
					done = true
					if len(queue) > 0 {
						return popQueue()
					}
					return nil
				}
				continue
			}

			if r == '\n' || r == ';' || eof {
				e := expr(true)
				if e.typ != 0 {
					emit(e)
				}
				emit(Segment{';', i, nil})
				strip = true
				if r == '\n' {
					lineStart = true
					lineIndent = 0
					lineIndentHasSpaces = false
				}
			} else {
				if r == '\\' && !escape {
					escape = true
				} else {
					buf = append(buf, r)
				}
			}

			if eof {
				if quote != 0 {
					panic("unclosed quote")
				}
				emitFinalDedents(i)
				done = true
				if len(queue) > 0 {
					return popQueue()
				}
				return nil
			}
			if len(queue) > 0 {
				return popQueue()
			}
		}
	}

	return yield
}

func Build_ast_from_tokens(next TokenSource) *Token {
	root := &Token{Typ: 'P', Toks: make([]*Token, 0)}
	mainParams := &Token{Typ: 'T'}
	mainBlock := &Token{Typ: 'B', Toks: make([]*Token, 0)}
	main := &Token{Typ: 'F', Toks: []*Token{mainParams, mainBlock}}
	root.Toks = append(root.Toks, main)

	curr := mainBlock
	stack := make([]*Token, 0)
	var pcurr *Token
	pstack := make([]*Token, 0)
	var pendingBlock *Token
	var pendingParent *Token
	var elseCandidate *Token

	openPendingBlock := func() {
		stack = append(stack, curr)
		pstack = append(pstack, pcurr)
		curr = pendingBlock
		pcurr = pendingParent
		pendingBlock = nil
		pendingParent = nil
		elseCandidate = nil
	}

	closeBlock := func(tok *Token) {
		if len(stack) == 0 {
			panic(ParseError{
				Msg:   "unexpected dedent",
				Start: tok.Start,
				End:   tok.End,
			})
		}
		closingParent := pcurr
		inheritRangeFromChildren(curr)
		if closingParent != nil {
			closingParent.End = curr.End
		}
		curr = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		pcurr = pstack[len(pstack)-1]
		pstack = pstack[:len(pstack)-1]
		if closingParent != nil && closingParent.Typ == 'K' && len(closingParent.Buf) > 0 && closingParent.Buf[0] == xxisToken.IF && closingParent.Toks[2] == nil {
			elseCandidate = closingParent
		} else {
			elseCandidate = nil
		}
	}

	for {
		cmd := next()
		if cmd == nil {
			break
		}

		if cmd.Typ == tokenIndent {
			if pendingBlock == nil {
				panic(ParseError{
					Msg:   "unexpected indent",
					Start: cmd.Start,
					End:   cmd.End,
				})
			}
			openPendingBlock()
			continue
		}
		if cmd.Typ == tokenDedent {
			if pendingBlock != nil {
				panic(ParseError{
					Msg:   "expected indented block",
					Start: cmd.Start,
					End:   cmd.End,
				})
			}
			closeBlock(cmd)
			continue
		}
		if pendingBlock != nil {
			panic(ParseError{
				Msg:   "expected indented block",
				Start: cmd.Start,
				End:   cmd.End,
			})
		}

		if is_if_cmd(cmd) {
			elseCandidate = nil
			cond, body, mode, exprIndex := parse_if(cmd)
			inheritRangeFromChildren(cond)
			block := &Token{Typ: 'B', Toks: make([]*Token, 0)}
			inheritRangeFromChildren(block)
			ifBuf := []rune{xxisToken.IF, mode}
			if mode == xxisToken.IfCondExpr {
				ifBuf = append(ifBuf, rune(exprIndex))
			}
			if_tok := &Token{Typ: 'K', Buf: ifBuf, Toks: []*Token{cond, block, nil}}
			curr.Toks = append(curr.Toks, if_tok)
			if_tok.Start = cmd.Start
			if body != nil {
				block.Toks = append(block.Toks, body)
				block.Start = body.Start
				block.End = body.End
				if_tok.End = cmd.End
			} else {
				pendingBlock = block
				pendingParent = if_tok
			}
		} else if is_var_cmd(cmd) {
			elseCandidate = nil
			varTok := parse_var(cmd)
			curr.Toks = append(curr.Toks, varTok)
		} else if is_function_cmd(cmd) {
			elseCandidate = nil
			params := parse_function(cmd)
			block := &Token{Typ: 'B', Toks: make([]*Token, 0)}
			fn := &Token{Typ: 'F', Start: cmd.Start, Toks: []*Token{params, block}}
			root.Toks = append(root.Toks, fn)

			pendingBlock = block
			pendingParent = fn
		} else if is_single_word(cmd, "else") {
			if elseCandidate == nil {
				panic(ParseError{
					Msg:   "else outside if",
					Start: cmd.Start,
					End:   cmd.End,
				})
			}
			block := &Token{Typ: 'B', Toks: make([]*Token, 0)}
			elseCandidate.Toks[2] = block
			pendingBlock = block
			pendingParent = elseCandidate
			elseCandidate = nil
		} else {
			elseCandidate = nil
			cmd.Typ = 'C'
			curr.Toks = append(curr.Toks, split_and_or(cmd))
		}
	}

	if pendingBlock != nil {
		panic("expected indented block")
	}
	if len(stack) > 0 {
		panic("unclosed block")
	}

	inheritRangeFromChildren(mainBlock)
	main.Start = mainBlock.Start
	main.End = mainBlock.End
	inheritRangeFromChildren(root)

	return root
}
