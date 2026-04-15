package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode"
	"unicode/utf8"
	// "path/filepath"
	// "context"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type IndexedRuneSource func() (int, rune, bool)

func enumerate_file(path string) IndexedRuneSource {
	file, err := os.Open(path)
	check(err)

	var i int = -1
	reader := bufio.NewReader(file)
	return func() (int, rune, bool) {
		r, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return i, 0, true
			}
			panic(fmt.Errorf("error reading rune: %w", err))
		}
		i++
		return i, r, false
	}
}

func enumerate_string(str string) IndexedRuneSource {
	var i int
	return func() (int, rune, bool) {
		if i >= len(str) {
			return i, 0, true
		}
		r, size := utf8.DecodeRuneInString(str[i:])
		idx := i
		i += size
		return idx, r, false
	}
}

func enumerate_runes(rns []rune) IndexedRuneSource {
	var i int
	return func() (int, rune, bool) {
		if i >= len(rns) {
			return i, 0, true
		}
		idx := i
		i++
		return idx, rns[idx], false
	}
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






type Segment struct {
	typ    rune
	offset int
	buf    []rune
}

/*
func get_segments(read IndexedRuneSource) []Segment {
	m := 0
	escape := false
	comment := false
	buf := make([]rune, 0)
	typ := 'R'
	var quote rune = 0
	strip := true
	segments := make([]Segment, 0)

	expr := func(final bool) Segment {
		e := buf
		if final {
			e = trim_buffer(e, false, true)
		}
		buf = make([]rune, 0)
		if len(e) > 0 {
			return Segment{typ, m, e}
		} else {
			return Segment{}
		}
	}

	segments = append(segments, Segment{'L', 0, nil})

	for {
		i, r, eof := read()

		if r == '\n' {
			segments = append(segments, Segment{'L', i + 1, nil})
		}

		if escape {
			escape = false
			buf = append(buf, r)
			continue
		}

		if strip {
			switch r {
			case ' ', '\t':
				continue
			}
			strip = false
		}

		if comment {
			if r == '\n' || eof {
				e := expr(false)
				strip = true
				if e.typ != 0 {
					segments = append(segments, e)
				}
			} else {
				buf = append(buf, r)
			}
			continue
		}

		if quote == 0 && r == '#' {
			e := expr(true)
			if e.typ != 0 {
				segments = append(segments, e)
				segments = append(segments, Segment{';', i, nil})
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
					segments = append(segments, e)
				}
				m = i + 1
				if r == '"' {
					typ = 'D'
				} else {
					typ = 'S'
				}
				quote = r
			} else {
				if quote == r {
					e := expr(false)
					if e.typ != 0 {
						segments = append(segments, e)
					}
					typ = 'R'
					quote = 0
					m = i + 1
				} else {
					buf = append(buf, r)
				}
			}
			continue
		}

		if r == '\n' || r == ';' || eof {
			e := expr(true)
			if e.typ != 0 {
				segments = append(segments, e)
			}
			segments = append(segments, Segment{';', i, nil})
			strip = true
		} else {
			if r == '\\' && !escape {
				escape = true
			} else {
				buf = append(buf, r)
			}
		}

		if eof {
			break
		}
	}

	if quote != 0 {
		// segments = append(segments, Segment{'E', -1, nil})
		panic("unclosed quote")
	}

	return segments
}
*/

type ParseContext0 struct {
	root  *Token
	curr  *Token
	stack []*Token

	text   []rune
	m      int
	escape bool
	state  int
	quote  rune
}

func (self *ParseContext0) init() {
	self.root = &Token{}
	self.root.typ = 'I'
	self.root.toks = make([]*Token, 0)
	self.curr = self.root
	self.stack = make([]*Token, 0)
}

func (self *ParseContext0) new_segment(text []rune) {
	self.text = text
	self.m = 0
	self.escape = false
}

func (self *ParseContext0) append_runes(typ rune, buf []rune) {
	self.curr.toks = append(self.curr.toks, &Token{typ, buf, nil})
}

func (self *ParseContext0) append_token(tok *Token) {
	self.curr.toks = append(self.curr.toks, tok)
}

func (self *ParseContext0) append_text(i int) {
	if i-self.m > 0 {
		self.append_runes('T', self.text[self.m:i])
	}
}

func (self *ParseContext0) start_subcmd() {
	sub := &Token{'S', nil, make([]*Token, 0)}
	self.append_token(sub)
	self.push(sub)
}

/*
func (self *ParseContext0) push_access(i):
      s = []
      self.sub.append(['A', self.text[self.m:i], s])
      self.subs.append(self.sub)
      self.sub = s
      self.subtype = 'A'
*/

func (self *ParseContext0) push(tok *Token) {
	self.stack = append(self.stack, self.curr)
	self.curr = tok
}

func (self *ParseContext0) pop() {
	self.curr = self.stack[len(self.stack)-1]
	self.stack = self.stack[0 : len(self.stack)-1]
}

func subcmd_by_text(ctx *ParseContext0) *Token {
	read := enumerate_runes(ctx.text)
	for {
		i, r, eof := read()

		if r == '\\' {
			ctx.escape = true
			continue
		}

		if ctx.escape {
			ctx.escape = false
			continue
		}

		if ctx.quote == 0 {
			if r == '\'' || r == '"' {
				ctx.quote = r
				continue
			}
		} else {
			if r == ctx.quote {
				ctx.quote = 0
			}
			continue
		}

		if ctx.state == 0 && r == '$' {
			ctx.state = 1
			continue
		}

		if ctx.state == 1 {
			if is_iden_start(r) {
				ctx.append_text(i - 1)
				ctx.m = i
				ctx.state = 2
			} else if r == '(' {
				ctx.append_text(i - 1)
				ctx.start_subcmd()
				ctx.m = i + 1
				ctx.state = 0
			} else {
				ctx.state = 0
			}
		} else if ctx.state == 2 {
			if !is_iden_char(r) {
				if r == '[' {
					// ctx.push_access(i)
					ctx.m = i + 1
				} else {
					ctx.append_runes('V', ctx.text[ctx.m:i])
					ctx.m = i
				}
				if r == '$' {
					ctx.state = 1
				} else {
					ctx.state = 0
				}
			}
		}

		if ctx.state == 0 && len(ctx.stack) > 0 {
			closed := false
			if ctx.curr.typ == 'S' && r == ')' {
				closed = true
			} else if ctx.curr.typ == 'A' && r == ']' {
				closed = true
			}
			if closed {
				ctx.append_text(i)
				if len(ctx.curr.toks) == 0 {
					ctx.append_runes('_', nil)
				}
				ctx.pop()
				ctx.m = i + 1
			}
		}

		if eof {
			ctx.append_text(i)
			break
		}
	}

	return ctx.root
}

/*
func get_metaexpressions(segments []Segment) [][]Segment {
	res := make([][]Segment, 0)
	acc := make([]Segment, 0)

	for _, seg := range segments {
		switch seg.typ {
		case ';':
			if len(acc) > 0 {
				res = append(res, acc)
				acc = make([]Segment, 0)
			}
		case 'R', 'S', 'D':
			acc = append(acc, seg)
		}
	}
	return res
}
*/

func subcmd_by_segment(segments []Segment) *Token {
	ctx := &ParseContext0{}
	ctx.init()
	for _, seg := range segments {
		if seg.typ == 'S' || seg.typ == 'D' {
			subctx := &ParseContext0{}
			subctx.init()
			subctx.new_segment(seg.buf)
			s := subcmd_by_text(subctx)
			s.typ = 'Q'
			ctx.append_token(s)
		} else {
			ctx.new_segment(seg.buf)
			subcmd_by_text(ctx)
		}
	}
	return ctx.root
}

/*
func build_ast(metas [][]Segment) *Token {
	cmds := make([]*Token, 0, len(metas))
	for _, meta := range metas {
		cmds = append(cmds, subcmd_by_segment(meta))
	}

	root := &Token{typ: 'F', toks: make([]*Token, 0)}
	curr := root
	stack := make([]*Token, 0)
	var pcurr *Token
	pstack := make([]*Token, 0)

	for _, cmd := range cmds {
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
         block := &Token{typ: 'B', toks: make([]*Token, 0)}
         if_tok := &Token{typ: 'K', buf: []rune{IF}, toks: []*Token{cond, block, nil}}
         curr.toks = append(curr.toks, if_tok)
			if body != nil {
            block.toks = append(block.toks, body)
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
         if pcurr.buf[0] != IF {
            panic("else outside if / 2")
         }
         block := &Token{typ: 'B', toks: make([]*Token, 0)}
         pcurr.toks[2] = block
         curr = block
		} else {
		   cmd.typ = 'C'
		   curr.toks = append(curr.toks, cmd)
		}
	}

	if len(stack) > 0 {
		panic("unclosed block")
	}

	return root
}
*/







type Function struct {
	code []Instr
}

type CompileFrame struct {
	f *Function
	// s []*Instruction
}

type Compiler struct {
	funcs map[string]*Function
	tokens []*Token

	f *Function
	// stack []*Instruction

	frames []CompileFrame
}

func New_Compiler() *Compiler {
	self := &Compiler{}
	self.funcs = make(map[string]*Function)
	self.tokens = make([]*Token, 0)
	self.f = &Function{}
	self.f.code = make([]Instr, 0)
	self.funcs["main"] = self.f
	return self
}

func (self *Compiler) push(op, val int) int {
	self.f.code = append(self.f.code, Instr{op, val})
	return len(self.f.code) - 1
}

func (self *Compiler) finish() {
	self.push(HALT, 0)
}

func (self *Compiler) splitAndOr(tok *Token) *Token {
	if tok.typ != 'C' {
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

	for _, t := range tok.toks {
		if t.typ != 'T' {
			if stripLeft {
				stripLeft = false
			}
			curr = append(curr, t)
			continue
		}

		buf := t.buf
		if stripLeft {
			buf = lstrip(buf)
			stripLeft = false
		}

		pos := 0
		for {
			opPos, op := findOp(buf, pos)
			if opPos == -1 {
				if pos == 0 {
					if len(buf) > 0 {
						curr = append(curr, &Token{typ: 'T', buf: buf})
					}
				} else if pos < len(buf) {
					frag := buf[pos:]
					if len(frag) > 0 {
						curr = append(curr, &Token{typ: 'T', buf: frag})
					}
				}
				break
			}

			found = true
			left := rstrip(buf[pos:opPos])
			if len(left) > 0 {
				curr = append(curr, &Token{typ: 'T', buf: left})
			}

			if len(curr) > 0 {
				parts = append(parts, &Token{typ: 'C', toks: curr})
				ops = append(ops, op)
			}
			curr = make([]*Token, 0)

			pos = opPos + 2
			for pos < len(buf) && is_space(buf[pos]) {
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
		parts = append(parts, &Token{typ: 'C', toks: curr})
	}
	if len(parts) == 0 || len(ops) == 0 || len(parts) != len(ops)+1 {
		return tok
	}

	block := &Token{typ: 'B', toks: make([]*Token, len(parts))}
	for i, part := range parts {
		if i > 0 { part.typ = ops[i-1] }
		block.toks[i] = part
	}
	return block
}

func (self *Compiler) process(tok *Token) {
	switch tok.typ {
	   case 'C':
	      split := self.splitAndOr(tok)
	      if split.typ == 'B' {
	      	for _, t := range split.toks {
   		      self.process(t)
	      	}
	      } else {
	         // fmt.Printf("try to teplace %s\n", split.repr())
   		   self.replace(split, 0, true)
	      }
		case 'K':
		   switch tok.buf[0] {
		      case IF:
		         self.process(tok.toks[0])
		         i := self.push(JMPN, 0)
		         for _, t := range tok.toks[1].toks {
		            self.process(t)
		         }
   		      self.f.code[i].arg = len(self.f.code)
		         if tok.toks[2] != nil {
   		         self.f.code[i].arg++
   		         i = self.push(JMP, 0)
   		         for _, t := range tok.toks[2].toks {
   		            self.process(t)
   		         }
   		         self.f.code[i].arg = len(self.f.code)
   		      }
		   }
      case '&':
        self.push(JMPN, len(self.f.code)+2)
        tok.typ = 'C'
        self.replace(tok, 0, true)
      case '|':
        self.push(JMPZ, len(self.f.code)+2)
        tok.typ = 'C'
        self.replace(tok, 0, true)
   }
}

func (self *Compiler) replace(tok *Token, reserved int, add bool) {
   // fmt.Printf("replace %c %d %s\n", tok.typ, reserved, tok.repr())
	for i := 0; i < len(tok.toks); i++ {
		t := tok.toks[i]
		switch t.typ {
		case 'S':
			t.buf = []rune{rune(reserved)}
			replacement := &Token{typ: 'R'}
			replacement.buf = []rune{rune(reserved)}
			self.replace(t, reserved, true)
			reserved++
			tok.toks[i] = replacement
		case 'Q':
			self.replace(t, reserved, false)
			// reserved++
		}
	}
	if add {
      self.push(CMD, len(self.tokens))
      self.tokens = append(self.tokens, tok)
	}
}
