package compiler

import xxisVm "github.com/Greccl/xxis/internal/vm"
import xxisToken "github.com/Greccl/xxis/internal/token"
// import xxisParser "github.com/Greccl/xxis/internal/parser"

type Function = xxisVm.Function
type Program = xxisVm.Program
type Instr = xxisVm.Instr
type Token = xxisToken.Token





type CompileFrame struct {
	f *Function
	// s []*Instruction
}

type Compiler struct {
	Funcs map[string]*Function
	Tokens []*Token

	f *Function
	// stack []*Instruction

	frames []CompileFrame
}

func New_Compiler() *Compiler {
	self := &Compiler{}
	self.Funcs = make(map[string]*Function)
	self.Tokens = make([]*Token, 0)
	self.f = &Function{}
	self.f.Code = make([]Instr, 0)
	self.Funcs["main"] = self.f
	return self
}

func (self *Compiler) push(op, val int) int {
	self.f.Code = append(self.f.Code, Instr{op, val})
	return len(self.f.Code) - 1
}

func (self *Compiler) Finish() {
	self.push(xxisVm.HALT, 0)
}

/*
func (self *Compiler) splitAndOr(tok *Token) *Token {
	if tok.Typ != 'C' {
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
			buf = xxisParser.Trim_buffer(buf, true, false)
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
			left := xxisParser.Trim_buffer(buf[pos:opPos], false, true)
			if len(left) > 0 {
				curr = append(curr, &Token{Typ: 'T', Buf: left})
			}

			if len(curr) > 0 {
				parts = append(parts, &Token{Typ: 'C', Toks: curr})
				ops = append(ops, op)
			}
			curr = make([]*Token, 0)

			pos = opPos + 2
			for pos < len(buf) && xxisParser.Is_space(buf[pos]) {
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
		if i > 0 { part.Typ = ops[i-1] }
		block.Toks[i] = part
	}
	return block
}
*/

func (self *Compiler) Process(tok *Token) {
	switch tok.Typ {
		case 'B':
			for _, t := range tok.Toks {
				self.Process(t)
			}
	   case 'C':
         /*
	      split := self.splitAndOr(tok)
	      if split.Typ == 'B' {
	      	for _, t := range split.Toks {
   		      self.Process(t)
	      	}
	      } else {
	         // fmt.Printf("try to teplace %s\n", split.repr())
   		   self.replace(split, 0, true)
	      }
	      */
			self.replace(tok, 0, true)
		case 'K':
		   switch tok.Buf[0] {
		      case xxisToken.IF:
		         self.Process(tok.Toks[0])
		         i := self.push(xxisVm.JMPN, 0)
		         for _, t := range tok.Toks[1].Toks {
		            self.Process(t)
		         }
   		      self.f.Code[i].Arg = len(self.f.Code)
		         if tok.Toks[2] != nil {
   		         self.f.Code[i].Arg++
   		         i = self.push(xxisVm.JMP, 0)
   		         for _, t := range tok.Toks[2].Toks {
   		            self.Process(t)
   		         }
   		         self.f.Code[i].Arg = len(self.f.Code)
   		      }
		   }
      case '&':
        self.push(xxisVm.JMPN, len(self.f.Code)+2)
        tok.Typ = 'C'
        self.replace(tok, 0, true)
      case '|':
        self.push(xxisVm.JMPZ, len(self.f.Code)+2)
        tok.Typ = 'C'
        self.replace(tok, 0, true)
   }
}

func (self *Compiler) replace(tok *Token, reserved int, add bool) {
   // fmt.Printf("replace %c %d %s\n", tok.typ, reserved, tok.repr())
	for i := 0; i < len(tok.Toks); i++ {
		t := tok.Toks[i]
		switch t.Typ {
		case 'S':
			// t.buf = []rune{rune(reserved)}
			t.Buf[1] = rune(reserved)
			replacement := &Token{Typ: 'R'}
			replacement.Buf = []rune{rune(reserved)}
			self.replace(t, reserved, true)
			reserved++
			tok.Toks[i] = replacement
		case 'Q':
			self.replace(t, reserved, false)
		}
	}
	if add {
      self.push(xxisVm.CMD, len(self.Tokens))
      self.Tokens = append(self.Tokens, tok)
	}
}
