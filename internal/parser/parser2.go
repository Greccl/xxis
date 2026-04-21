package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	// "unicode"
	"unicode/utf8"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type IndexedRuneSource func() (int, rune, bool)

func Enumerate_file(path string) IndexedRuneSource {
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

func Enumerate_string(str string) IndexedRuneSource {
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

func Enumerate_runes(rns []rune) IndexedRuneSource {
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
	self.root.Typ = 'I'
	self.root.Toks = make([]*Token, 0)
	self.curr = self.root
	self.stack = make([]*Token, 0)
}

func (self *ParseContext0) new_segment(text []rune) {
	self.text = text
	self.m = 0
	self.escape = false
}

func (self *ParseContext0) append_runes(typ rune, buf []rune) {
	self.curr.Toks = append(self.curr.Toks, &Token{typ, buf, nil})
}

func (self *ParseContext0) append_token(tok *Token) {
	self.curr.Toks = append(self.curr.Toks, tok)
}

func (self *ParseContext0) append_text(i int) {
	if i-self.m > 0 {
		self.append_runes('T', self.text[self.m:i])
	}
}

func (self *ParseContext0) start_subcmd(t rune) {
	sub := &Token{'S', []rune{t, -1}, make([]*Token, 0)}
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
	read := Enumerate_runes(ctx.text)
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

		if ctx.state == 0 && r == '!' {
			ctx.state = 3
			continue
		}

		if ctx.state == 1 {
			if is_iden_start(r) {
				ctx.append_text(i - 1)
				ctx.m = i
				ctx.state = 2
			} else if r == '(' {
				ctx.append_text(i - 1)
				ctx.start_subcmd('S')
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
		} else if ctx.state == 3 {
			if r == '(' {
				ctx.append_text(i - 1)
				ctx.start_subcmd('X')
				ctx.m = i + 1
				ctx.state = 0
			} else {
				ctx.state = 0
			}
		}

		if ctx.state == 0 && len(ctx.stack) > 0 {
			closed := false
			t := ctx.curr.Typ
			if (t == 'S' || t == 'X') && r == ')' {
				closed = true
			} else if t == 'A' && r == ']' {
				closed = true
			}
			if closed {
				ctx.append_text(i)
				if len(ctx.curr.Toks) == 0 {
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

func subcmd_by_segment(segments []Segment) *Token {
	ctx := &ParseContext0{}
	ctx.init()
	for _, seg := range segments {
		if seg.typ == 'Q' {
			subctx := &ParseContext0{}
			subctx.init()
			subctx.new_segment(seg.buf)
			s := subcmd_by_text(subctx)
			s.Typ = 'Q'
			ctx.append_token(s)
		} else {
			ctx.new_segment(seg.buf)
			subcmd_by_text(ctx)
		}
	}
	return ctx.root
}