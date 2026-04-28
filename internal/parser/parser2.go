package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type IndexedRuneSource func() (int, rune, bool)
type LineInfoResolver func(int) int

func Enumerate_file(path string) (IndexedRuneSource, LineInfoResolver) {
	file, err := os.Open(path)
	check(err)

   var lines = []int{0}
	var i int = -1
	reader := bufio.NewReader(file)

	read := func() (int, rune, bool) {
		r, _, err := reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return i, 0, true
			}
			panic(fmt.Errorf("error reading rune: %w", err))
		}
		i++
		if r == '\n' {
		   lines = append(lines, i)
		}
		return i, r, false
	}

	resolver := func(offset int) int {
	   for i, start := range lines {
	      if offset < start {
	         return i
	      }
	   }
	   return 1
	}

	return read, resolver
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
	base   int
	escape bool
	state  int
	quote  rune
	strip  bool
}

func (self *ParseContext0) init() {
	self.root = &Token{}
	self.root.Typ = 'I'
	self.root.Toks = make([]*Token, 0)
	self.curr = self.root
	self.stack = make([]*Token, 0)
}

func (self *ParseContext0) new_segment(seg Segment) {
	self.text = seg.buf
	self.m = 0
	self.base = seg.offset
	self.escape = false
}

func (self *ParseContext0) append_runes(typ rune, buf []rune, s, e int) {
	self.curr.Toks = append(self.curr.Toks, &Token{Typ:typ, Start:s, End:e, Buf:buf, Toks:nil})
}

func (self *ParseContext0) append_token(tok *Token) {
	self.curr.Toks = append(self.curr.Toks, tok)
}

func (self *ParseContext0) append_text(i int) {
	if i-self.m > 0 {
		self.append_runes('T', self.text[self.m:i], self.base+self.m, self.base+i)
	}
}

func (self *ParseContext0) start_subcmd(t rune, s int) {
	sub := &Token{Typ:'S', Buf:[]rune{t, -1}, Toks:make([]*Token, 0), Start:s}
	self.append_token(sub)
	self.push(sub)
	self.strip = true
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

      if ctx.strip {
      	switch r {
      	case ' ', '\t':
      		if !eof { continue }
      	}
      	ctx.strip = false
      	ctx.m = i
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
				ctx.start_subcmd('S', ctx.base+i-1)
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
					ctx.append_runes('V', ctx.text[ctx.m:i], ctx.base+ctx.m-1, ctx.base+i)
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
				ctx.start_subcmd('X', ctx.base+i-1)
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
					ctx.append_runes('_', nil, 0, 0)
				} else {
				   ctx.curr.End = ctx.base + i + 1
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
			subctx.new_segment(seg)
			s := subcmd_by_text(subctx)
			s.Typ = 'Q'
			s.Start = seg.offset - 1
			s.End = seg.offset + len(seg.buf) + 1
			ctx.append_token(s)
		} else {
			ctx.new_segment(seg)
			subcmd_by_text(ctx)
		}
	}
	ctx.root.Start = segments[0].offset
	last := segments[len(segments) - 1]
	ctx.root.End = last.offset + len(last.buf)
	// ctx.root.Dump()
	return ctx.root
}