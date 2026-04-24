package parser

import xxisToken "github.com/Greccl/xxis/internal/token"

type Token = xxisToken.Token

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

	meta := make([]Segment, 0)
	var pending *Token
	done := false

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

	emit := func(seg Segment) bool {
		switch seg.typ {
		case 'R', 'Q':
			meta = append(meta, seg)
		case ';':
			strip = true
			if len(meta) > 0 {
				pending = subcmd_by_segment(meta)
				meta = make([]Segment, 0)
				return true
			}
		}
		return false
	}

	yield := func() *Token {
      pending = nil

		if done { return nil }

		for {
			i, r, eof := read()

			if escape {
				escape = false
				buf = append(buf, r)
				if !eof { continue }
			}

			if strip {
				switch r {
				case ' ', '\t':
					if !eof { continue }
				}
				strip = false
			}

			if comment {
				if r == '\n' || eof {
					e := expr(false)
					strip = true
					if e.typ != 0 {
						if emit(e) {
							return pending
						}
					}
				} else {
					buf = append(buf, r)
				}
				if eof {
					if quote != 0 {
						panic("unclosed quote")
					}
					done = true
					if pending != nil {
						return pending
					}
					return nil
				}
				continue
			}

			if quote == 0 && r == '#' {
				e := expr(true)
				if e.typ != 0 {
					if emit(e) {
						return pending
					}
					if emit(Segment{';', i, nil}) {
						return pending
					}
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
						if emit(e) {
							return pending
						}
					}
					m = i + 1
					typ = 'Q'
					quote = r
				} else {
					if quote == r {
						e := expr(false)
						if e.typ != 0 {
							if emit(e) {
								return pending
							}
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
					done = true
					if pending != nil {
						return pending
					}
					return nil
				}
				continue
			}

			if r == '\n' || r == ';' || eof {
				e := expr(true)
				if e.typ != 0 {
					if emit(e) {
						return pending
					}
				}
				if emit(Segment{';', i, nil}) {
					return pending
				}
				strip = true
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
				done = true
				if pending != nil {
					return pending
				}
				return nil
			}
		}
	}

	return yield
}





func Build_ast_from_tokens(next TokenSource) *Token {
	root := &Token{Typ: 'F', Toks: make([]*Token, 0)}
	curr := root
	stack := make([]*Token, 0)
	var pcurr *Token
	pstack := make([]*Token, 0)

	for {
		cmd := next()
		if cmd == nil { break }

		if is_single_word(cmd, "end") {
			if len(stack) == 0 {
				panic("unexpected end")
			}
			curr = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			pcurr = pstack[len(pstack)-1]
			pstack = pstack[:len(pstack)-1]
		} else if is_if_cmd(cmd) {
			cond, body := parse_if(cmd)
			block := &Token{Typ: 'B', Toks: make([]*xxisToken.Token, 0)}
			if_tok := &Token{Typ: 'K', Buf: []rune{xxisToken.IF}, Toks: []*Token{cond, block, nil}}
			curr.Toks = append(curr.Toks, if_tok)
			if body != nil {
				block.Toks = append(block.Toks, body)
			} else {
				stack = append(stack, curr)
				pstack = append(pstack, pcurr)
				curr = block
				pcurr = if_tok
			}
		} else if is_single_word(cmd, "else") {
			if pcurr == nil {
				panic("else outside if / 1")
			}
			if pcurr.Buf[0] != xxisToken.IF {
				panic("else outside if / 2")
			}
			block := &Token{Typ: 'B', Toks: make([]*Token, 0)}
			pcurr.Toks[2] = block
			curr = block
		} else {
			cmd.Typ = 'C'
			curr.Toks = append(curr.Toks, split_and_or(cmd))
		}
	}

	if len(stack) > 0 {
		panic("unclosed block")
	}

	return root
}
